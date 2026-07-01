---
name: test
description: Comprehensive Testing skill for validating implementation quality through unit tests, integration tests, and E2E tests. Use this skill after implementation is complete to run all tests, fix bugs, generate test reports with evidence (screenshots/videos for UI), and ensure code quality before documentation. Creates test checklists, test reports, and fixes all identified bugs.
---

# Comprehensive Testing

Execute comprehensive testing after implementation to validate code quality, fix bugs, and generate test reports with evidence.

## Variables

branch_name: Current git branch name (e.g., "feature/new-UI")
plan_file_path: $1 (optional) - Path to the plan file that was implemented

## Instructions

- IMPORTANT: This skill runs AFTER implementation and BEFORE document-results.
- Run all relevant tests (unit, integration, E2E) based on what was implemented.
- Fix any bugs found during testing.
- Generate comprehensive test reports with evidence.
- For UI changes, capture screenshots and/or video recordings.
- Create test checklists and mark items as passed/failed.
- Only proceed to documentation when all tests pass.

### Workflow Position

```
... → implement → TEST → document-results → review → ...
```

### Testing Workflow

1. **Read State File**: Check `agentic/{branch-name}/state.json` for implementation details
2. **Read Implementation Footprints**: Load all `foot-impl-*.md` files to understand changes
3. **Create Test Checklist**: Generate checklist based on acceptance criteria
4. **Run Unit Tests**: Execute unit tests for changed components
5. **Run Integration Tests**: Execute integration tests for affected modules
6. **Run E2E Tests (MANDATORY for UI changes)**: Execute Playwright spec tests
   - **ALWAYS** run Playwright specs when UI was changed. E2E is NOT optional.
   - **Extend existing spec files** in `viz/frontend/` (Vitest) — do NOT create new MCP markdown scripts.
   - Run with: `cd viz/frontend && npm run test`
   - After execution, **read** `evidence/test-results/results.json` to parse pass/fail counts.
   - If failures: read trace files, view failure screenshots, analyze errors.
   - Read HTML report summary from `evidence/playwright-report/index.html`.
7. **Fix Bugs**: Resolve any test failures
8. **Capture Evidence**: Screenshots/videos captured automatically by Playwright config
9. **Commit Updated Snapshots**: If `toHaveScreenshot()` baselines changed intentionally, commit the new snapshots in `e2e/__snapshots__/`.
10. **Generate Test Report**: Comprehensive report with all results
11. **Update Bead**: If `bead_id` exists in `state.json`, run `bd update $bead_id --notes="Tests: Unit=$unit_pass/$unit_total, E2E=$e2e_pass/$e2e_total. Bugs fixed: $count"`
12. **Update State**: Track testing status in state.json

---

## Test Types

### Lint & TypeScript Validation (MANDATORY)

**ALWAYS run lint before proceeding.**

```bash
# Backend (pytest + ruff)
pytest
ruff check .  # if installed

# Frontend
cd viz/frontend && npm run lint
cd viz/frontend && npx tsc --noEmit  # typecheck
```

**BLOCKING**: If lint/typecheck fails, you MUST fix the errors before proceeding. Do not skip lint failures.

Common issues to watch for:

- **Unused imports**
- Undefined variables
- TypeScript type errors (tsc --noEmit)
- Missing or incorrect type annotations

**Evidence from footprint analysis:**

- All 5 bugs in foot-test-bugfix.md were lint errors (unused imports)
- These could have been caught earlier with automated lint check

### Unit Tests

Run unit tests for changed files:

```bash
# Backend (pytest)
pytest viz/backend/tests/ agents/tests/ -v -k "{changed_module}"

# Frontend (Vitest)
cd viz/frontend && npm run test
```

### Integration Tests

Run integration tests for affected modules:

```bash
# API Integration tests
cd services/{service-name}
bun test:e2e -- --testPathPattern="{module}.integration"

# Database integration
bun test -- --testPathPattern="integration"
```

### E2E Tests (Playwright) — MANDATORY for UI Changes

**ALWAYS** run Playwright specs when UI was changed. Extend existing spec files, do NOT create MCP markdown scripts.

#### Available Spec Files

| Spec File                      | Coverage                                                              | Page Objects Used              |
| ------------------------------ | --------------------------------------------------------------------- | ------------------------------ |
| `e2e/tasks.spec.ts`            | Task CRUD, board, detail, comments, status, labels                    | TasksBoardPage, TaskDetailPage |
| `e2e/kanban-board.spec.ts`     | Kanban columns, create, navigate, status change                       | TasksBoardPage, TaskDetailPage |
| `e2e/task-page-linear.spec.ts` | Three-column layout, keyboard shortcuts, command palette, attachments | TaskDetailPage                 |
| `e2e/leads.spec.ts`            | CRM leads, tabs, form, no-title check                                 | LeadsPage                      |
| `e2e/theme.spec.ts`            | Theme toggle, persistence, visual                                     | -                              |
| `e2e/chat.spec.ts`             | Chat tabs, conversations, messaging                                   | NavigationPage                 |

#### Page Objects (`e2e/pages/`)

