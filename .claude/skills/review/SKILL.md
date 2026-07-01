---
name: review
description: Code Review skill for reviewing implemented changes before commit and PR. Use this skill after documentation is complete to perform a thorough code review, validate implementation quality, and prepare for commit. Creates footprints and updates state for traceability.
---

# Review Skill

This skill performs a comprehensive code review of implemented changes. It should be used **after documentation** is complete and **before commit and PR creation**.

## Workflow Position

```
... → /implement → /document → /review → /commit → /pull_request
```

## Variables

- `branch_name`: Current git branch (auto-detected)
- `descriptive_name`: Derived from state.json or branch name

## Instructions

### 1. Load Context

First, read the current state to understand what was implemented:

```
agentic/{branch_name}/state.json
```

If state.json exists, extract:
- `plan_file_path` - The original plan
- `latest_footprint` - Most recent footprint
- `descriptive_name` - Feature/bug/patch name

If state.json doesn't exist, derive context from:
- Git branch name
- Recent git commits
- Changed files in working directory

### 2. Gather Changes for Review

Collect all changes that need review:

1. **Git diff analysis**: Run `git diff --name-only HEAD~{N}` to find changed files (where N is commits since branch creation or last review)
2. **Changed file review**: Read each modified file to understand the changes
3. **Plan comparison**: Compare implementation against the original plan (if available)

### 3. Review Checklist

Perform review against these criteria:

#### Code Quality
- [ ] Code follows project style guidelines and conventions
- [ ] No unnecessary complexity or over-engineering
- [ ] Functions/methods have single responsibility
- [ ] Variable and function names are clear and descriptive
- [ ] No dead code or commented-out code
- [ ] No console.log/print statements left in (unless intentional)

#### Functionality
- [ ] Implementation matches the plan/requirements
- [ ] All acceptance criteria are met
- [ ] Edge cases are handled appropriately
- [ ] Error handling is adequate

#### Security
- [ ] No hardcoded credentials or secrets
- [ ] Input validation where needed
- [ ] No SQL injection, XSS, or other OWASP vulnerabilities
- [ ] Proper authentication/authorization checks

#### Testing
- [ ] Unit tests exist for new functionality
- [ ] Tests cover edge cases
- [ ] All tests pass

#### Documentation
- [ ] Code comments where logic is complex
- [ ] API documentation if applicable
- [ ] README updates if needed

#### Type Safety Audit

**@ts-nocheck Files Check**:
```bash
# Before review
git diff master --name-only | xargs grep -l "// @ts-nocheck" || echo "No @ts-nocheck files"

# Flag if count increased
```

**Questions to Ask**:
- Are new `@ts-nocheck` files added?
- Can type errors be fixed instead of suppressed?
- Is this temporary or permanent?
- Is there a tracking issue to remove it?

**Unsafe Type Assertions**:
```typescript
// ❌ WRONG - assumes structure without validation
const lead = data as Lead

// ✅ CORRECT - validates before assertion
if (!data.id || !data.name) throw new Error('Invalid lead data')
const lead = data as Lead

// ✅ BETTER - use Zod validation
const lead = LeadSchema.parse(data)
```

**Empty String Fallbacks**:
```typescript
// ❌ WRONG - masks missing config
const url = process.env.API_URL || ""

// ✅ CORRECT - fail fast
const url = process.env.API_URL!
if (!url) throw new Error('API_URL required')
```

**Evidence from footprint analysis:**
- 18 @ts-nocheck files found during migration review
- 12 unsafe type assertions identified
- 6 empty string URL fallbacks masked config issues

#### Elysia-Specific Review Checklist

**Route Registration Order**:
- [ ] Public routes registered BEFORE `authGuard()`
- [ ] Specific routes BEFORE generic `:id` patterns
- [ ] Guards placed after routes they should protect

**TypeBox Schema Validation**:
- [ ] Request body schema defined
- [ ] Response schema defined
- [ ] Required fields marked with `t.Required()`
- [ ] Optional fields use `t.Optional()`

**Auth & Security**:
- [ ] Protected routes validate `companyId` matches `ctx.user.companyId`
- [ ] No hardcoded credentials or tokens
- [ ] Error messages don't leak sensitive data

**Eden Treaty Client**:
- [ ] DELETE requests include empty body `{}`
- [ ] No empty string URL fallbacks
- [ ] Error handling present
- [ ] **Eden Treaty errors**: Verify no raw `throw error` from Eden Treaty calls
- [ ] Check `getErrorMessage()` usage in all API hooks
- [ ] Verify error messages are user-readable strings, not `[object Object]`

**Evidence from footprint analysis:**
- Route ordering issues found in 8 workflows
- Missing companyId validation found in 5 workflows
- Eden Treaty DELETE empty body issue in 7 workflows

