---
name: chore
description: Chore Planning skill for creating plans for maintenance tasks, refactoring, documentation, or technical improvements. Use this skill when you need to plan a chore, refactor code, update dependencies, or when the classify skill routes a request as a chore. Creates plans in specs folder with footprints and state management.
---

# Chore Planning

Create a new plan to resolve a chore using the exact specified markdown `Plan Format`. Follow the `Instructions` to create the plan and use the `Relevant Files` to focus on the right files.

## Variables

branch_name: Current git branch name (e.g., "chore/update-deps")
descriptive_name: $1 (optional) - Short descriptive name for the chore

## Instructions

- IMPORTANT: You're writing a plan to resolve a chore based on the `Chore` that will add value to the application.
- IMPORTANT: The `Chore` describes the chore that will be resolved but remember we're not resolving the chore, we're creating the plan that will be used to resolve the chore based on the `Plan Format` below.
- You're writing a plan to resolve a chore, it should be simple but we need to be thorough and precise so we don't miss anything or waste time with any second round of changes.
- Create the plan in the `specs/` directory with filename: `sdlc_planner-{descriptive-name}.md`
  - Use `{descriptive-name}` from the variable if provided, otherwise derive from the chore (e.g., "update-readme", "fix-tests", "refactor-auth")
- Use the plan format below to create the plan.
- Research the codebase and put together a plan to accomplish the chore.
- IMPORTANT: Replace every <placeholder> in the `Plan Format` with the requested value. Add as much detail as needed to accomplish the chore.
- Use your reasoning model: THINK HARD about the plan and the steps to accomplish the chore.
- Respect requested files in the `Relevant Files` section.
- Start your research by reading the `docs/README.md` file.

**Backend (agentic-ai-landing-zone):**

- `viz/backend/` - FastAPI backend, LangGraph API
- `agents/` - LangGraph agents (task_manager, calendar_agent, finance_tracker, bookmark-classifier, content-classifier)
- `mcp-servers-tasks/`, `mcp-servers-lesson-credits/` - MCP servers (FastMCP, stdio)

**Frontend:**

- `viz/frontend/` - React + Vite + TanStack + Tailwind

**Infrastructure:**

- `scripts/` - Setup and utility scripts

## Relevant Files

Focus on the following files:

- `CLAUDE.md` - Project context and architecture
- `docs/` - Project documentation
- `viz/backend/`, `viz/frontend/`, `agents/`, `mcp-servers-*/` - Main code
- `scripts/` - Setup and utility scripts
- `.cursor/rules/project-structure.mdc` - Structure reference

- Read `.claude/commands/conditional_docs.md` to check if your task requires additional documentation

## Plan Format

```md
# Chore: <chore name>

## Metadata

branch_name: `{branch-name}`
descriptive_name: `{descriptive-name}`

## Chore Description

<describe the chore in detail>

## Relevant Files

Use these files to resolve the chore:

<find and list the files that are relevant to the chore describe why they are relevant in bullet points. If there are new files that need to be created to accomplish the chore, list them in an h3 'New Files' section.>

## Step by Step Tasks

IMPORTANT: Execute every step in order, top to bottom.

<list step by step tasks as h3 headers plus bullet points. use as many h3 headers as needed to accomplish the chore. Order matters, start with the foundational shared changes required to fix the chore then move on to the specific changes required to fix the chore. Your last step should be running the `Validation Commands` to validate the chore is complete with zero regressions.>

## Validation Commands

Execute every command to validate the chore is complete with zero regressions.

<list commands you'll use to validate with 100% confidence the chore is complete with zero regressions. every command must execute without errors so be specific about what you want to run to validate the chore is complete with zero regressions. Don't validate with curl commands.>

- `pytest` - Run backend tests (from project root)
- `cd viz/frontend && npm run test` - Run frontend tests (Vitest) if chore affects frontend
- `cd viz/frontend && npm run build` - Run frontend build if chore affects frontend

## Notes

<optionally list any additional notes or context that are relevant to the chore that will be helpful to the developer>
```

