---
name: feature
description: Feature Planning skill for creating implementation plans for new features. Use this skill when you need to plan a new feature, create a detailed implementation plan, or when the classify skill routes a request as a feature. Creates plans in specs folder with footprints and state management.
---

# Feature Planning

Create a new plan to implement a feature using the exact specified markdown `Plan Format`. Follow the `Instructions` to create the plan and use the `Relevant Files` to focus on the right files.

## Variables

branch_name: Current git branch name (e.g., "feature/new-UI")
descriptive_name: $1 (optional) - Short descriptive name for the feature

## Instructions

- IMPORTANT: You're writing a plan to implement a net new feature based on the `Feature` that will add value to the application.
- IMPORTANT: The `Feature` describes the feature that will be implemented but remember we're not implementing a new feature, we're creating the plan that will be used to implement the feature based on the `Plan Format` below.
- Create the plan in the `specs/` directory with filename: `sdlc_planner-{descriptive-name}.md`
  - Use `{descriptive-name}` from the variable if provided, otherwise derive from the feature (e.g., "add-auth-system", "implement-search", "create-dashboard")
- Use the `Plan Format` below to create the plan.
- Research the codebase to understand existing patterns, architecture, and conventions before planning the feature.
- IMPORTANT: Replace every <placeholder> in the `Plan Format` with the requested value. Add as much detail as needed to implement the feature successfully.
- Use your reasoning model: THINK HARD about the feature requirements, design, and implementation approach.
- Follow existing patterns and conventions in the codebase. Don't reinvent the wheel.
- Design for extensibility and maintainability.
- If you need a new library: Python → `uv add <pkg>` or `pip install`; Frontend → `npm install <pkg>` in viz/frontend. Report in `Notes` section.
- **Feature Flags** (optional): For toggleable features, consider adding a feature flag task if the project supports it.

### Mobile UI Checklist (When mobile/responsive mentioned)

When planning features that involve mobile UI or responsive design, include these considerations in the plan:

- [ ] **Touch target specification**: Minimum 44px for all interactive elements (buttons, links, form controls)
- [ ] **Sheet/Drawer overlay positioning**: Account for fixed elements
  - X button positioning (typically `absolute right-4 top-4`)
  - Plan for visible headers vs sr-only headers
  - Consider `pr-12` padding to avoid X button overlap with content
- [ ] **Fixed element accommodation**: Headers, floating buttons, navigation bars
- [ ] **i18n file updates**: Check all locales need updates (en, es, ru, uk)
- [ ] **Desktop experience preservation**: Verify desktop layout unchanged after mobile work
- [ ] **Breakpoint testing plan**: Test at 320px, 768px, 1024px, 1280px

**Evidence from footprint analysis:**
- X button intersection issues required mid-workflow fixes (foot-impl-start-sidebar-fix.md)
- Touch targets had to be added during implementation (foot-feature-planning-mobile.md)
- i18n for 4 locales was added after initial plan (foot-commit-pr.md)
- If the feature includes UI components and the project has E2E tests (e.g. Playwright): add a task to create/update E2E tests. For agentic-ai-landing-zone: add Vitest unit tests in `viz/frontend/src/**/__tests__/` and pytest tests in `viz/backend/tests/` or `agents/tests/` as appropriate.
- Respect requested files in the `Relevant Files` section.
- Start your research by reading `CLAUDE.md` and `docs/` if present.

**Backend (agentic-ai-landing-zone):**

- `viz/backend/` — FastAPI backend, LangGraph API
- `agents/` — LangGraph agents (task_manager, calendar_agent, finance_tracker, bookmark-classifier, content-classifier)
- `mcp-servers-tasks/` — Task Manager MCP server (FastMCP, stdio)

**Frontend:**

- `viz/frontend/` — React + Vite + TanStack + Tailwind (npm run dev → localhost:5173)

**See** `.cursor/rules/project-structure.mdc` for full structure and commands.

## Relevant Files

Focus on files relevant to the feature. For agentic-ai-landing-zone:

- `CLAUDE.md` — Project context and architecture
- `viz/backend/`, `viz/frontend/`, `agents/`, `mcp-servers-tasks/` — Main code
- `scripts/` — Setup and utility scripts
- `.cursor/rules/project-structure.mdc` — Structure reference