| Page Object      | Key Methods                                                                                             |
| ---------------- | ------------------------------------------------------------------------------------------------------- |
| `LoginPage`      | `goto()`, `login(email, password)`                                                                      |
| `NavigationPage` | `navigateTo(name)`, `navigateToTasks()`, `navigateToLeads()`                                            |
| `TasksBoardPage` | `goto()`, `expectBoardVisible()`, `expectColumnsVisible()`, `clickTask()`                               |
| `TaskDetailPage` | `goto(id)`, `goBack()`, `openCommandPalette()`, `changeStatus()`, `addComment()`, `toggleLeftSidebar()` |
| `LeadsPage`      | `goto()`, `expectTabsVisible()`, `switchTab()`, `clickAddLead()`, `expectNoTitle()`                     |

#### Running E2E Tests

```bash
cd services/app-client

# Run all E2E tests
bun run e2e

# Run specific spec file
bun run e2e -- e2e/tasks.spec.ts

# Run with UI mode for debugging
bun run e2e:ui

# View HTML report after run
bun run e2e:report
```

#### Parsing Results Programmatically

After running `bun run e2e`, read `evidence/test-results/results.json`:

```bash
# Parse pass/fail counts
cat evidence/test-results/results.json | jq '.stats'
```

The JSON report contains:

- `stats.expected` — passed test count
- `stats.unexpected` — failed test count
- `stats.flaky` — flaky test count
- `suites[].specs[].tests[].results[]` — individual test results with screenshots, video paths

#### Visual Regression with `toHaveScreenshot()`

Add visual regression checkpoints to specs:

```typescript
await expect(page).toHaveScreenshot("feature-name.png", {
  maxDiffPixelRatio: 0.01,
  mask: [page.locator("[data-testid='timestamp']")],
});
```

- First run creates baseline snapshots in `e2e/__snapshots__/`
- Subsequent runs compare against baselines
- Update baselines: `bun run e2e -- --update-snapshots`
- Commit updated snapshots when changes are intentional

#### Adding Tests for New Features

When adding tests for a new feature:

1. **Check if a Page Object exists** for the feature area in `e2e/pages/`
2. **Create a new Page Object** if needed, following the pattern in existing files
3. **Add tests to an existing spec** if the feature extends existing functionality
4. **Create a new spec file** only if it's a completely new feature area
5. **Use shared fixtures** from `e2e/fixtures/test-data.ts` for test data
6. **Add `toHaveScreenshot()`** for key visual states

---

## Test Evidence Capture

### For UI Testing

Always capture visual evidence when testing UI changes:

#### Screenshots

```typescript
// In Playwright tests
await page.screenshot({
  path: `agentic/{branch-name}/evidence/screenshots/{feature}-{step}.png`,
  fullPage: true,
});

// Element-specific screenshot
await locator.screenshot({
  path: `agentic/{branch-name}/evidence/screenshots/{feature}-{element}.png`,
});
```

#### Video Recording

Configure in playwright.config.ts:

```typescript
use: {
  video: 'on', // or 'retain-on-failure'
}
```

Videos saved to: `agentic/{branch-name}/evidence/videos/`

#### Manual Evidence Capture

For manual verification, use these commands:

```bash
# Create evidence directories
mkdir -p agentic/{branch-name}/evidence/screenshots
mkdir -p agentic/{branch-name}/evidence/videos

# Take screenshot with browser dev tools or:
# - macOS: Cmd+Shift+4
# - Linux: gnome-screenshot or flameshot
# Save to: agentic/{branch-name}/evidence/screenshots/
```

---

## Screenshot Verification (REQUIRED)

**CRITICAL**: After capturing any screenshot, you MUST verify its contents by viewing it.

### Why Screenshot Verification

Screenshots serve as visual evidence that UI expectations are met. Without viewing the screenshots, you cannot confirm:

- The correct UI elements are visible
- The layout matches the expected design
- Text content is displayed correctly
- Interactive states (hover, focus, disabled) are correct
- No visual regressions exist

### Screenshot Verification Workflow

After EVERY screenshot capture:

1. **View the Screenshot**: Use the `Read` tool to view the image file

   ```
   Read the screenshot file at: agentic/{branch-name}/evidence/screenshots/{test-name}/{screenshot-name}.png
   ```

2. **Verify Against Expectations**: Compare what you see with what was expected:
   - Does the layout match the expected structure?
   - Are all required UI elements visible?
   - Is the text/content correct?
   - Are colors, spacing, and styling correct?
   - Are interactive elements in the expected state?

3. **Document Verification Result**: In the test report, for each screenshot include:
   - What was expected to be shown
   - What was actually captured
   - PASS/FAIL determination
   - Any discrepancies noted

### Screenshot Verification Template

For each screenshot, document:

```markdown
### Screenshot: {screenshot-name}.png

**Expected**: {describe what should be visible in the screenshot}

- Element A should be visible at top
- Element B should show text "XYZ"
- Button should be in enabled state

**Actual**: {describe what was captured - AFTER viewing the screenshot}

- Element A is visible at top ✓
- Element B shows text "XYZ" ✓
- Button is enabled ✓

**Verification**: PASS/FAIL

**Notes**: {any observations or discrepancies}
```

### Integration with E2E Tests

When running E2E tests that capture screenshots:

1. Take screenshot at designated step
2. **Immediately view the screenshot using Read tool**
3. Verify the screenshot shows expected content
4. If verification fails, mark the test step as FAILED
5. Continue to next step or halt test based on severity

### Example Screenshot Verification

```
Step 5: Capture login form screenshot
→ Screenshot saved: agentic/feature/auth/evidence/screenshots/login/01_login_form.png

→ VIEW SCREENSHOT (using Read tool to view the image)

Expected:
- Login form with email and password fields
- "Sign In" button visible and enabled
- Company logo at top

Actual (after viewing screenshot):
- Login form visible with email/password fields ✓
- "Sign In" button visible and enabled ✓
- Company logo displayed at top ✓

Verification: PASS
```