## Chore

Extract the chore details from user input or context provided.

---

## Footprint and State Management

After creating the plan, you MUST create a footprint and update the state file to document the planning process.

### Footprint Creation

Create a footprint file at: `agentic/{branch-name}/footprints/foot-chore-planning.md`

Where `{branch-name}` is the current git branch name (e.g., "chore/update-deps")

**Footprint Template:**

```markdown
# Chore Planning Footprint

**Date**: {ISO 8601 timestamp}
**Issue/Request**: {issue title}
**Type**: Chore Planning

## Input Analysis

### Issue Title
{issue title from issue_json}

### Issue Body
{issue body from issue_json}

### Context
{any additional context gathered during research}

## Planning Process

### Step 1: Codebase Research
- **Files Examined**: {list of files read during research}
- **Current State**: {description of current state}
- **Desired State**: {description of desired state after chore}

### Step 2: Scope Assessment
- **Changes Required**: {list of changes needed}
- **Affected Components**: {list of affected components}
- **Impact Assessment**: {scope and impact of the chore}

### Step 3: Plan Creation
- **Plan File**: {path to created plan file}
- **Tasks Created**: {count of tasks}
- **E2E Tests Required**: {yes/no}

## Planning Result

**Category**: chore
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
  "latest_footprint": "agentic/{branch-name}/footprints/foot-chore-planning.md",
  "latest_command": "chore",
  "plan_file_path": "specs/sdlc_planner-{descriptive-name}.md",
  "next_command_metadata": {
    "command": "/implement",
    "category": "chore",
    "confidence": "{high|medium|low}",
    "reasoning": "Chore plan created, ready for implementation",
    "required_context": "{plan file path}"
  },
  "next_command": "/implement",
  "planning_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "descriptive_name": "{descriptive-name}",
  "chore_title": "{chore title}",
  "need_e2e_tests": true/false
}
```

---

### Beads Integration

If `bead_id` exists in `agentic/{branch-name}/state.json`, update the bead after planning:

```bash
bd update $bead_id --notes="Plan created: chore - $plan_file_path"
```

The bead is created and managed by the `/sdlc` orchestrator. Planning skills only need to update it. If `bead_id` is null or missing, skip silently.

---

## Required Actions

After completing chore planning, you MUST:

1. **Create Plan File**: Write the plan to `specs/sdlc_planner-{descriptive-name}.md`
2. **Create Footprint**: Document planning process in `agentic/{branch-name}/footprints/foot-chore-planning.md`
3. **Update State**: Create/update `agentic/{branch-name}/state.json` with planning results

## Report

- IMPORTANT: Return a JSON object with the following structure:
```json
{
  "plan_file_path": "specs/sdlc_planner-{descriptive-name}.md",
  "need_e2e_tests": true/false,
  "footprint_path": "agentic/{branch-name}/footprints/foot-chore-planning.md",
  "state_path": "agentic/{branch-name}/state.json"
}
```

- `plan_file_path`: The full path to the created plan file
- `need_e2e_tests`: Set to `true` if the chore affects UI, user interactions, or frontend code that requires e2e testing. Set to `false` if the chore is backend-only or doesn't require e2e tests.
- `footprint_path`: The full path to the created footprint file
- `state_path`: The full path to the state file

**Determining need_e2e_tests:**
- Set to `true` if the chore:
  - Affects UI components or user interactions
  - Modifies frontend code (viz/frontend/, .tsx, .jsx, .css files)
  - Requires user-facing validation
- Set to `false` if the chore:
  - Is backend-only (services/core, services/auth, services/gateway, etc.)
  - Doesn't affect UI or user interactions
  - Only modifies configuration, documentation, or backend logic
  - Most chores are backend-only and should set this to `false`
