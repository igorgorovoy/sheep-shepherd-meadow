---
name: bug
description: Bug Planning skill for creating plans to resolve bugs and issues. Use this skill when you need to plan a bug fix, investigate and resolve an error, or when the classify skill routes a request as a bugfix. Creates plans in specs folder with footprints and state management.
---

# Bug Planning

Create a new plan to resolve a bug using the exact specified markdown `Plan Format`. Follow the `Instructions` to create the plan and use the `Relevant Files` to focus on the right files.

## Variables

branch_name: Current git branch name (e.g., "fix/login-error")
descriptive_name: $1 (optional) - Short descriptive name for the bug fix

## Instructions

- IMPORTANT: You're writing a plan to resolve a bug based on the `Bug` that will add value to the application.
- IMPORTANT: The `Bug` describes the bug that will be resolved but remember we're not resolving the bug, we're creating the plan that will be used to resolve the bug based on the `Plan Format` below.
- You're writing a plan to resolve a bug, it should be thorough and precise so we fix the root cause and prevent regressions.
- Create the plan in the `specs/` directory with filename: `sdlc_planner-{descriptive-name}.md`
  - Use `{descriptive-name}` from the variable if provided, otherwise derive from the bug (e.g., "fix-login-error", "resolve-timeout", "patch-memory-leak")
- Use the plan format below to create the plan.
- Research the codebase to understand the bug, reproduce it, and put together a plan to fix it.
- IMPORTANT: Replace every <placeholder> in the `Plan Format` with the requested value. Add as much detail as needed to fix the bug.
- Use your reasoning model: THINK HARD about the bug, its root cause, and the steps to fix it properly.
- IMPORTANT: Be surgical with your bug fix, solve the bug at hand and don't fall off track.
- IMPORTANT: We want the minimal number of changes that will fix and address the bug.
- If you need a new library, use `bun add <library_name>` or `bun add -D <library_name>` and be sure to report it in the `Notes` section of the `Plan Format`.
- IMPORTANT: If the bug affects the UI or user interactions:
  - Add a task in the `Step by Step Tasks` section to create a separate E2E test file in `.claude/commands/e2e/test_<descriptive_name>.md` based on examples in that directory
  - Add E2E test validation to your Validation Commands section
  - IMPORTANT: When you fill out the `Plan Format: Relevant Files` section, add an instruction to read `.claude/commands/test_e2e.md`, and `.claude/commands/e2e/test_theme_change.md` to understand how to create an E2E test file. List your new E2E test file to the `Plan Format: New Files` section.
  - IMPORTANT: All E2E tests must include authentication check at the beginning: After navigating to the application URL, check if user is logged in. If not logged in, perform login using credentials from `services/auth/deploy/.env` (read `ROOT_USER` and `ROOT_PASSWORD` from lines 6-7). This ensures tests work reliably in all environments.
  - To be clear, we're not creating a new E2E test file, we're creating a task to create a new E2E test file in the `Plan Format` below
- Respect requested files in the `Relevant Files` section.
- Start your research by reading the `docs/README.md` file.

**Backend (agentic-ai-landing-zone):**

- `viz/backend/` - FastAPI backend, LangGraph API
- `agents/` - LangGraph agents
- `mcp-servers-tasks/`, `mcp-servers-lesson-credits/` - MCP servers

**Frontend:**

- `viz/frontend/` - React + Vite + TanStack + Tailwind

## Relevant Files

Focus on the following files:

- `CLAUDE.md` - Project context
- `docs/` - Project documentation
- `viz/backend/`, `viz/frontend/`, `agents/`, `mcp-servers-*/` - Main code
- `.cursor/rules/project-structure.mdc` - Structure reference

## Plan Format