#### Security Review Checklist

**Company Scoping** (Critical):
```typescript
// ❌ WRONG - any user can access any company's leads
app.get('/api/crm/leads/:id', async ({ params }) => {
  return prisma.lead.findUnique({ where: { id: params.id } })
})

// ✅ CORRECT - validates company ownership
app.get('/api/crm/leads/:id', async ({ params, user }) => {
  const lead = await prisma.lead.findUnique({ where: { id: params.id } })
  if (lead.companyId !== user.companyId) {
    throw new AppError('Not found', { status: 404 })
  }
  return lead
})
```

**Checklist**:
- [ ] All company-scoped queries filter by `companyId`
- [ ] User can only access their company's data
- [ ] 404 returned (not 403) to prevent enumeration
- [ ] No direct ID access without ownership check

**Evidence from footprint analysis:**
- 5 missing companyId validations found during review

#### Mobile/Responsive Checks (When mobile changes involved)

When reviewing code that includes mobile or responsive design changes:

- [ ] Touch targets are minimum 44px (all interactive elements: buttons, links, form controls)
- [ ] Sheet/Drawer headers visible with proper padding (`pr-12` to avoid X button overlap)
- [ ] Fixed elements (X buttons, headers, floating buttons) don't overlap content
- [ ] Desktop experience unchanged (verify at 1024px+)
- [ ] Tablet breakpoint tested (768px-1024px)
- [ ] Mobile breakpoint tested (< 768px)
- [ ] aria-labels present for mobile navigation buttons (accessibility)
- [ ] No unused imports (run lint to verify)

**Evidence from footprint analysis:**
- foot-review-mobile.md: Quality assessment included touch targets, accessibility
- foot-impl-complete-sidebar-fix.md: Fix was needed because initial review didn't catch X button overlap

#### AI Tool Review Checklist (When AI tool changes involved)

When reviewing code that includes AI tool changes:

- [ ] **Argument normalization applied**
  - Tool accepts string (direct ID)
  - Tool accepts object with camelCase key
  - Tool accepts object with snake_case key
- [ ] **Response format matches similar tools**
  - Success/error structure consistent
  - Return types match tool description
- [ ] **Error handling consistent**
  - Errors include meaningful messages
  - Errors don't expose internal details
- [ ] **Tool description updated**
  - Description reflects current behavior
  - Accepted argument formats documented
- [ ] **System prompt reflects new capabilities**
  - New tools mentioned in system prompt
  - Tool usage examples accurate

**Evidence from footprint analysis:**
- bug/not-always-create-a-comment: Tool consistency was root cause

#### Gateway Schema Review (When GraphQL changes involved)

When reviewing code that includes GraphQL schema changes:

- [ ] **Generated types match gateway schema**
  - Run `pnpm codegen` succeeds without errors
  - Generated types used correctly in code
- [ ] **Field prefixes handled**
  - Core fields use `core_` prefix in gateway
  - Auth fields use `auth_` prefix in gateway
  - Frontend queries use correct prefixed names
- [ ] **Codegen runs without errors**
  - No schema mismatches
  - No missing field errors
- [ ] **Frontend queries use correct field names**
  - Query field names match gateway schema
  - No hardcoded unprefixed names

**Evidence from footprint analysis:**
- feature-add-context-attachments: Gateway schema issues were common blocker

#### Fleet Test Label Verification

Before completing review, verify that test routing labels have been applied to the bead. Without labels, the fleet orchestrator won't route the feature for testing after PR creation.

**Check labels**:
```bash
# Get bead_id from state
bead_id=$(jq -r '.bead_id' agentic/$(git branch --show-current | sed 's|/|-|g')/state.json)

# Check if labels exist
bd show $bead_id  # Look for test:* labels
```

**If no test labels**:
- Run `/test-labels` or apply manually based on changed files
- `viz/backend/**`, `agents/**`, `mcp-servers*/**` → `test:backend`
- `viz/frontend/**` → `test:ui`
- Docs-only changes → `test:skip`

**Note**: Missing test labels is a minor review finding, not a blocker. Document it and apply them before commit.

See `/candidate-dev` for the full dev-candidate pipeline and `/test-labels` for auto-labeling rules.

---

#### Pre-existing Failure Handling

When tests fail during review:

1. **Check if failure existed before your changes**
   - Compare with master branch
   - Check if test was already failing

2. **If pre-existing**:
   - Document in footprint as "out-of-scope"
   - Do NOT fix in this workflow
   - Create separate bug report if needed
   - Mark review as APPROVED with note

3. **If caused by your changes**:
   - Fix immediately
   - Re-run full test suite
   - Update implementation footprint

**Evidence from footprint analysis:**
- feature-task-page-scroll-zones: Pre-existing test failure documented as out-of-scope

