---
name: implement
description: Implementation skill for executing plans created by planning skills (feature, bug, patch, chore, prototype). Use this skill when you have a plan ready in specs/ directory and need to implement it step by step. Creates footprints at each milestone, commits after each footprint, and tracks implementation progress in state.json. Handles all phases, validation, and reporting.
---

# Plan Implementation

Execute an implementation plan by following each step, creating footprints to document progress, and committing after each milestone.

## Variables

branch_name: Current git branch name (e.g., "feature/new-UI")
plan_file_path: $1 (optional) - Path to the plan file in specs/ directory

## Instructions

- IMPORTANT: You're implementing a plan that was created by a planning skill (feature, bug, patch, chore, or prototype).
- IMPORTANT: Read the plan file to understand what needs to be implemented.
- Execute each step in the plan's "Step by Step Tasks" section in order.
- Create footprints at key milestones to document implementation progress.
- Commit after creating each footprint with a descriptive commit message.
- Update the state file to track implementation progress.

### Implementation Workflow

1. **Read State File**: Check `agentic/{branch-name}/state.json` to find the plan file path
2. **Read Plan File**: Load and understand the plan from `specs/` directory
3. **Execute Steps**: Follow the "Step by Step Tasks" section step by step
4. **Create Footprints**: At each significant milestone, create a footprint
5. **Commit Changes**: After each footprint, commit the changes
6. **Run Validation**: Execute the "Validation Commands" from the plan
7. **Update State**: Update state file with implementation status

### Footprint Strategy

Create footprints at these key points:
- **foot-impl-start.md**: When starting implementation
- **foot-impl-phase-{N}.md**: After completing each phase (Phase 1, Phase 2, etc.)
- **foot-impl-complete.md**: When all steps are done and validated

Each footprint should document:
- What was accomplished
- Files created/modified
- Decisions made during implementation
- Any issues encountered and how they were resolved
- Next steps

### Commit Strategy

After each footprint:
1. Stage all relevant changes (avoid staging unrelated files)
2. Create a commit with format: `{type}({scope}): {description}`
3. Include footprint file in the commit
4. Reference the plan file in commit message if helpful

**Project Structure (agentic-ai-landing-zone):** See `.cursor/rules/project-structure.mdc`
- Backend: `viz/backend/`, `agents/`, `mcp-servers-tasks/`
- Frontend: `viz/frontend/`
- Tests: pytest (root), vitest in viz/frontend

**Database migrations:** If the plan includes schema changes: use project's migration tool (e.g. Alembic for SQLAlchemy). No Prisma in this project.

**UI Implementation:** When the plan includes frontend/UI changes:
- Add user-facing strings to translation/i18n files if the project uses them
- Never hardcode text strings in components when i18n exists

### In-Loop Review

When the user wants to manually review work during implementation (e.g. UI changes), suggest running `/in_loop_review {branch}`. This command checks out the branch, starts backend and frontend, and opens http://localhost:5173 for visual inspection.

### Beads Progress Tracking

If `bead_id` exists in `state.json`, update the bead at key implementation milestones:

- **At start**: `bd update $bead_id --status=in_progress`
- **After each phase**: `bd update $bead_id --notes="Phase $N complete: $phase_summary"`
- **At completion**: `bd update $bead_id --notes="Implementation complete. Validation: $status"`

Read `bead_id` from `agentic/{branch-name}/state.json`. If `bead_id` is null or missing, skip bead updates silently.

### Documentation Placement Rules

**NEVER create** standalone `IMPLEMENTATION_SUMMARY.md`, `*_IMPLEMENTATION.md`, or `MIGRATION_SUMMARY.md` files. These are ephemeral build logs that duplicate footprint content.

| What to create | Where | Example |
|----------------|-------|---------|
| Implementation progress log | Footprints in `agentic/` | `foot-impl-phase-1.md` |
| Module API/usage docs | Co-located `README.md` near code | `src/shared/events/README.md` |
| Results overview | `docs/results/` (via `/document-results` skill) | `docs/results/feature-x-results.md` |

**Module README.md files ARE encouraged** — they are living API docs that help both humans and LLMs discover how to use a module. Place them next to the code they document.

### Pre-Commit Verification (MANDATORY)

Before creating any commit, run lint/typecheck on affected code. **Do not skip this step.**

```bash
# Backend (viz/backend, agents)
pytest                    # Run tests
ruff check .              # Lint (if installed)

# Frontend (viz/frontend)
cd viz/frontend && npm run lint
cd viz/frontend && npm run test
cd viz/frontend && npx tsc --noEmit   # typecheck
```

**BLOCKING**: If lint/typecheck fails, fix ALL errors before committing. Do not commit with lint or type errors.