```md
# Bug: <bug name>

## Metadata

branch_name: `{branch-name}`
descriptive_name: `{descriptive-name}`

## Bug Description

<describe the bug in detail, including symptoms and expected vs actual behavior>

## Problem Statement

<clearly define the specific problem that needs to be solved>

## Solution Statement

<describe the proposed solution approach to fix the bug>

## Steps to Reproduce

<list exact steps to reproduce the bug>

## Root Cause Analysis

<analyze and explain the root cause of the bug>

## Relevant Files

Use these files to fix the bug:

<find and list the files that are relevant to the bug describe why they are relevant in bullet points. If there are new files that need to be created to fix the bug, list them in an h3 'New Files' section.>

## Step by Step Tasks

IMPORTANT: Execute every step in order, top to bottom.

<list step by step tasks as h3 headers plus bullet points. use as many h3 headers as needed to fix the bug. Order matters, start with the foundational shared changes required to fix the bug then move on to the specific changes required to fix the bug. Include tests that will validate the bug is fixed with zero regressions.>

<If the bug affects UI, include a task to create a E2E test file. Your task should look like: "Read `.claude/commands/e2e/test_theme_change.md` and `.claude/commands/e2e/test_create_lead.md` and create a new E2E test file in `.claude/commands/e2e/test_<descriptive_name>.md` that validates the bug is fixed, be specific with the steps to prove the bug is fixed. We want the minimal set of steps to validate the bug is fixed and screen shots to prove it if possible. IMPORTANT: The E2E test must include authentication check at the beginning: After navigating to the application URL, check if user is logged in. If not logged in, perform login using credentials from `services/auth/deploy/.env` (read `ROOT_USER` and `ROOT_PASSWORD` from lines 6-7). This ensures tests work reliably in all environments.">

<Your last step should be running the `Validation Commands` to validate the bug is fixed with zero regressions.>

## Validation Commands

Execute every command to validate the bug is fixed with zero regressions.

<list commands you'll use to validate with 100% confidence the bug is fixed with zero regressions. every command must execute without errors so be specific about what you want to run to validate the bug is fixed with zero regressions. Include commands to reproduce the bug before and after the fix.>

<If you created an E2E test, include the following validation step: "Read .claude/commands/test_e2e.md`, then read and execute your new E2E `.claude/commands/e2e/test_<descriptive_name>.md` test file to validate this functionality works.">

- `pytest` - Run backend tests (from project root)
- `cd viz/frontend && npm run test` - Run frontend tests if bug affects frontend
- `cd viz/frontend && npm run build` - Run frontend build if bug affects frontend

## Notes

<optionally list any additional notes or context that are relevant to the bug that will be helpful to the developer>
```

## Bug

Extract the bug details from user input or context provided.

---

## Diagnostics (agentic-ai-landing-zone)

### FastAPI / Python

- Check env vars: missing config can cause silent failures
- FastAPI: validate Pydantic models, check router order for 404s

---

## AI Tool Bug Patterns (MCP / LangGraph)

### AI Tool Bug Indicators

Classify as bug when:
- "AI tool doesn't accept [format]" → argument normalization bug
- "inconsistent behavior across similar tools" → pattern consistency bug
- "tool works with object but not string" → argument flexibility bug
- "AI tool returns wrong format" → response format bug
- "tool description doesn't match behavior" → documentation bug

### AI Tool Bug Fix Pattern

When fixing AI tool bugs:

1. **Identify the affected tool(s)**
   - Find the tool definition in `mcp-servers-*/` or `agents/`
   - Check the argument schema

2. **Review similar tools for normalization pattern**
   - Look at how other tools handle similar arguments
   - Check for existing normalization utilities

3. **Apply flexible argument handling**
   - Accept string (direct ID)
   - Accept object with camelCase key
   - Accept object with snake_case key

4. **Test all argument formats**
   - Test with string: `"attachment-id"`
   - Test with camelCase: `{ attachmentId: "..." }`
   - Test with snake_case: `{ attachment_id: "..." }`

5. **Update tool description for accepted formats**
   - Document all accepted argument formats
   - Update system prompt if needed

**Code Pattern for Flexible Arguments (Python):**

```python
def normalize_args(args: str | dict) -> dict:
    if isinstance(args, str):
        return {"id": args}
    return {k.replace("_", ""): v for k, v in args.items()}