Include only files that exist and are relevant. Ignore non-existent paths.

## Plan Format

```md
# Feature: <feature name>

## Metadata

branch_name: `{branch-name}`
descriptive_name: `{descriptive-name}`

## Feature Description

<describe the feature in detail, including its purpose and value to users>

## User Story

As a <type of user>
I want to <action/goal>
So that <benefit/value>

## Problem Statement

<clearly define the specific problem or opportunity this feature addresses>

## Solution Statement

<describe the proposed solution approach and how it solves the problem>

## Relevant Files

Use these files to implement the feature:

<find and list the files that are relevant to the feature describe why they are relevant in bullet points. If there are new files that need to be created to implement the feature, list them in an h3 'New Files' section.>

## Implementation Plan

### Phase 1: Foundation

<describe the foundational work needed before implementing the main feature>

### Phase 2: Core Implementation

<describe the main implementation work for the feature>

### Phase 3: Integration

<describe how the feature will integrate with existing functionality>

## Step by Step Tasks

IMPORTANT: Execute every step in order, top to bottom.

<list step by step tasks as h3 headers plus bullet points. use as many h3 headers as needed to implement the feature. Order matters, start with the foundational shared changes required then move on to the specific implementation. Include creating tests throughout the implementation process.>

<If the feature affects UI, include a task to create a E2E test file (like `.claude/commands/e2e/test_theme_change.md` and `.claude/commands/e2e/test_create_lead.md`) as one of your early tasks. That e2e test should validate the feature works as expected, be specific with the steps to demonstrate the new functionality. We want the minimal set of steps to validate the feature works as expected and screen shots to prove it if possible. IMPORTANT: The E2E test must include authentication check at the beginning: After navigating to the application URL, check if user is logged in. If not logged in, perform login using credentials from `services/auth/deploy/.env` (read `ROOT_USER` and `ROOT_PASSWORD` from lines 6-7). This ensures tests work reliably in all environments.>

<Your last step should be running the `Validation Commands` to validate the feature works correctly with zero regressions.>

## Testing Strategy

### Unit Tests

<describe unit tests needed for the feature>

### Edge Cases

<list edge cases that need to be tested>

## Acceptance Criteria

<list specific, measurable criteria that must be met for the feature to be considered complete>

## Validation Commands

Execute every command to validate the feature works correctly with zero regressions.

<list commands you'll use to validate with 100% confidence the feature is implemented correctly with zero regressions. every command must execute without errors so be specific about what you want to run to validate the feature works as expected. Include commands to test the feature end-to-end.>

<If you created an E2E test, include the following validation step: `Read .claude/commands/test_e2e.md`, then read and execute your new E2E `.claude/commands/e2e/test_<descriptive_name>.md` test file to validate this functionality works.>

- `pytest` - Run backend tests (from project root)
- `cd viz/frontend && npm run test` - Run frontend tests
- `cd viz/frontend && npm run build` - Run frontend build
- Read `docs/services/<service-name>.md` files for each affected service and include any special test rules, test commands, or test requirements documented in the "Тесты" or "Testing" sections

## Notes

<optionally list any additional notes, future considerations, or context that are relevant to the feature that will be helpful to the developer>
```

## Feature

Extract the feature details from user input or context provided.

---

## Special Feature Types

### AI Tool Feature Checklist

When implementing AI tools:

1. **Tool Definition**
   - Clear description of when to use
   - Argument schema with types
   - Example usage in description

2. **Argument Normalization**
   - Apply flexible handling pattern
   - Support string, camelCase, snake_case

   ```typescript
   private normalizeArgs<T extends { id?: string }>(
     args: string | T | { [key: string]: string }
   ): T {
     if (typeof args === 'string') {
       return { id: args } as T;
     }
     const normalized = { ...args } as T;
     Object.keys(args).forEach(key => {
       const camelKey = key.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
       if (camelKey !== key && !(camelKey in normalized)) {
         (normalized as any)[camelKey] = (args as any)[key];
       }
     });
     return normalized;
   }
   ```

3. **Response Format**
   - Consistent success/error structure
   - Match existing tool patterns