Common issues:
- Remove unused imports
- Fix TypeScript type errors
- Use underscore prefix (`_variable`) for intentionally unused parameters

### Project-Specific Patterns

Add project-specific implementation patterns in a `.claude/rules/` or skill extension. Examples:
- API route ordering (specific before generic)
- Auth/access control patterns
- UI component conventions (e.g. shadcn Sheet, form patterns)

### Numeric Value Safety

- [ ] Use `!= null` instead of falsy checks when zero is a valid value
- [ ] Especially for: prices, quantities, indices, ratings, percentages
- WRONG: `if (unitPrice) { ... }` -- skips valid zero
- RIGHT: `if (unitPrice != null) { ... }`
- Add terminal `?? 0` fallback for chained optional access: `a ?? b ?? 0`

**Evidence from footprint analysis:**
- 2 bug fixes involved falsy checks on numeric values where zero is valid (prices, quantities)

### Entity Type Normalization (Lookup Tables)

When adding entity types (lead sources, task statuses, etc.), use lookup table pattern:

**Schema**:
```prisma
model LeadSource {
  id          String   @id @default(cuid())
  name        String   @unique
  description String?
  leads       Lead[]
}

model Lead {
  id          String      @id @default(cuid())
  sourceId    String
  Source      LeadSource  @relation(fields: [sourceId], references: [id])
}
```

**Benefits**:
- Type safety through foreign keys
- Prevents invalid enum values
- UI can fetch available types dynamically
- Easy to add new types without schema migration
- Better data integrity vs string enums

### Prisma Relation Normalization

Prisma 7 returns relations with PascalCase names. Route handlers MUST normalize:

```typescript
function normalizeEntity(entity: PrismaEntity) {
  return {
    ...entity,
    relatedItems: entity.RelatedItems?.map(normalizeChild),
    RelatedItems: undefined, // Remove PascalCase key
  };
}
```

**Checklist**:
- [ ] All Prisma `include` relations normalized to camelCase in route response
- [ ] PascalCase keys set to `undefined` to prevent leaking to frontend
- [ ] Query schema accepts parameters frontend sends (e.g., `includeHistory: t.Optional(t.Boolean())`)

**Evidence from footprint analysis:**
- 3 bug fixes were caused by PascalCase relation names not normalized in route handlers
- `LeadHistory` relation appeared as PascalCase in Prisma responses but frontend expected camelCase

### Auth Token Refresh Mutex

When implementing token refresh logic, prevent race conditions:

```typescript
let refreshPromise: Promise<void> | null = null

async function refreshTokenWithMutex() {
  if (refreshPromise) {
    await refreshPromise
    return
  }

  refreshPromise = (async () => {
    try {
      const { data } = await api.auth.refresh.post({
        refreshToken: getRefreshToken()
      })
      setAccessToken(data.accessToken)
    } finally {
      refreshPromise = null
    }
  })()

  await refreshPromise
}
```

**Prevents**: Multiple simultaneous refresh calls when multiple API requests fail with 401

**Evidence from footprint analysis:**
- Auth refresh race condition discovered in auth module migration

### Gateway Schema Alignment (LEGACY)

**Note**: This is a legacy GraphQL pattern. For new connect-rpc services, use `bun run proto:generate` instead.

For legacy GraphQL changes in deprecated services:

1. **Update gateway schemas**:
   ```bash
   ./scripts/update-gateway-schemas.sh
   ```

2. **Restart gateway** to pick up changes:
   ```bash
   docker-compose restart gateway
   ```

3. **Run codegen**:
   ```bash
   bun codegen
   ```

**Evidence from footprint analysis:**
- feature-add-context-attachments: Gateway schema mismatch was common blocker

### AI Tool Implementation Pattern

When implementing AI tools, use this argument normalization template:

```typescript
/**
 * Normalize tool arguments to handle flexible input formats:
 * - String (direct ID): "attachment-id"
 * - Object with camelCase: { attachmentId: "..." }
 * - Object with snake_case: { attachment_id: "..." }
 */
private normalizeArgs<T extends { id?: string }>(
  args: string | T | { [key: string]: string }
): T {
  if (typeof args === 'string') {
    return { id: args } as T;
  }
  const normalized = { ...args } as T;
  // Handle snake_case variants
  Object.keys(args).forEach(key => {
    const camelKey = key.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
    if (camelKey !== key && !(camelKey in normalized)) {
      (normalized as any)[camelKey] = (args as any)[key];
    }
  });
  return normalized;
}
```

### CSS Scroll Zone Pattern

For independent scroll zones in flex layouts (e.g., three-panel layouts):