### Screenshot Verification in Test Reports

Include a dedicated section in test reports:

```markdown
## Screenshot Verification Results

| Screenshot           | Expected                       | Actual                          | Status |
| -------------------- | ------------------------------ | ------------------------------- | ------ |
| 01_initial_state.png | Dashboard with 3-column layout | Dashboard shows 3-column layout | PASS   |
| 02_after_click.png   | Modal dialog open              | Modal visible with form         | PASS   |
| 03_error_state.png   | Error message in red           | Error message displayed         | PASS   |

All {N} screenshots verified successfully.
```

---

## Relevant Files

Focus on the following files:

- `agentic/{branch-name}/state.json` - Contains workflow state
- `agentic/{branch-name}/footprints/foot-impl-*.md` - Implementation footprints
- `specs/sdlc_planner-*.md` - Plan files with acceptance criteria
- `specs/patch/patch-*.md` - Patch plan files
- `viz/backend/tests/**/*.py` - Backend tests (pytest)
- `agents/tests/**/*.py` - Agent tests (pytest)
- `viz/frontend/src/**/*.test.ts` - Frontend unit tests (Vitest)

### Test Configuration Files

- `viz/frontend/vitest.config.ts` - Vitest configuration
- `pytest.ini` or `pyproject.toml` - pytest configuration

---

## Footprint Strategy

Create footprints at these key points:

- **foot-test-start.md**: When starting testing phase
- **foot-test-unit.md**: After unit tests complete
- **foot-test-integration.md**: After integration tests complete
- **foot-test-e2e.md**: After E2E tests complete (with evidence)
- **foot-test-bugfix.md**: After fixing discovered bugs (if any)
- **foot-test-complete.md**: When all tests pass

---

## Beads Test Tracking

If `bead_id` exists in `state.json`, update the bead after testing completes:

```bash
bd update $bead_id --notes="Tests: Unit=$unit_pass/$unit_total, E2E=$e2e_pass/$e2e_total. Bugs fixed: $count"
```

Read `bead_id` from `agentic/{branch-name}/state.json`. If `bead_id` is null or missing, skip bead updates silently.

---

## Fleet Test Label Auto-Apply

After all tests pass, auto-apply test routing labels to the bead so the fleet orchestrator can route it to the correct tester environment. This uses the same label rules from the `/test-labels` skill.

### Quick Label Application

```bash
# Get changed files
changed_files=$(git diff --name-only origin/master...HEAD)

# Apply labels based on path patterns
echo "$changed_files" | grep -qE "^(viz/backend|agents|mcp-servers)" && (bd label add $bead_id test:backend)
echo "$changed_files" | grep -q "^viz/frontend/" && (bd label add $bead_id test:ui)
```

### Label Routing Summary

| Label            | Fleet Tester | What Runs                                           |
| ---------------- | ------------ | --------------------------------------------------- |
| `test:backend`   | `tester-be`  | `bun test` (5 min, 1 retry)                         |
| `test:typecheck` | `tester-be`  | `bun run typecheck` (2 min)                         |
| `test:migration` | `tester-be`  | `prisma migrate deploy` + `prisma generate` (3 min) |
| `test:coupled:*` | `tester-be`  | Falls back to `test:backend` suite                  |
| `test:ui`        | `tester-ui`  | `bunx playwright test` (10 min, 2 retries)          |

See `/candidate-dev` for the full dev-candidate pipeline that runs after PR creation.

---

## Test Checklist

Create at: `agentic/{branch-name}/test-checklist.md`

```markdown
# Test Checklist for {plan-title}

**Date**: {ISO 8601 timestamp}
**Branch**: {branch-name}
**Plan File**: {plan file path}

## Pre-Test Checks

- [ ] All implementation phases completed
- [ ] Code compiles without errors
- [ ] No TypeScript/ESLint errors
- [ ] Dependencies installed correctly

## Unit Tests

- [ ] New unit tests written for new code
- [ ] Existing unit tests updated for modified code
- [ ] All unit tests pass
- [ ] Coverage meets threshold (>80%)

| Test File       | Status    | Coverage |
| --------------- | --------- | -------- |
| {file1}.spec.ts | PASS/FAIL | XX%      |
| {file2}.spec.ts | PASS/FAIL | XX%      |

## Integration Tests

- [ ] API endpoints tested
- [ ] Database operations tested
- [ ] Service interactions tested
- [ ] All integration tests pass

| Test File                   | Status    | Notes   |
| --------------------------- | --------- | ------- |
| {file1}.integration.spec.ts | PASS/FAIL | {notes} |

## E2E Tests (if UI changes)

- [ ] User flows tested
- [ ] Cross-browser compatibility verified
- [ ] Responsive design tested
- [ ] Accessibility checked
- [ ] Screenshots captured
- [ ] Video recorded (if needed)

| Test Scenario | Status    | Evidence                                        |
| ------------- | --------- | ----------------------------------------------- |
| {scenario1}   | PASS/FAIL | [screenshot](./evidence/screenshots/{name}.png) |
| {scenario2}   | PASS/FAIL | [video](./evidence/videos/{name}.webm)          |

## Acceptance Criteria Validation

{List each acceptance criteria from plan with pass/fail status}

- [ ] AC1: {description} - PASS/FAIL
- [ ] AC2: {description} - PASS/FAIL

## Bug Fixes Applied

| Bug ID  | Description   | Fix Applied | Verified |
| ------- | ------------- | ----------- | -------- |
| BUG-001 | {description} | {fix}       | YES/NO   |

## Manual Verification

- [ ] Feature works as expected
- [ ] No console errors
- [ ] Performance acceptable
- [ ] No visual regressions

## Sign-off

- [ ] All tests pass
- [ ] All bugs fixed
- [ ] Ready for documentation
```