4. **Testing**
   - Unit tests for each argument format
   - Manual prompt testing documented
   - System prompt updated

### BDD Testing Infrastructure

For features requiring E2E behavior testing:

```
specs/{feature}/
├── features/           # Gherkin .feature files
├── step-definitions/   # Jest step implementations
└── support/            # Shared utilities
```

Use jest-cucumber for step definition binding:

```typescript
import { defineFeature, loadFeature } from 'jest-cucumber';

const feature = loadFeature('./features/my-feature.feature');

defineFeature(feature, test => {
  test('Scenario name', ({ given, when, then }) => {
    given('precondition', async () => { ... });
    when('action', async () => { ... });
    then('result', async () => { ... });
  });
});
```

### FastAPI Router Module Pattern (viz/backend)

**Module Structure**:
```
viz_backend/routers/
├── __init__.py
├── leads.py          # Lead routes
├── contacts.py       # Contact routes
└── ...
viz_backend/
├── store.py          # Data/store layer
└── ...
```

**Router** (`routers/leads.py`):
```python
from fastapi import APIRouter
router = APIRouter(prefix="/leads", tags=["leads"])

@router.get("")
async def list_leads(limit: int = 10, offset: int = 0):
    # Use store or repository layer
    return {"data": [], "total": 0}
```

### Event Bus / Async (EMM)

**When to Use**:
- Cross-module communication
- Async workflows
- Audit logging

**Pattern** (Python): Use FastAPI BackgroundTasks or LangGraph for agent workflows.

### Mobile-Responsive Design Checklist

**Breakpoints** (Tailwind):
- `sm:` - 640px (mobile landscape)
- `md:` - 768px (tablet) - Primary breakpoint
- `lg:` - 1024px (desktop)
- `xl:` - 1280px (wide desktop)

**Touch Targets**:
- Minimum 44px x 44px (Apple HIG)
- Use `p-3` (12px) for buttons minimum
- Space interactive elements 8px apart

**Layout Patterns**:
```tsx
// Mobile-first responsive grid
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>

// Responsive navigation
<nav className="flex flex-col md:flex-row gap-2 md:gap-4">
  <Button>Item 1</Button>
  <Button>Item 2</Button>
</nav>

// Conditional rendering for mobile
<div className="hidden md:block">Desktop only content</div>
<div className="block md:hidden">Mobile only content</div>
```

**Testing**:
```typescript
// Playwright mobile viewport
test.use({ viewport: { width: 375, height: 667 } })  // iPhone SE

test('mobile navigation', async ({ page }) => {
  await page.goto('/')
  const menu = page.getByRole('button', { name: 'Menu' })
  await expect(menu).toBeVisible()
})
```

**Checklist**:
- [ ] All interactive elements >= 44px touch target
- [ ] Grid/flex layouts stack on mobile
- [ ] Text readable without zoom (16px+ body)
- [ ] Forms single-column on mobile
- [ ] Navigation accessible on small screens
- [ ] Tested at 375px width (mobile) and 768px (tablet)

**Evidence from footprint analysis:**
- Mobile responsive issues found in 6 UI implementations

### Playwright Page Object Model Pattern

**Structure**:
```typescript
// tests/pages/leads.page.ts
export class LeadsPage {
  constructor(private page: Page) {}

  // Locators
  get createButton() {
    return this.page.getByRole('button', { name: 'Create Lead' })
  }

  get nameInput() {
    return this.page.getByLabel('Name')
  }

  get saveButton() {
    return this.page.getByRole('button', { name: 'Save' })
  }

  // Actions
  async goto() {
    await this.page.goto('/crm/leads')
  }

  async createLead(name: string, phone: string) {
    await this.createButton.click()
    await this.nameInput.fill(name)
    await this.page.getByLabel('Phone').fill(phone)
    await this.saveButton.click()
  }

  // Assertions
  async expectLeadVisible(name: string) {
    await expect(
      this.page.getByText(name)
    ).toBeVisible()
  }
}
```

**Usage**:
```typescript
test('create lead', async ({ page }) => {
  const leadsPage = new LeadsPage(page)

  await leadsPage.goto()
  await leadsPage.createLead('John Doe', '+1234567890')
  await leadsPage.expectLeadVisible('John Doe')
})
```

