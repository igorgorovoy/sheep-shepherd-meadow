---
name: patch
description: Patch Planning skill for creating focused plans to resolve specific issues with minimal, targeted changes. Use this skill when you need to plan a quick fix, small improvement, or when the classify skill routes a request as a patch. Creates plans in specs/patch folder with footprints and state management.
---

# Patch Planning

Create a **focused patch plan** to resolve a specific issue. Follow the `Instructions` to create a concise plan that addresses the issue with minimal, targeted changes.

## Variables

branch_name: Current git branch name (e.g., "patch/fix-button-color")
descriptive_name: $1 (optional) - Short descriptive name for the patch
review_change_request: $2
spec_path: $3 if provided, otherwise leave it blank
agent_name: $4 if provided, otherwise use 'patch_agent'
issue_screenshots: $ARGUMENT (optional) - comma-separated list of screenshot paths if provided

## Instructions

- IMPORTANT: You're creating a patch plan to fix a specific review issue. Keep changes small, focused, and targeted
- Read the original specification (spec) file at `spec_path` if provided to understand the context and requirements
- IMPORTANT: Use the `review_change_request` to understand exactly what needs fixing and use it as the basis for your patch plan
- If `issue_screenshots` are provided, examine them to better understand the visual context of the issue
- Create the patch plan in `specs/patch/` directory with filename: `patch-{descriptive-name}.md`
  - Use `{descriptive-name}` from the variable if provided, otherwise derive from the issue (e.g., "fix-button-color", "update-validation", "correct-layout")
- IMPORTANT: This is a PATCH - keep the scope minimal. Only fix what's described in the `review_change_request` and nothing more. Address only the `review_change_request`.
- Run `git diff --stat`. If changes are available, use them to understand what's been done in the codebase and so you can understand the exact changes you should detail in the patch plan.
- Ultra think about the most efficient way to implement the solution with minimal code changes
- Base your `Plan Format: Validation` on the validation steps from `spec_path` if provided
  - If any tests fail in the validation steps, you must fix them.
  - If not provided, READ `.claude/commands/test.md: ## Test Execution Sequence` and execute the tests to understand the tests that need to be run to validate the patch.
- Replace every <placeholder> in the `Plan Format` with specific implementation details

**Backend Service:**

- `viz/backend/**`, `agents/**`, `mcp-servers-*/**` - Backend (Python, FastAPI, LangGraph)
  - `src/modules/auth/` - Authentication & authorization
  - `src/modules/core/` - Core business logic (CRM, Projects, Orders)
  - `src/modules/fs/` - File storage (S3)
  - `src/modules/gateway/` - API gateway layer
  - `src/modules/notification/` - Notifications
  - `src/modules/user-log/` - User activity audit

**Frontend Applications:**

- `viz/frontend/**` - Frontend (React, Vite, TanStack, Tailwind)

**Note:** Services `core`, `auth`, `fs`, `gateway`, `private-gateway`, `notification`, `user-log`, `mcp-server`, `client`, `admin`, `investor-client` are DEPRECATED (reference only).

## Relevant Files

Focus on the following files:
- `README.md` - Contains the project overview and instructions.
- `app/server/**` - Contains the codebase server.
- `app/client/**` - Contains the codebase client.
- `scripts/**` - Contains the scripts to start and stop the server + client.

- Read `.claude/commands/conditional_docs.md` to check if your task requires additional documentation
- If your task matches any of the conditions listed, reference those documentation files to understand the context better when creating your patch plan

Ignore all other files in the codebase.

## Plan Format