```

---

## Footprint and State Management

After creating the plan, you MUST create a footprint and update the state file to document the planning process.

### Footprint Creation

Create a footprint file at: `agentic/{branch-name}/footprints/foot-bug-planning.md`

Where `{branch-name}` is the current git branch name (e.g., "fix/login-error")

**Footprint Template:**

```markdown
# Bug Planning Footprint

**Date**: {ISO 8601 timestamp}
**Issue/Request**: {issue title}
**Type**: Bug Planning

## Input Analysis

### Issue Title
{issue title from issue_json}

### Issue Body
{issue body from issue_json}

### Context
{any additional context gathered during research}

## Planning Process

### Step 1: Bug Investigation
- **Files Examined**: {list of files read during investigation}
- **Error Messages**: {any error messages found}
- **Reproduction Steps**: {steps to reproduce the bug}

### Step 2: Root Cause Analysis
- **Root Cause**: {identified root cause}
- **Affected Components**: {list of affected components}
- **Impact Assessment**: {severity and scope of the bug}

### Step 3: Solution Design
- **Proposed Fix**: {description of the fix}
- **Files to Modify**: {list of files that need changes}
- **Risk Assessment**: {potential risks of the fix}

### Step 4: Plan Creation
- **Plan File**: {path to created plan file}
- **Tasks Created**: {count of tasks}
- **E2E Tests Required**: {yes/no}

## Planning Result

**Category**: bugfix
**Plan File Path**: specs/sdlc_planner-{descriptive-name}.md
**Confidence**: {high|medium|low}

**Key Decisions Made**:
{list key decisions made during planning}

**Risks Identified**:
{list potential risks}

## Next Steps

**Next Command**: /implement
**Required Context**: {context needed for implementation}
```

### State File Update

Create or update the state file at: `agentic/{branch-name}/state.json`

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-bug-planning.md",
  "latest_command": "bug",
  "plan_file_path": "specs/sdlc_planner-{descriptive-name}.md",
  "next_command_metadata": {
    "command": "/implement",
    "category": "bugfix",
    "confidence": "{high|medium|low}",
    "reasoning": "Bug plan created, ready for implementation",
    "required_context": "{plan file path}"
  },
  "next_command": "/implement",
  "planning_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "descriptive_name": "{descriptive-name}",
  "bug_title": "{bug title}",
  "need_e2e_tests": true/false
}
```

---

### Beads Integration

If `bead_id` exists in `agentic/{branch-name}/state.json`, update the bead after planning:

```bash
bd update $bead_id --notes="Plan created: bugfix - $plan_file_path"
```

The bead is created and managed by the `/sdlc` orchestrator. Planning skills only need to update it. If `bead_id` is null or missing, skip silently.

---

## Required Actions

After completing bug planning, you MUST:

1. **Create Plan File**: Write the plan to `specs/sdlc_planner-{descriptive-name}.md`
2. **Create Footprint**: Document planning process in `agentic/{branch-name}/footprints/foot-bug-planning.md`
3. **Update State**: Create/update `agentic/{branch-name}/state.json` with planning results

## Report

- IMPORTANT: Return a JSON object with the following structure:
```json
{
  "plan_file_path": "specs/sdlc_planner-{descriptive-name}.md",
  "need_e2e_tests": true/false,
  "footprint_path": "agentic/{branch-name}/footprints/foot-bug-planning.md",
  "state_path": "agentic/{branch-name}/state.json"
}
```

- `plan_file_path`: The full path to the created plan file
- `need_e2e_tests`: Set to `true` if the bug affects UI, user interactions, or frontend code that requires e2e testing. Set to `false` if the bug is backend-only or doesn't require e2e tests.
- `footprint_path`: The full path to the created footprint file
- `state_path`: The full path to the state file

**Determining need_e2e_tests:**
- Set to `true` if the bug:
  - Affects UI components or user interactions
  - Is visible to users in the frontend
  - Requires frontend changes to fix (viz/frontend/, .tsx, .jsx, .css files)
  - Needs user-facing validation to confirm fix
- Set to `false` if the bug:
  - Is backend-only (services/core, services/auth, services/gateway, etc.)
  - Doesn't affect UI or user interactions
  - Only requires backend logic changes