### Comms Escalation for Critical Findings

The reviewer self-decides APPROVED vs CHANGES_REQUESTED based on the checklist.

**Self-fix flow**: If CHANGES_REQUESTED, the SDLC orchestrator loops back to fix issues and re-reviews. No user involvement needed.

**Escalate via comms only when**:
- Critical security vulnerability found (hardcoded secrets, SQL injection, auth bypass)
- Destructive data operation without confirmation (DELETE without WHERE, DROP TABLE)

If comms_enabled and critical security finding: **ask the user directly in Cursor chat**:
"Review [{branch_name}]: Critical security finding: {description}. Proceed? Options: Fix and re-review, Approve anyway, Abort"
Wait for reply. On timeout or unclear: default to "Fix and re-review".

If `comms_enabled` is false or comms is unavailable, skip silently - default to CHANGES_REQUESTED for critical security issues.

### 4. Run Validation Commands

If a plan file exists with Validation Commands section, execute them:

```bash
# Example validations
bun test
bun lint
bun type-check
```

### 5. Document Review Findings

Create review summary with:
- **Files Reviewed**: List of all files examined
- **Issues Found**: Any problems discovered (categorized by severity)
- **Suggestions**: Optional improvements (not blockers)
- **Validation Results**: Output from validation commands
- **Review Decision**: APPROVED / CHANGES_REQUESTED

### 6. Create Footprint

Create footprint at: `agentic/{branch_name}/footprints/foot-review.md`

**Footprint Format:**

```markdown
# Review Footprint

**Timestamp**: {ISO 8601 timestamp}
**Branch**: {branch_name}
**Reviewer**: Claude Agent

## Review Summary

**Plan File**: {plan_file_path or "N/A"}
**Files Reviewed**: {count}
**Review Decision**: {APPROVED | CHANGES_REQUESTED}

## Files Reviewed

{list of files with brief notes}

## Review Checklist Results

### Code Quality
- {checklist results}

### Functionality
- {checklist results}

### Security
- {checklist results}

### Testing
- {checklist results}

### Documentation
- {checklist results}

## Validation Results

```
{output from validation commands}
```

## Issues Found

{list of issues by severity: Critical, Major, Minor}

## Suggestions (Non-blocking)

{optional improvements}

## Next Steps

{APPROVED: proceed with /commit and /pull_request}
{CHANGES_REQUESTED: list specific changes needed}
```

### 7. Update State

Update `agentic/{branch_name}/state.json`:

```json
{
  "latest_footprint": "agentic/{branch_name}/footprints/foot-review.md",
  "latest_command": "review",
  "review_timestamp": "{ISO 8601}",
  "review_decision": "{APPROVED | CHANGES_REQUESTED}",
  "files_reviewed": {count},
  "issues_found": {count},
  "next_command": "/commit",
  "next_command_metadata": {
    "command": "commit",
    "ready_for_commit": {true | false},
    "blocking_issues": {count}
  }
}
```

Merge these fields with existing state (preserve previous fields like `plan_file_path`, `descriptive_name`, etc.).

### 8. Beads Review Tracking

If `bead_id` exists in `state.json`, update the bead after review completes:

```bash
bd update $bead_id --notes="Review: $decision. Issues: $count"
```

Read `bead_id` from `agentic/{branch-name}/state.json`. If `bead_id` is null or missing, skip bead updates silently.

## Required Actions

You MUST complete these actions:

1. **Read state.json** to load implementation context
2. **Analyze all changed files** using git diff and file reads
3. **Execute validation commands** from plan or standard validations
4. **Create footprint** at `agentic/{branch_name}/footprints/foot-review.md`
5. **Update state.json** with review results
6. **Report review decision** with clear next steps

## Report

After completing the review, output a JSON report:

```json
{
  "review_decision": "APPROVED | CHANGES_REQUESTED",
  "files_reviewed": {count},
  "issues_found": {
    "critical": {count},
    "major": {count},
    "minor": {count}
  },
  "validation_passed": {true | false},
  "footprint_path": "agentic/{branch_name}/footprints/foot-review.md",
  "state_path": "agentic/{branch_name}/state.json",
  "next_command": "/commit | /implement (if changes needed)",
  "summary": "{brief summary of review findings}"
}
```

## Handling Review Outcomes

### If APPROVED
- Proceed to `/commit` to create a git commit
- Then use `/pull_request` to create a PR

### If CHANGES_REQUESTED
- List specific changes needed in the footprint
- Set `next_command` to `/implement` in state
- User should address issues and re-run `/review`

## Quick Review Mode

For small patches or minor changes, use abbreviated review:
- Focus on changed lines only
- Skip full checklist if changes are trivial
- Still create footprint and update state
- Mark as "Quick Review" in footprint