```md
# Patch: <concise patch title>

## Metadata
branch_name: `{branch-name}`
descriptive_name: `{descriptive-name}`
review_change_request: `{review_change_request}`

## Issue Summary
**Original Spec:** <spec_path>
**Issue:** <brief description of the review issue based on the `review_change_request`>
**Solution:** <brief description of the solution approach based on the `review_change_request`>

## Files to Modify
Use these files to implement the patch:

<list only the files that need changes - be specific and minimal>

## Implementation Steps
IMPORTANT: Execute every step in order, top to bottom.

<list 2-5 focused steps to implement the patch. Each step should be a concrete action.>

### Step 1: <specific action>
- <implementation detail>
- <implementation detail>

### Step 2: <specific action>
- <implementation detail>
- <implementation detail>

<continue as needed, but keep it minimal>

## Validation
Execute every command to validate the patch is complete with zero regressions.

<list 1-5 specific commands or checks to verify the patch works correctly>

## Patch Scope
**Lines of code to change:** <estimate>
**Risk level:** <low|medium|high>
**Testing required:** <brief description>
```

---

## Footprint and State Management

After creating the plan, you MUST create a footprint and update the state file to document the planning process.

### Footprint Creation

Create a footprint file at: `agentic/{branch-name}/footprints/foot-patch-planning.md`

Where `{branch-name}` is the current git branch name (e.g., "patch/fix-button-color")

**Footprint Template:**

```markdown
# Patch Planning Footprint

**Date**: {ISO 8601 timestamp}
**Issue/Request**: {review_change_request summary}
**Type**: Patch Planning

## Input Analysis

### Change Request
{review_change_request}

### Original Spec
{spec_path if provided}

### Context
{any additional context, git diff analysis}

## Planning Process

### Step 1: Issue Analysis
- **Issue Type**: {type of issue - visual, logic, validation, etc.}
- **Affected Files**: {list of files affected}
- **Scope Assessment**: {minimal/small/medium}

### Step 2: Solution Design
- **Proposed Fix**: {description of the fix}
- **Lines to Change**: {estimate}
- **Risk Level**: {low|medium|high}

### Step 3: Plan Creation
- **Plan File**: {path to created patch plan file}
- **Steps Created**: {count of steps}

## Planning Result

**Category**: patch
**Plan File Path**: specs/patch/patch-{descriptive-name}.md
**Confidence**: {high|medium|low}

**Key Decisions Made**:
{list key decisions made during planning}

## Next Steps

**Next Command**: /implement
**Required Context**: {context needed for implementation}
```

### State File Update

Create or update the state file at: `agentic/{branch-name}/state.json`

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-patch-planning.md",
  "latest_command": "patch",
  "plan_file_path": "specs/patch/patch-{descriptive-name}.md",
  "next_command_metadata": {
    "command": "/implement",
    "category": "patch",
    "confidence": "{high|medium|low}",
    "reasoning": "Patch plan created, ready for implementation",
    "required_context": "{plan file path}"
  },
  "next_command": "/implement",
  "planning_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "descriptive_name": "{descriptive-name}",
  "review_change_request": "{review_change_request}",
  "spec_path": "{spec_path if provided}",
  "need_e2e_tests": false
}
```

---

### Beads Integration

If `bead_id` exists in `agentic/{branch-name}/state.json`, update the bead after planning:

```bash
bd update $bead_id --notes="Plan created: patch - $plan_file_path"
```

The bead is created and managed by the `/sdlc` orchestrator. Planning skills only need to update it. If `bead_id` is null or missing, skip silently.

---

## Required Actions

After completing patch planning, you MUST:

1. **Create Plan File**: Write the plan to `specs/patch/patch-{descriptive-name}.md`
2. **Create Footprint**: Document planning process in `agentic/{branch-name}/footprints/foot-patch-planning.md`
3. **Update State**: Create/update `agentic/{branch-name}/state.json` with planning results

## Report

- IMPORTANT: Return a JSON object with the following structure:
```json
{
  "plan_file_path": "specs/patch/patch-{descriptive-name}.md",
  "need_e2e_tests": false,
  "footprint_path": "agentic/{branch-name}/footprints/foot-patch-planning.md",
  "state_path": "agentic/{branch-name}/state.json"
}
```

- `plan_file_path`: The full path to the created patch plan file
- `need_e2e_tests`: Usually `false` for patches unless the patch specifically requires visual validation
- `footprint_path`: The full path to the created footprint file
- `state_path`: The full path to the state file