---

## Test Report Format

Create at: `agentic/{branch-name}/test-report.md`

```markdown
# Test Report: {plan-title}

**Generated**: {ISO 8601 timestamp}
**Branch**: {branch-name}
**Implementation Status**: Completed
**Overall Test Status**: PASS/FAIL

---

## Executive Summary

{Brief summary of testing outcomes - 2-3 sentences}

### Quick Stats

| Metric      | Value         |
| ----------- | ------------- |
| Total Tests | {count}       |
| Passed      | {count}       |
| Failed      | {count}       |
| Skipped     | {count}       |
| Coverage    | {percentage}% |
| Bugs Found  | {count}       |
| Bugs Fixed  | {count}       |

---

## Test Execution Details

### Unit Tests

**Command**: `bun test`
**Duration**: {duration}
**Result**: PASS/FAIL
```

{test output - truncated if too long}

```

#### Coverage Report

| File/Module | Statements | Branches | Functions | Lines |
|-------------|------------|----------|-----------|-------|
| {module1} | XX% | XX% | XX% | XX% |
| {module2} | XX% | XX% | XX% | XX% |
| **Total** | XX% | XX% | XX% | XX% |

### Integration Tests

**Command**: `bun test:e2e`
**Duration**: {duration}
**Result**: PASS/FAIL

```

{test output}

```

### E2E Tests

**Command**: `bun e2e`
**Duration**: {duration}
**Result**: PASS/FAIL
**Browser**: Chromium/Firefox/WebKit

```

{test output}

```

---

## Visual Evidence

### Screenshots

| Screenshot | Description | Status |
|------------|-------------|--------|
| ![{name}](./evidence/screenshots/{name}.png) | {description} | {status} |

### Video Recordings

| Recording | Description | Duration |
|-----------|-------------|----------|
| [{name}](./evidence/videos/{name}.webm) | {description} | {duration} |

---

## Bugs Discovered and Fixed

### BUG-001: {Bug Title}

**Severity**: High/Medium/Low
**Component**: {component name}
**Description**: {detailed description}
**Root Cause**: {root cause analysis}
**Fix Applied**: {description of fix}
**Files Changed**:
- `{file1.ts}` - {change description}
- `{file2.ts}` - {change description}

**Verification**: {how fix was verified}

---

## Acceptance Criteria Results

| # | Criteria | Status | Evidence |
|---|----------|--------|----------|
| 1 | {AC description} | PASS | {test name or screenshot} |
| 2 | {AC description} | PASS | {test name or screenshot} |
| 3 | {AC description} | FAIL | {reason} |

---

## Performance Metrics (if applicable)

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Load Time | {ms} | {ms} | {diff} |
| Bundle Size | {kb} | {kb} | {diff} |
| Memory Usage | {mb} | {mb} | {diff} |

---

## Recommendations

### Improvements Identified
{List any code quality or performance improvements discovered during testing}

### Technical Debt
{List any technical debt identified}

### Future Test Coverage
{Areas that could benefit from additional testing}

---

## Conclusion

{Final assessment of implementation quality and readiness for documentation}

**Ready for Documentation**: YES/NO
**Next Step**: /document-results
```

---

## Footprint Formats

### Start Footprint

Create at: `agentic/{branch-name}/footprints/foot-test-start.md`

```markdown
# Testing Start Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: Testing Start

## Testing Scope

### Implementation Summary

{brief summary of what was implemented}

### Components to Test

{list of components/modules requiring testing}

### Test Types Required

- [ ] Unit Tests
- [ ] Integration Tests
- [ ] E2E Tests
- [ ] Manual Verification

### Acceptance Criteria to Validate

{list from plan file}

## Testing Strategy

### Test Execution Order

1. Lint and type check
2. Unit tests
3. Integration tests
4. E2E tests (if UI changes)
5. Manual verification

### Evidence to Capture

{list of screenshots/videos needed}

### Test Environment

- Node Version: {version}
- Test Framework: Jest/Vitest/Playwright
- Browser: Chromium

## Next Steps

**Current Step**: Lint and type check
**Next Milestone**: Unit tests complete
```

### E2E Test Footprint

Create at: `agentic/{branch-name}/footprints/foot-test-e2e.md`