```tsx
<div className="flex flex-col min-h-0">
  <aside className="min-h-0 flex flex-col overflow-hidden">
    <div className="flex-1 overflow-y-auto">
      {/* Left scroll zone - scrolls independently */}
    </div>
  </aside>
  <main className="min-h-0 overflow-auto">
    {/* Main scroll zone - scrolls independently */}
  </main>
  <aside className="min-h-0 overflow-hidden">
    {/* Right scroll zone */}
  </aside>
</div>
```

**Key**: `min-h-0` on flex children allows them to shrink below their content size, enabling `overflow-auto` to work.

**Evidence from footprint analysis:**
- feature-task-page-scroll-zones: Scroll zone pattern implementation

## Relevant Files

Focus on the following files:

- `agentic/{branch-name}/state.json` - Contains the current workflow state
- `specs/sdlc_planner-*.md` - Plan files created by planning skills
- `specs/patch/patch-*.md` - Patch plan files
- `docs/README.md` - Project overview
- `docs/services/README.md` - Services documentation
- `docs/rules/backend/` - Backend rules
- `docs/rules/frontend/` - Frontend rules

Read `.claude/commands/conditional_docs.md` to check if your task requires additional documentation.

## Footprint Format

### Start Footprint

Create at: `agentic/{branch-name}/footprints/foot-impl-start.md`

```markdown
# Implementation Start Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: Implementation Start

## Plan Summary

### Plan Title
{title from plan file}

### Plan Type
{feature|bugfix|chore|patch|prototype}

### Total Phases
{count of phases}

### Total Steps
{count of step by step tasks}

## Implementation Strategy

### Execution Order
{list the order of execution based on plan's Step by Step Tasks}

### Expected Outcomes
{list expected outcomes from plan's Acceptance Criteria}

### Risk Mitigation
{any risks identified and how they will be handled}

## Initial State

### Files to Create
{list new files from plan}

### Files to Modify
{list existing files to be modified}

### Dependencies
{any dependencies to install}

## Next Steps

**Current Step**: Step 1
**Next Milestone**: Phase 1 Complete
```

### Phase Footprint

Create at: `agentic/{branch-name}/footprints/foot-impl-phase-{N}.md`

```markdown
# Implementation Phase {N} Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: Implementation Phase {N}
**Phase Name**: {phase name from plan}

## Phase Summary

### Completed Steps
{list steps completed in this phase with checkmarks}

### Files Created
{list new files created with brief description}

### Files Modified
{list files modified with brief description of changes}

## Implementation Details

### Key Changes
{describe the main changes made}

### Decisions Made
{list any implementation decisions made}

### Issues Encountered
{list any issues and how they were resolved}

### Tests Added/Modified
{list any tests added or modified}

## Validation

### Tests Run
{list tests executed and results}

### Manual Verification
{any manual verification performed}

## Commit Information

**Commit Message**: {commit message used}
**Files Committed**: {count of files}

## Progress

**Phases Completed**: {N} of {total}
**Steps Completed**: {count} of {total}
**Estimated Remaining**: {remaining phases/steps}

## Next Steps

**Next Phase**: Phase {N+1}
**Next Step**: {next step from plan}
```

### Completion Footprint

Create at: `agentic/{branch-name}/footprints/foot-impl-complete.md`

```markdown
# Implementation Complete Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: Implementation Complete

## Implementation Summary

### Plan Title
{title from plan file}

### Total Time
{implementation duration if trackable}

### Phases Completed
{list all phases with completion status}

### Steps Completed
{list all steps with completion status}

## Final State

### Files Created
{list all new files created}

### Files Modified
{list all files modified}

### Dependencies Added
{list any dependencies added}

## Validation Results

### Unit Tests
{results of unit test execution}

### Integration Tests
{results of integration tests if applicable}

### E2E Tests
{results of E2E tests if applicable}

### Build Validation
{results of build commands}

### Manual Verification
{any manual verification performed}

## Acceptance Criteria Check

{list each acceptance criteria from plan with pass/fail status}

## Commits Made

{list all commits made during implementation with messages}

## Notes

### Deviations from Plan
{any deviations from the original plan and reasons}

### Future Improvements
{any suggested future improvements}

### Lessons Learned
{any lessons learned during implementation}

## Next Steps

**Recommended Action**: /test
**Additional Testing**: Comprehensive testing phase will validate all changes
```

## State File Update

Update the state file at: `agentic/{branch-name}/state.json`