**Benefits**:
- Reusable page interactions
- Single source of truth for selectors
- Easier to maintain when UI changes
- More readable tests

**Evidence from footprint analysis:**
- Playwright POM discovered independently 3 times during migration

### WebSocket/Subscription Features

When implementing real-time features:

1. **Protocol**: Use graphql-ws for GraphQL subscriptions
2. **Server-side**:
   - Configure PubSub module in NestJS
   - Add subscription resolvers
   - Define subscription types in schema
3. **Client-side**:
   - Use graphql-ws client
   - Handle connection lifecycle
4. **Testing**: Test with ws client or graphql-ws playground

---

## Footprint and State Management

After creating the plan, you MUST create a footprint and update the state file to document the planning process.

### Footprint Creation

Create a footprint file at: `agentic/{branch-name}/footprints/foot-feature-planning.md`

Where `{branch-name}` is the current git branch name (e.g., "feature/new-UI")

**Footprint Template:**

```markdown
# Feature Planning Footprint

**Date**: {ISO 8601 timestamp}
**Issue/Request**: {issue title}
**Type**: Feature Planning

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
- **Patterns Identified**: {existing patterns found}
- **Architecture Notes**: {relevant architecture observations}

### Step 2: Requirements Analysis
- **Functional Requirements**: {list functional requirements}
- **Non-Functional Requirements**: {list non-functional requirements}
- **Dependencies**: {list dependencies}

### Step 3: Plan Creation
- **Plan File**: {path to created plan file}
- **Phases Identified**: {list of phases}
- **Tasks Created**: {count of tasks}
- **E2E Tests Required**: {yes/no}

## Planning Result

**Category**: feature
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
  "latest_footprint": "agentic/{branch-name}/footprints/foot-feature-planning.md",
  "latest_command": "feature",
  "plan_file_path": "specs/sdlc_planner-{descriptive-name}.md",
  "next_command_metadata": {
    "command": "/implement",
    "category": "feature",
    "confidence": "{high|medium|low}",
    "reasoning": "Feature plan created, ready for implementation",
    "required_context": "{plan file path}"
  },
  "next_command": "/implement",
  "planning_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "descriptive_name": "{descriptive-name}",
  "feature_title": "{feature title}",
  "need_e2e_tests": true/false
}
```

---

### Beads Integration

If `bead_id` exists in `agentic/{branch-name}/state.json`, update the bead after planning:

```bash
bd update $bead_id --notes="Plan created: feature - $plan_file_path"
```

The bead is created and managed by the `/sdlc` orchestrator. Planning skills only need to update it. If `bead_id` is null or missing, skip silently.

---

## Required Actions

After completing feature planning, you MUST:

1. **Create Plan File**: Write the plan to `specs/sdlc_planner-{descriptive-name}.md`
2. **Create Footprint**: Document planning process in `agentic/{branch-name}/footprints/foot-feature-planning.md`
3. **Update State**: Create/update `agentic/{branch-name}/state.json` with planning results

## Report

- IMPORTANT: Return a JSON object with the following structure:
```json
{
  "plan_file_path": "specs/sdlc_planner-{descriptive-name}.md",
  "need_e2e_tests": true/false,
  "footprint_path": "agentic/{branch-name}/footprints/foot-feature-planning.md",
  "state_path": "agentic/{branch-name}/state.json"
}
```

- `plan_file_path`: The full path to the created plan file
- `need_e2e_tests`: Set to `true` if the feature includes UI components, user interactions, or frontend changes that require e2e testing. Set to `false` if the feature is backend-only or doesn't require e2e tests.
- `footprint_path`: The full path to the created footprint file
- `state_path`: The full path to the state file

**Determining need_e2e_tests:**
- Set to `true` if the feature:
  - Adds or modifies UI components
  - Changes user interactions
  - Modifies frontend code (viz/frontend/, .tsx, .jsx, .css files)
  - Requires user-facing validation
- Set to `false` if the feature:
  - Is backend-only (services/core, services/auth, services/gateway, etc.)
  - Doesn't affect UI or user interactions
  - Only modifies configuration, documentation, or backend logic