```markdown
# E2E Testing Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: E2E Testing

## E2E Test Execution

### Tests Run

{list of E2E test files executed}

### Browser Coverage

- [x] Chromium
- [ ] Firefox
- [ ] WebKit

### Results Summary

| Test    | Status | Duration |
| ------- | ------ | -------- |
| {test1} | PASS   | {time}   |
| {test2} | FAIL   | {time}   |

## Visual Evidence

### Screenshots Captured

| Screenshot                                               | Step               | Status |
| -------------------------------------------------------- | ------------------ | ------ |
| [login-page.png](../evidence/screenshots/login-page.png) | Login form display | PASS   |
| [dashboard.png](../evidence/screenshots/dashboard.png)   | After login        | PASS   |

### Screenshot Verification Results

**IMPORTANT**: Each screenshot was viewed and verified against expectations.

| #   | Screenshot           | Expected                            | Actual (Verified)                                | Status  |
| --- | -------------------- | ----------------------------------- | ------------------------------------------------ | ------- |
| 1   | 01_initial_state.png | Login page with email/password form | Login form visible with fields and submit button | ✅ PASS |
| 2   | 02_dashboard.png     | Dashboard with navigation sidebar   | Dashboard layout with 3-column structure         | ✅ PASS |
| 3   | 03_task_detail.png   | Task detail panel with metadata     | Task #123 showing title, description, status     | ✅ PASS |

**Verification Summary**: {N}/{N} screenshots verified successfully

### Detailed Screenshot Verifications

#### Screenshot 1: 01_initial_state.png

- **Expected**: {detailed description of expected state}
- **Actual**: {what was seen after viewing the screenshot}
- **Verification**: PASS/FAIL
- **Notes**: {any observations}

#### Screenshot 2: 02_dashboard.png

- **Expected**: {detailed description}
- **Actual**: {what was seen}
- **Verification**: PASS/FAIL
- **Notes**: {observations}

### Videos Recorded

| Video                                               | Scenario              | Duration   |
| --------------------------------------------------- | --------------------- | ---------- |
| [user-flow.webm](../evidence/videos/user-flow.webm) | Complete user journey | {duration} |

## Issues Found

### Issue 1

- **Test**: {test name}
- **Error**: {error message}
- **Screenshot**: {link}
- **Screenshot Verification**: {what was seen in the screenshot that indicated failure}

## Bug Fixes Required

{list of bugs to fix before proceeding}

## Next Steps

**Current Step**: Fix E2E failures (if any)
**Next Milestone**: All tests pass
```

### Bug Fix Footprint

Create at: `agentic/{branch-name}/footprints/foot-test-bugfix.md`

```markdown
# Bug Fix Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: Bug Fixes During Testing

## Bugs Fixed

### BUG-001: {Title}

**Discovered In**: {test type - unit/integration/e2e}
**Test File**: {test file path}
**Error Message**:
```

{error message}

```

**Root Cause**: {explanation}

**Fix Applied**:
- File: `{file path}`
- Changes: {description of changes}

**Verification**:
- Test now passes: YES
- Related tests affected: {list or NONE}

### BUG-002: {Title}
{repeat structure}

## Files Modified

| File | Changes | Reason |
|------|---------|--------|
| {file1.ts} | {changes} | {bug reference} |
| {file2.ts} | {changes} | {bug reference} |

## Re-Test Results

After fixes, all tests re-run:

- Unit Tests: PASS
- Integration Tests: PASS
- E2E Tests: PASS

## Commit Information

**Commit Message**: fix({scope}): {description of fixes}
**Files Committed**: {count}

## Next Steps

**Current Step**: Generate final test report
**Next Milestone**: Testing complete
```

### Completion Footprint

Create at: `agentic/{branch-name}/footprints/foot-test-complete.md`

```markdown
# Testing Complete Footprint

**Date**: {ISO 8601 timestamp}
**Plan File**: {path to plan file}
**Type**: Testing Complete

## Testing Summary

### Overall Status: PASS

### Test Execution Summary

| Test Type   | Total | Passed | Failed | Skipped |
| ----------- | ----- | ------ | ------ | ------- |
| Unit        | {n}   | {n}    | 0      | {n}     |
| Integration | {n}   | {n}    | 0      | {n}     |
| E2E         | {n}   | {n}    | 0      | {n}     |
| **Total**   | {n}   | {n}    | 0      | {n}     |

### Coverage Summary

| Metric     | Value | Threshold | Status |
| ---------- | ----- | --------- | ------ |
| Statements | XX%   | 80%       | PASS   |
| Branches   | XX%   | 75%       | PASS   |
| Functions  | XX%   | 80%       | PASS   |
| Lines      | XX%   | 80%       | PASS   |

## Bugs Fixed

| Bug ID  | Description   | Status |
| ------- | ------------- | ------ |
| BUG-001 | {description} | FIXED  |
| BUG-002 | {description} | FIXED  |

## Evidence Collected

### Screenshots

- {count} screenshots captured
- Location: `agentic/{branch-name}/evidence/screenshots/`

### Videos

- {count} videos recorded
- Location: `agentic/{branch-name}/evidence/videos/`

## Artifacts Generated

- `agentic/{branch-name}/test-checklist.md` - Test checklist
- `agentic/{branch-name}/test-report.md` - Full test report
- `agentic/{branch-name}/evidence/` - Visual evidence

## Acceptance Criteria Status

All acceptance criteria validated: YES

{list each AC with PASS status}

## Quality Assessment

### Code Quality

- No TypeScript errors
- No ESLint warnings
- All tests pass
- Coverage thresholds met

### Implementation Quality

- Feature works as expected
- No regressions introduced
- Performance acceptable

## Commits Made

| Commit | Message                 |
| ------ | ----------------------- |
| {hash} | fix({scope}): {message} |

## Next Steps

**Recommended Action**: /document-results
**Ready for Documentation**: YES
```

---

## State File Update

Update the state file at: `agentic/{branch-name}/state.json`

### Workflow Step Tracking

| Testing Phase    | prev_step | next_step        |
| ---------------- | --------- | ---------------- |
| Starting testing | implement | document-results |
| During testing   | implement | document-results |
| Testing complete | test      | document-results |