### During Implementation

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-impl-phase-{N}.md",
  "latest_command": "implement",
  "plan_file_path": "{plan file path}",
  "implementation_status": "in_progress",
  "implementation_progress": {
    "current_phase": {N},
    "total_phases": {total},
    "current_step": {step number},
    "total_steps": {total},
    "completed_steps": [{list of completed step numbers}],
    "commits": [{list of commit hashes}]
  },
  "next_command_metadata": {
    "command": "/implement",
    "category": "implementation",
    "confidence": "high",
    "reasoning": "Continuing implementation",
    "required_context": "{current phase/step info}"
  },
  "next_command": "/implement",
  "implementation_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}"
}
```

### After Completion

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-impl-complete.md",
  "latest_command": "implement",
  "plan_file_path": "{plan file path}",
  "implementation_status": "completed",
  "implementation_progress": {
    "current_phase": {total},
    "total_phases": {total},
    "current_step": {total},
    "total_steps": {total},
    "completed_steps": [{all step numbers}],
    "commits": [{all commit hashes}]
  },
  "next_command_metadata": {
    "command": "/test",
    "category": "testing",
    "confidence": "high",
    "reasoning": "Implementation complete, ready for comprehensive testing",
    "required_context": "All validation passed"
  },
  "next_command": "/test",
  "completion_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "bead_id": "{bead-id from state}",
  "validation_passed": true/false
}
```

---

## Implementation Steps

### Step 1: Initialize

1. Read state file from `agentic/{branch-name}/state.json`
2. Get `plan_file_path` from state or use provided argument
3. Read the plan file
4. Create `foot-impl-start.md` footprint
5. Commit the start footprint

```bash
git add agentic/{branch-name}/footprints/foot-impl-start.md
git commit -m "chore(impl): start implementation of {plan-title}"
```

### Step 2: Execute Plan Phases

For each phase in the plan:

1. Read the phase instructions from the plan
2. Execute each step in the phase:
   - Read relevant files
   - Make necessary code changes
   - Write/edit files
   - Run tests for changed components
3. After completing the phase:
   - Create `foot-impl-phase-{N}.md` footprint
   - Stage and commit all changes including footprint

```bash
git add {changed-files} agentic/{branch-name}/footprints/foot-impl-phase-{N}.md
git commit -m "{type}({scope}): {phase description}

Implements Phase {N} of {plan-title}

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

### Step 3: Run Validation

1. Execute all commands from the plan's "Validation Commands" section
2. Fix any failures before proceeding
3. If E2E tests are required, execute them

### Step 4: Finalize

1. Create `foot-impl-complete.md` footprint
2. Update state file with completion status
3. Final commit

```bash
git add agentic/{branch-name}/footprints/foot-impl-complete.md agentic/{branch-name}/state.json
git commit -m "chore(impl): complete implementation of {plan-title}

All validation passed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Required Actions

During implementation, you MUST:

1. **Read Plan**: Load the plan file and understand all requirements
2. **Create Start Footprint**: Document the beginning of implementation
3. **Execute Steps**: Follow each step in order, making necessary code changes
4. **Database migrations**: When the plan includes schema changes: use project's migration tool (e.g. Alembic for SQLAlchemy). EMM uses lakeFS for storage — no traditional DB migrations.
5. **Create Phase Footprints**: Document progress after each phase
6. **Commit Regularly**: Commit after each footprint with descriptive messages
7. **Run Validation**: Execute all validation commands from the plan
8. **Create Completion Footprint**: Document the final state
9. **Update State**: Keep state.json updated throughout

## Report

After completing implementation, return a JSON object:

```json
{
  "plan_file_path": "{path to plan file}",
  "implementation_status": "completed|failed|partial",
  "phases_completed": {count},
  "total_phases": {count},
  "steps_completed": {count},
  "total_steps": {count},
  "commits_made": [{list of commit hashes}],
  "validation_passed": true/false,
  "footprints": [
    "agentic/{branch-name}/footprints/foot-impl-start.md",
    "agentic/{branch-name}/footprints/foot-impl-phase-1.md",
    "agentic/{branch-name}/footprints/foot-impl-complete.md"
  ],
  "state_path": "agentic/{branch-name}/state.json",
  "notes": "{any additional notes about the implementation}"
}
```

## Error Handling

If implementation fails at any point:

1. Create a footprint documenting the failure
2. Update state file with `implementation_status: "failed"`
3. Include error details in the footprint
4. Do NOT mark steps as completed if they failed
5. Report the failure in the final JSON response

### Comms Escalation on Failure

If implementation fails and cannot be self-recovered, notify via comms (if `comms_enabled` in state.json):

```bash
curl -s -X POST http://localhost:8000/api/comms/notify \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "SDLC [{branch_name}] Implementation FAILED: {error_summary}. Workflow paused."
  }'
```

If `comms_enabled` is false or comms is unavailable, skip silently and report failure through normal state/footprint channels.

## Resuming Implementation

If resuming a partial implementation:

1. Read state file to get current progress
2. Find the last completed step
3. Continue from the next incomplete step
4. Create phase footprint when phase completes
5. Continue normal workflow