### During Testing

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-test-unit.md",
  "latest_command": "test",
  "plan_file_path": "{plan file path}",
  "implementation_status": "completed",
  "testing_status": "in_progress",
  "prev_step": "implement",
  "next_step": "document-results",
  "testing_progress": {
    "current_phase": "unit|integration|e2e|bugfix",
    "unit_tests": {
      "status": "passed|failed|pending",
      "total": 0,
      "passed": 0,
      "failed": 0,
      "coverage": 0
    },
    "integration_tests": {
      "status": "passed|failed|pending",
      "total": 0,
      "passed": 0,
      "failed": 0
    },
    "e2e_tests": {
      "status": "passed|failed|pending|skipped",
      "total": 0,
      "passed": 0,
      "failed": 0
    },
    "bugs_found": 0,
    "bugs_fixed": 0,
    "evidence_captured": {
      "screenshots": 0,
      "videos": 0
    }
  },
  "next_command_metadata": {
    "command": "/test",
    "category": "testing",
    "confidence": "high",
    "reasoning": "Continuing testing",
    "required_context": "{current phase info}"
  },
  "next_command": "/test",
  "testing_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}"
}
```

### After Testing Complete

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-test-complete.md",
  "latest_command": "test",
  "plan_file_path": "{plan file path}",
  "implementation_status": "completed",
  "testing_status": "completed",
  "prev_step": "test",
  "next_step": "document-results",
  "testing_progress": {
    "current_phase": "complete",
    "unit_tests": {
      "status": "passed",
      "total": "{count}",
      "passed": "{count}",
      "failed": 0,
      "coverage": "{percentage}"
    },
    "integration_tests": {
      "status": "passed",
      "total": "{count}",
      "passed": "{count}",
      "failed": 0
    },
    "e2e_tests": {
      "status": "passed|skipped",
      "total": "{count}",
      "passed": "{count}",
      "failed": 0
    },
    "bugs_found": "{count}",
    "bugs_fixed": "{count}",
    "evidence_captured": {
      "screenshots": "{count}",
      "videos": "{count}"
    }
  },
  "test_artifacts": {
    "checklist": "agentic/{branch-name}/test-checklist.md",
    "report": "agentic/{branch-name}/test-report.md",
    "evidence_dir": "agentic/{branch-name}/evidence/"
  },
  "next_command_metadata": {
    "command": "/document-results",
    "category": "documentation",
    "confidence": "high",
    "reasoning": "All tests passed, ready for documentation",
    "required_context": "Testing complete"
  },
  "next_command": "/document-results",
  "testing_completion_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "all_tests_passed": true
}
```

---

## Implementation Steps

### Step 1: Initialize Testing

1. Read state file from `agentic/{branch-name}/state.json`
2. Verify implementation is complete (`implementation_status: "completed"`)
3. Read implementation footprints to understand what was changed
4. Create test checklist based on plan's acceptance criteria
5. Create `foot-test-start.md` footprint
6. Create evidence directories

```bash
mkdir -p agentic/{branch-name}/evidence/screenshots
mkdir -p agentic/{branch-name}/evidence/videos

git add agentic/{branch-name}/footprints/foot-test-start.md agentic/{branch-name}/test-checklist.md
git commit -m "test({scope}): start testing phase

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

### Step 2: Run Unit Tests

1. Identify changed modules from implementation footprints
2. Run unit tests for those modules
3. Capture test results and coverage
4. Create `foot-test-unit.md` if all pass
5. If failures, proceed to bug fixing

```bash
# Run unit tests with coverage
cd services/{service-name}
bun test:cov

# Check for errors
bun lint
bun typecheck
```

### Step 3: Run Integration Tests

1. Identify affected integration tests
2. Execute integration tests
3. Capture results
4. Create `foot-test-integration.md` if all pass

```bash
cd services/{service-name}
bun test:e2e -- --testPathPattern="integration"
```

### Step 4: Run E2E Tests (MANDATORY for UI changes)

1. Identify which Playwright spec files cover the changed features
2. Run Playwright tests (extend existing specs if needed, do NOT create MCP scripts)
3. Parse `evidence/test-results/results.json` for pass/fail counts
4. If failures: view screenshots, read traces, fix issues
5. If visual regression baselines changed intentionally, update snapshots
6. Create `foot-test-e2e.md` with evidence links

```bash
cd services/app-client

# Run all E2E tests
bun run e2e

# Parse results
cat evidence/test-results/results.json | jq '.stats'

# If visual snapshots need updating
bun run e2e -- --update-snapshots

# View HTML report
bun run e2e:report
```

### Step 5: Fix Bugs (if any failures)

1. Analyze test failures
2. Identify root cause
3. Apply fixes
4. Re-run failed tests
5. Create `foot-test-bugfix.md`
6. Commit fixes

```bash
git add {fixed-files}
git commit -m "fix({scope}): {bug description}

Fixes test failures discovered during testing phase.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

### Step 6: Generate Test Report

1. Compile all test results
2. Include evidence (screenshots/videos)
3. Document bugs found and fixed
4. Validate acceptance criteria
5. Create comprehensive test report

### Step 7: Finalize

1. Verify all tests pass
2. Verify all bugs fixed
3. Complete test checklist
4. Create `foot-test-complete.md`
5. Update state file
6. Commit final test artifacts

```bash
git add agentic/{branch-name}/footprints/foot-test-complete.md \
        agentic/{branch-name}/test-report.md \
        agentic/{branch-name}/test-checklist.md \
        agentic/{branch-name}/state.json \
        agentic/{branch-name}/evidence/

git commit -m "test({scope}): complete testing phase

All tests passed:
- Unit tests: {count} passed
- Integration tests: {count} passed
- E2E tests: {count} passed
- Bugs fixed: {count}

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Required Actions

During testing, you MUST:

1. **Verify Implementation**: Ensure implementation is complete before testing
2. **Create Test Checklist**: Document all items to test
3. **Run All Test Types**: Unit, integration, and E2E as appropriate
4. **Capture Evidence**: Screenshots/videos for UI changes
5. **VERIFY SCREENSHOTS**: After capturing any screenshot, you MUST:
   - Use the Read tool to view the screenshot image
   - Compare what you see against what was expected
   - Document the verification result (expected vs actual)
   - Mark PASS/FAIL for each screenshot
   - If verification fails, mark the test step as failed
6. **Fix All Bugs**: Resolve every test failure before proceeding
7. **Create Footprints**: Document progress at each milestone
8. **Commit After Milestones**: Commit after each footprint
9. **Generate Test Report**: Comprehensive report with evidence AND screenshot verifications
10. **Update State**: Keep state.json updated throughout

---

## Report

After completing testing, return a JSON object:

```json
{
  "plan_file_path": "{path to plan file}",
  "testing_status": "completed|failed",
  "unit_tests": {
    "total": "{count}",
    "passed": "{count}",
    "failed": 0,
    "coverage": "{percentage}%"
  },
  "integration_tests": {
    "total": "{count}",
    "passed": "{count}",
    "failed": 0
  },
  "e2e_tests": {
    "total": "{count}",
    "passed": "{count}",
    "failed": 0,
    "skipped": "{count}"
  },
  "bugs_found": "{count}",
  "bugs_fixed": "{count}",
  "evidence": {
    "screenshots": "{count}",
    "videos": "{count}",
    "location": "agentic/{branch-name}/evidence/"
  },
  "screenshot_verifications": {
    "total": "{count}",
    "passed": "{count}",
    "failed": 0,
    "results": [
      {
        "screenshot": "{filename}.png",
        "expected": "{description of expected content}",
        "actual": "{description of what was seen after viewing}",
        "status": "passed|failed"
      }
    ]
  },
  "artifacts": {
    "checklist": "agentic/{branch-name}/test-checklist.md",
    "report": "agentic/{branch-name}/test-report.md"
  },
  "footprints": [
    "agentic/{branch-name}/footprints/foot-test-start.md",
    "agentic/{branch-name}/footprints/foot-test-unit.md",
    "agentic/{branch-name}/footprints/foot-test-integration.md",
    "agentic/{branch-name}/footprints/foot-test-e2e.md",
    "agentic/{branch-name}/footprints/foot-test-bugfix.md",
    "agentic/{branch-name}/footprints/foot-test-complete.md"
  ],
  "all_tests_passed": true,
  "all_screenshots_verified": true,
  "ready_for_documentation": true,
  "commits_made": ["{list of commit hashes}"],
  "state_path": "agentic/{branch-name}/state.json"
}
```

---

## Error Handling

If testing fails at any point:

1. Create footprint documenting the failure
2. Update state file with `testing_status: "failed"`
3. Include error details and failed test output
4. Do NOT proceed to documentation until all tests pass
5. Report the failure in the final JSON response

---

## Resuming Testing

If resuming partial testing:

1. Read state file to get current progress
2. Find the last completed test phase
3. Continue from the next incomplete phase
4. Create appropriate footprint when phase completes
5. Continue normal workflow

---

## Data Integrity Test Pattern (EMM)

**Purpose**: Validate store/repository data integrity

**Pattern** (pytest):

```python
# viz/backend/tests/test_store.py
def test_store_integrity(store):
    """Verify store returns consistent data."""
    items = store.list_items()
    assert all(item.id for item in items)
```

**When to Run**:

- After store/repository changes
- Before committing

---

## Playwright Page Object Model Testing

**Page Object**:

```typescript
export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/login");
  }

  async login(email: string, password: string) {
    await this.page.getByLabel("Email").fill(email);
    await this.page.getByLabel("Password").fill(password);
    await this.page.getByRole("button", { name: "Log in" }).click();
  }

  async expectError(message: string) {
    await expect(this.page.getByText(message)).toBeVisible();
  }
}
```

**Test File**:

```typescript
test.describe("Login", () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
  });

  test("successful login", async ({ page }) => {
    await loginPage.login("user@example.com", "password");
    await expect(page).toHaveURL("/dashboard");
  });

  test("invalid credentials", async () => {
    await loginPage.login("user@example.com", "wrong");
    await loginPage.expectError("Invalid credentials");
  });
});
```

**Benefits**:

- Reusable page interactions
- Easier maintenance when UI changes
- More readable tests
- Type-safe locators

---

## Eden Treaty Test Client Pattern

**Setup**:

```typescript
import { treaty } from "@elysiajs/eden";
import { app } from "../src/index"; // Elysia app

const api = treaty(app);

// Alternative: Use app.handle() directly
const response = await app.handle(
  new Request("http://localhost/api/crm/leads"),
);
```

**Test Pattern**:

```typescript
describe("Lead API", () => {
  beforeEach(async () => {
    await prisma.lead.deleteMany();
  });

  test("create lead", async () => {
    const { data, error } = await api.api.crm.leads.post({
      name: "Test Lead",
      sourceId: "source-123",
      companyId: "company-123",
    });

    expect(error).toBeNull();
    expect(data.name).toBe("Test Lead");
  });

  test("list leads", async () => {
    // Arrange
    await prisma.lead.create({
      data: { name: "Lead 1", sourceId: "...", companyId: "..." },
    });

    // Act
    const { data } = await api.api.crm.leads.get({
      query: { limit: 10 },
    });

    // Assert
    expect(data.data).toHaveLength(1);
    expect(data.total).toBe(1);
  });
});
```

**Benefits**:

- Type-safe test client
- No HTTP server needed
- Matches production client usage
- Fast test execution

---

## Pre-Existing Error Isolation

**Problem**: Tests fail due to pre-existing issues unrelated to current changes

**Strategy**:

1. **Identify pre-existing errors**:

   ```bash
   git stash  # Stash current changes
   # Note which tests fail
   git stash pop
   ```

2. **Document pre-existing failures**:

   ```markdown
   ## Pre-Existing Test Failures

   The following tests fail on main branch before changes:
   - `user.test.ts:45` - Type error in old code
   - `lead.test.ts:120` - Missing database constraint
   ```

3. **Skip or isolate**:

   ```typescript
   // Option 1: Skip pre-existing failures
   test.skip('old broken test', async () => { ... })

   // Option 2: Flag but don't block
   test('old broken test', async () => {
     try {
       // test logic
     } catch (error) {
       console.warn('Pre-existing failure:', error)
     }
   })
   ```

4. **Focus on new tests**:
   ```bash
   # Run only new test files
   bun test tests/new-feature.test.ts
   ```

**When to Use**:

- Large codebase with tech debt
- Migration in progress
- Tight deadlines (fix later)

**Important**: Always file issues for pre-existing failures and link in PR

**Evidence from footprint analysis:**

- Pre-existing error isolation saved hours of debugging during migration

---

## Related Skills

- **bdd-test** - Behavior-Driven Development testing with jest-cucumber
- **e2e-api-test** - API testing patterns with Jest and Supertest
- **playwright** - Browser E2E testing patterns and best practices
- **unit-test** - Unit testing with optimal coverage strategies
- **implement** - Prior step: implementation
- **document-results** - Next step: documentation

---

## BDD Testing with jest-cucumber

For behavior-driven tests, use jest-cucumber to bind Gherkin feature files to Jest:

### 1. Create Feature File

```gherkin
# features/attachment.feature
Feature: AI tool attachment handling

  Scenario: Read attachment with string ID
    Given a task with attachments
    When I call readTaskAttachment with "attachment-id"
    Then I should receive the attachment content

  Scenario: Remove attachment with object argument
    Given a task with attachments
    When I call removeTaskAttachment with attachmentId "attachment-id"
    Then the attachment should be removed
```

### 2. Create Step Definitions

```typescript
import { defineFeature, loadFeature } from "jest-cucumber";

const feature = loadFeature("./features/attachment.feature");

defineFeature(feature, (test) => {
  let result: any;
  let task: Task;

  test("Read attachment with string ID", ({ given, when, then }) => {
    given("a task with attachments", async () => {
      task = await createTaskWithAttachments();
    });

    when(/I call readTaskAttachment with "(.*)"/, async (id) => {
      result = await service.readTaskAttachment(id);
    });

    then("I should receive the attachment content", async () => {
      expect(result).toBeDefined();
      expect(result.content).toBeTruthy();
    });
  });

  test("Remove attachment with object argument", ({ given, when, then }) => {
    given("a task with attachments", async () => {
      task = await createTaskWithAttachments();
    });

    when(/I call removeTaskAttachment with attachmentId "(.*)"/, async (id) => {
      result = await service.removeTaskAttachment({ attachmentId: id });
    });

    then("the attachment should be removed", async () => {
      expect(result.success).toBe(true);
    });
  });
});
```

### BDD Test Structure

```
specs/{feature}/
├── features/           # Gherkin .feature files
├── step-definitions/   # Jest step implementations
└── support/            # Shared utilities
```

**Evidence from footprint analysis:**

- feature-add-context-attachments: BDD tests used for complex AI tool scenarios

---

## AI Tool Testing Pattern

When testing AI tools, always test all argument formats to ensure flexibility:

```typescript
describe("removeTaskAttachment", () => {
  it("should accept string argument", async () => {
    const result = await service.removeTaskAttachment("attachment-id");
    expect(result.success).toBe(true);
  });

  it("should accept object with camelCase key", async () => {
    const result = await service.removeTaskAttachment({
      attachmentId: "attachment-id",
    });
    expect(result.success).toBe(true);
  });

  it("should accept object with snake_case key", async () => {
    const result = await service.removeTaskAttachment({
      attachment_id: "attachment-id",
    });
    expect(result.success).toBe(true);
  });
});
```

### AI Tool Test Checklist

When writing tests for AI tools:

- [ ] **String argument**: Direct ID as string (`"attachment-id"`)
- [ ] **camelCase object**: `{ attachmentId: "..." }`
- [ ] **snake_case object**: `{ attachment_id: "..." }`
- [ ] **Error handling**: Invalid ID, missing attachment
- [ ] **Response format**: Consistent success/error structure

**Evidence from footprint analysis:**

- bug/not-always-create-a-comment: Argument format inconsistency was root cause
- feature-add-context-attachments: All 3 argument formats tested
