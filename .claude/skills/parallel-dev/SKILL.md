---
name: parallel-dev
description: Orchestrate parallel feature development from Task Manager board. Dispatches N Claude Code agents (each running /sdlc in an isolated worktree), then merges PRs sequentially into master.
---

# Parallel Dev — Multi-Agent Orchestrator

Thin orchestration layer over the existing `/sdlc` skill. Reads cards from a Task Manager board list, dispatches parallel workers, manages merge queue.

```
Board (Task Manager)
  │
  ├── Select N cards from list
  │
  ├── Batch 1: Agent(worktree) × N  ──→  /sdlc each  ──→  PRs
  │                                                         │
  │             Code Review (pr-review-toolkit) per PR  ◄───┘
  │                        │
  │              Only APPROVED PRs ──→ Merge Queue ──→ master
  │
  ├── Batch 2: ... (same flow)
  │
  └── Summary Report
```

## Variables

- `board_id`: Board ID (alternative to board_name)
- `board_name`: Board name to find by search (alternative to board_id)
- `list_name`: Source list to pull cards from (default: "This Sprint")
- `max_agents`: Max parallel workers per batch (default: 3, max: 5)
- `dry_run`: Preview without executing (default: false)
- `merge_strategy`: PR merge order — "sequential" (default), "priority-first", "smallest-first"
- `filter_labels`: Only process cards with these labels (comma-separated, OR logic)
- `exclude_labels`: Skip cards with these labels (default: agent:merged, agent:processing, agent:failed)
- `comms_enabled`: Send Slack notifications via comms API (default: false)
- `force`: Process cards even if they have agent:* labels (default: false)

## Instructions

### Phase 1: Read Board & Select Cards

1. **Find the board**:
   - If `board_id` provided: call `tasks_get_board(board_id)`
   - If `board_name` provided: call `tasks_list_workspaces()`, then `tasks_list_boards(workspace_id)` for each, match by name (case-insensitive)
   - If neither: STOP with error "Provide board_id or board_name"

2. **Find the target list**:
   - Match `list_name` against board's lists (case-insensitive)
   - If not found: STOP with error listing available lists

3. **Filter cards**:
   - Remove cards with labels matching `exclude_labels` (unless `force=true`)
   - If `filter_labels` provided: keep only cards matching at least one
   - Cards with empty/null description are NOT skipped — they go through enrichment (step 3a.5)
   - Only skip cards with no title AND no description AND no attachments
   - Log skipped cards and reasons

4. **Sort eligible cards**:
   - Priority: urgent=0, high=1, medium=2, low=3 (ascending)
   - Then story_points ascending (smaller tasks first)

5. **Build batches**:
   - Batch 1: first `max_agents` cards
   - Batch 2: next `max_agents` cards
   - Continue until all cards assigned

6. **Dry-run check**: If `dry_run=true`, output preview table and STOP:

```
## Dry Run: /parallel-dev

Board: {board_name} | List: {list_name} | Max Agents: {max_agents}

### Batch 1 ({n} cards)
| # | Key | Title | Priority | SP | Branch |
|---|-----|-------|----------|-----|--------|
| 1 | EMM-42 | Add dark mode | high | 3 | feature/emm-42-add-dark-mode |

### Skipped ({n} cards)
- EMM-45: "Refactor auth" — label agent:processing

Total: {n} cards in {n} batches. {n} skipped.
```

### Phase 2: Initialize State

Create orchestrator state directory and files:

```
agentic/parallel-dev-{YYYYMMDD-HHmmss}/
  state.json
  footprints/
  workers/
```

**state.json**:
```json
{
  "workflow": "parallel-dev",
  "workflow_status": "in_progress",
  "board_id": "...",
  "board_name": "...",
  "list_name": "This Sprint",
  "max_agents": 3,
  "lists": {
    "source_list_id": "...",
    "in_progress_list_id": "...",
    "done_list_id": "...",
    "blocked_list_id": "..."
  },
  "batches": [],
  "summary": { "total_cards": 0, "dispatched": 0, "successful": 0, "failed": 0, "skipped": 0, "prs_merged": 0, "worktrees_removed": 0, "branches_deleted": 0, "disk_freed_mb": 0 },
  "started_at": "...",
  "completed_at": null
}
```

Find list IDs for "In Progress", "Done", "Blocked" from the board's lists. If any list is missing, create it via `tasks_create_list`.

Create footprint: `foot-parallel-start.md` with board info, card selection, batch assignment.

### Phase 3: Execute Batches

FOR each batch:

#### 3a. Prepare cards

For each card in batch:
1. `tasks_move_card(card_id, in_progress_list_id)` — move to "In Progress"
2. `tasks_update_card(card_id, labels=[...existing, "agent:processing"])` — add tracking label
3. `tasks_add_comment(card_id, "parallel-dev: Agent started on branch feature/{card_key_lower}-{title_slug}", "parallel-dev")` — log start

Generate branch name per card: `feature/{card_key_lower}-{title_slug}`
- card_key_lower: lowercase card key (e.g., "emm-42")
- title_slug: first 3-5 words of title in kebab-case (e.g., "add-dark-mode")

#### 3a.5. Card Enrichment (for poorly described cards)

Cards with missing or minimal descriptions get enriched before dispatch. This ensures workers always receive actionable task specifications.

**Trigger**: card has no description, or description is shorter than 50 characters.

```
FOR each card in batch:

    needs_enrichment = (card.description is null/empty) OR (len(card.description) < 50)

    IF NOT needs_enrichment:
        CONTINUE  # well-described card, no action needed

    # 1. Check for attachments (screenshots, files)
    attachments = tasks_list_attachments(card_id)

    # 2. If no description AND no attachments — skip card entirely
    IF (card.description is null/empty) AND (no attachments):
        tasks_move_card(card_id, blocked_list_id)
        tasks_add_comment(card_id, "parallel-dev: Skipped — no description and no attachments. Please add a description or screenshot.", "enricher")
        tasks_update_card(card_id, labels=[swap "agent:processing" → "agent:needs-description"])
        remove card from batch
        CONTINUE

    # 3. Enrich: agent reads title + existing description + attachments → generates structured description
    enrichment = Agent({
      description: "Enrich card {card_key}: {card_title}",
      prompt: "This task card needs a better description for autonomous development.

               ## Current Card
               - Key: {card_key}
               - Title: {card_title}
               - Current Description: {card.description or '(empty)'}
               - Attachments: {attachment_list with URLs/paths}

               ## Instructions
               1. Read the title carefully
               2. If there are screenshot attachments — read them and extract visual context
               3. Generate a structured task description:

               ### What needs to be done
               (Clear statement derived from title + visual context)

               ### Acceptance criteria
               (What 'done' looks like — bullet points)

               ### Likely affected components
               (Which files/modules are probably involved, based on title keywords)

               Output ONLY the structured description text, nothing else."
    })

    # 4. Update card description
    IF card.description exists AND len(card.description) > 0:
        # Append enrichment to existing description
        new_description = card.description + "\n\n---\n*Auto-enriched by parallel-dev:*\n" + enrichment
    ELSE:
        # Create new description from enrichment
        new_description = enrichment

    tasks_update_card(card_id, description=new_description)
    tasks_add_comment(card_id, "parallel-dev: Description enriched from title + attachments", "enricher")
```

Enrichment runs **sequentially** before worker dispatch. After enrichment, all cards in the batch have actionable descriptions.

#### 3b. Dispatch workers in parallel

Launch ALL workers for this batch in a **single message with multiple Agent() calls**:

```
Agent({
  description: "SDLC: {card_key} {card_title}",
  isolation: "worktree",
  prompt: WORKER_PROMPT(card)
})
```

All Agent calls in the same message = true parallel execution.

Each worker runs in an isolated locked git worktree (created by the Agent runtime under `.claude/worktrees/agent-*`). Record the worktree path and branch for every dispatched worker in `workers/worker-{card_key}.json` — Phase 4 needs both to clean up after merge.

#### 3c. Collect results

For each worker result:
- Parse the JSON output (status, branch_name, pr_number, pr_url, error)
- Save to `workers/worker-{card_key}.json` (include `worktree_path` if known)

**If success**:
- `tasks_update_card(card_id, labels=[swap "agent:processing" → "agent:pr-created"])`
- `tasks_add_comment(card_id, "parallel-dev: PR #{pr_number} created — {pr_url}", "parallel-dev")`

**If failed**:
- `tasks_move_card(card_id, blocked_list_id)`
- `tasks_update_card(card_id, labels=[swap "agent:processing" → "agent:failed"])`
- `tasks_add_comment(card_id, "parallel-dev: Failed — {error}", "parallel-dev")`

Update state.json with batch results.

#### 3d. Code Review — Graph-Aware (per PR)

Two-phase review: first gather structural context via `code-review-graph`, then pass it to a reviewer agent. This gives the reviewer blast-radius awareness, risk scores, and test gap detection.

**Prerequisite**: `code-review-graph` MCP server configured in `.mcp.json` and graph built (`code-review-graph build`).

```
FOR each successful worker (with pr_number):

    # Phase A: Graph analysis (MCP tools from code-review-graph)
    changed_files = git diff --name-only origin/master...{worker.branch_name}

    impact = call MCP tool: get_impact_radius(file_paths=changed_files, max_depth=2)
    # Returns: affected functions, classes, callers — blast radius

    risk = call MCP tool: detect_changes(base="origin/master", head={worker.branch_name})
    # Returns: risk_level, changed_functions, test_gap_count, security_keywords, review_priority

    context = call MCP tool: get_review_context(changed_files=changed_files)
    # Returns: token-optimized source snippets for changed areas only

    # Phase B: Reviewer agent with graph context
    review_result = Agent({
      description: "Review PR #{pr_number}: {card_key}",
      subagent_type: "pr-review-toolkit:code-reviewer",
      prompt: "Review PR #{pr_number} on branch {branch_name}.
               The PR implements: {card_title}. {card_description}.

               ## Graph Analysis (from code-review-graph)

               ### Blast Radius
               {impact}

               ### Risk Analysis
               {risk}

               ### Code Context
               {context}

               ## Review Focus
               1. High-risk changes (risk > 0.7) — these MUST be carefully reviewed
               2. Functions in blast radius that may break due to this change
               3. Missing test coverage flagged by test_gap_count
               4. Security concerns from keyword analysis
               5. General code quality, style, potential bugs

               Report: APPROVED or CHANGES_REQUESTED with specific issues."
    })

    IF review_result contains APPROVED:
        tasks_add_comment(card_id, "parallel-dev: Code review APPROVED (risk: {risk.risk_level})", "reviewer")
        → proceed to merge queue

    IF review_result contains CHANGES_REQUESTED:
        tasks_move_card(card_id, blocked_list_id)
        tasks_update_card(card_id, labels=[swap "agent:pr-created" → "agent:review-failed"])
        tasks_add_comment(card_id, "parallel-dev: Code review — changes requested:\n{issues}", "reviewer")
        → skip merge for this PR
```

Review agents run **sequentially** (not parallel) to keep API usage reasonable. Only APPROVED PRs proceed to the merge queue.

**After batch merge**: run `code-review-graph update` to incrementally update the graph with merged changes (for the next batch).

#### 3e. Merge queue

Run merge queue for APPROVED PRs in this batch (see Phase 4).

#### 3f. Update code graph

After merge queue completes for this batch, incrementally update the code-review-graph so the next batch's reviews have accurate structural data:

```
execute: code-review-graph update
```

#### 3g. Next batch

If more batches remain, continue from 3a. Workers in the next batch fork from updated master (post-merge) and reviews use the updated graph.

### Phase 4: Merge Queue

Sequential merge of successful PRs into master. After each PR merges, immediately reclaim that worker's worktree and local branch so worktrees do not accumulate across batches (left unchecked, they pile up — observed up to 17 worktrees / ~3.0 GB, with dead-PID locks blocking later removal).

```
FUNCTION merge_queue(successful_workers):

    sorted = sort by priority (urgent first), then story_points (smallest first)
    merged_files = []

    FOR worker IN sorted:

        # 1. Fetch latest master
        git fetch origin master

        # 2. Check file overlap (warning, not blocking)
        changed = git diff --name-only origin/master...{worker.branch_name}
        overlap = intersection(changed, merged_files)
        IF overlap:
            tasks_add_comment(worker.card_id,
              "parallel-dev: Warning — file overlap with already-merged PRs: {overlap}")

        # 3. Merge via gh CLI
        result = execute: gh pr merge {worker.pr_number} --merge --delete-branch
        IF result FAILED:
            tasks_move_card(worker.card_id, blocked_list_id)
            tasks_update_card(worker.card_id, labels=[..."agent:failed"])
            tasks_add_comment(worker.card_id, "parallel-dev: Merge failed — {error}")
            CONTINUE to next worker   # do NOT clean up — leave worktree/branch for inspection

        # 4. Post-merge verification (lightweight — Go)
        IF any *.go file in changed:
            verify = execute: go build ./... && go vet ./...
            # If tests are quick, also: go test ./... — see Makefile targets (build, test, vet)
            IF verify FAILED:
                execute: git revert HEAD --no-edit
                execute: git push
                tasks_move_card(worker.card_id, blocked_list_id)
                tasks_add_comment(worker.card_id, "parallel-dev: Post-merge check failed (go build/vet), reverted")
                CONTINUE   # do NOT clean up — leave worktree/branch for inspection

        # 5. Success
        merged_files.extend(changed)
        tasks_move_card(worker.card_id, done_list_id)
        tasks_update_card(worker.card_id, labels=[..."agent:merged"], finished=true)
        tasks_add_comment(worker.card_id, "parallel-dev: Merged to master")

        # 6. Reclaim this worker's worktree + local branch (only after a CLEAN, MERGED success)
        cleanup_merged_worker(worker)
```

#### 4a. Per-worker cleanup after merge

Runs ONLY for a worker whose PR merged successfully and passed the post-merge check. Reclaims the isolated worktree and the now-merged local branch.

```
FUNCTION cleanup_merged_worker(worker):

    # SAFETY GATE — never destroy unmerged or dirty work.
    # Only worktrees/branches this skill created (.claude/worktrees/agent-*,
    # feature/* per card) are eligible. Skip and SURFACE anything else.

    wt = worker.worktree_path        # e.g. .claude/worktrees/agent-xxxx
    br = worker.branch_name          # e.g. feature/emm-42-add-dark-mode

    # 1. Confirm the branch is actually merged into master.
    #    `gh pr merge` above already merged the PR, but verify locally before -D.
    IF NOT (git branch --merged origin/master | contains br):
        tasks_add_comment(worker.card_id,
          "parallel-dev: Skipped worktree/branch cleanup — branch {br} not confirmed merged. Left for manual review.")
        RETURN

    # 2. Confirm the worktree has no uncommitted/dirty changes.
    IF wt exists:
        dirty = execute: git -C {wt} status --porcelain
        IF dirty is non-empty:
            tasks_add_comment(worker.card_id,
              "parallel-dev: Skipped worktree cleanup — {wt} has uncommitted changes. Left for manual review.")
            RETURN   # do NOT force-remove dirty work

        # 3. Remove the worktree. --force is acceptable here ONLY because we
        #    have already verified (a) merged and (b) clean above.
        execute: git worktree remove --force {wt}
        worktrees_removed += 1

    # 4. Delete the merged local branch (-D is safe: it is confirmed merged).
    IF git branch --list br is non-empty:
        execute: git branch -D {br}
        branches_deleted += 1

    tasks_add_comment(worker.card_id,
      "parallel-dev: Cleaned up worktree + local branch (merged & clean).")
```

> **SAFETY (must hold):** Auto-cleanup touches ONLY worktrees and branches this
> skill created (`.claude/worktrees/agent-*`, `feature/*` per card). Never
> force-remove a worktree with uncommitted/dirty changes, and never `git branch -D`
> a branch that is not confirmed merged into `origin/master`. If either check
> fails, **skip and surface it** (comment on the card + report in the summary) —
> leave the worktree/branch in place for manual review. Failed/reverted/
> review-rejected workers are NEVER auto-cleaned; their worktrees stay for
> inspection. (Mirrors the project rule on worktree hygiene.)

### Phase 5: Summary Report & Final Cleanup Sweep

Before writing the summary, run one idempotent cleanup sweep to catch leftovers from earlier batches, crashed workers, or stale locks left by dead processes. This is safe to run repeatedly.

#### 5a. Idempotent cleanup sweep

```
FUNCTION final_cleanup_sweep():

    # 1. Prune worktree admin records whose directories are already gone.
    execute: git worktree prune

    # 2. Walk remaining skill-created worktrees and reclaim dead-locked ones.
    FOR wt IN (git worktree list --porcelain → paths matching .claude/worktrees/agent-*):

        # SAFETY: dirty worktree → never auto-remove. Surface it.
        dirty = execute: git -C {wt} status --porcelain
        IF dirty is non-empty:
            report_line: "DIRTY worktree left in place: {wt}"
            CONTINUE

        # A worktree may be LOCKED. A lock is stale only if its locking PID is dead.
        lock_reason = lock info from `git worktree list --porcelain` (locked line)
        pid = parse PID from lock reason (if the lock encodes one)
        IF worktree is locked:
            IF pid is present AND `ps -p {pid}` shows the process ALIVE:
                # A live worker still owns it — do not touch.
                report_line: "Active worktree (pid {pid} alive) left in place: {wt}"
                CONTINUE
            ELSE:
                # Dead/absent PID → stale lock. Unlock, then remove.
                execute: git worktree unlock {wt}

        # Only here: clean (or just-unlocked-stale) worktree → remove.
        # Branch is verified merged before -D, same as Phase 4a.
        size_before = du -sm {wt}
        execute: git worktree remove --force {wt}
        worktrees_removed += 1
        disk_freed_mb += size_before

        br = branch that was checked out in {wt}
        IF br present AND (git branch --merged origin/master | contains br):
            execute: git branch -D {br}
            branches_deleted += 1

    # 3. Final prune to clear any records freed by the removals above.
    execute: git worktree prune
```

Record `worktrees_removed`, `branches_deleted`, and `disk_freed_mb` into
`state.json.summary`. Any DIRTY or still-ACTIVE worktree that was left in place
must be listed explicitly in the summary so a human can act on it.

#### 5b. Summary report

Create footprint `foot-parallel-complete.md` and output summary:

```
## Parallel Dev Complete

Board: {board_name} | List: {list_name}
Duration: {start} → {end}

### Results
| # | Key | Title | Status | PR |
|---|-----|-------|--------|-----|
| 1 | EMM-42 | Add dark mode | merged | #101 |
| 2 | EMM-43 | Fix login | merged | #102 |
| 3 | EMM-44 | Refactor auth | failed | — |

### Summary
- Total cards: 3
- Merged: 2
- Failed: 1
- Skipped: 0

### Cleanup
- Worktrees removed: {worktrees_removed}
- Local branches deleted: {branches_deleted}
- Disk freed: {disk_freed_mb} MB
- Left in place (needs attention): {list of dirty/active/unmerged worktrees, or "none"}
```

If `comms_enabled`: send summary to Slack via `curl -s -X POST http://localhost:8000/api/comms/notify`.

---

## Worker Prompt Template

```
FUNCTION WORKER_PROMPT(card):
    RETURN:

You are a worker agent in an isolated git worktree.
Run the existing /sdlc skill for this Task Manager card.

## Card
- Key: {card.key}
- Title: {card.title}
- Description: {card.description}
- Priority: {card.priority}
- Story Points: {card.story_points}

## Execute /sdlc
Run the /sdlc skill with these parameters:
- input: "{card.title}. {card.description}"
- approval_level: "none"
- skip_merge_approval: true
- skip_branch_switch: true

IMPORTANT:
- Use the existing /sdlc skill. Do NOT reimplement the SDLC pipeline.
- STOP after /commit-and-pr. Do NOT run /merge.
- The orchestrator handles merge coordination.
- If any phase fails, stop and report the error.

## Output
When complete, output this JSON:
{
  "status": "success" or "failed",
  "branch_name": "<your branch>",
  "pr_number": <number or null>,
  "pr_url": "<url or null>",
  "error": "<error message or null>"
}
```

---

## Error Handling

| Failure | Detection | Action |
|---------|-----------|--------|
| Board not found | tasks_get_board returns error | STOP, report error |
| List not found | No matching list name | STOP, list available lists |
| No eligible cards | All filtered out | STOP, report reasons |
| Card has no description | description is empty | Skip, add to skipped list |
| Worker fails | Agent returns status: "failed" | Card → Blocked, label agent:failed, continue (worktree left for inspection) |
| Worker timeout | Agent tool timeout | Treat as failure (worktree left for inspection) |
| Malformed worker output | JSON parse error | Treat as failure with parse error message |
| Code review rejects PR | Reviewer returns CHANGES_REQUESTED | Card → Blocked, label agent:review-failed, skip merge (worktree left for inspection) |
| Merge conflict | gh pr merge fails | Card → Blocked, comment with details, continue (no cleanup) |
| Post-merge check fails | go build/vet returns non-zero | git revert, card → Blocked (no cleanup) |
| Worktree dirty at cleanup | git status --porcelain non-empty | Skip removal, surface in card comment + summary |
| Worktree branch not merged | git branch --merged excludes it | Skip removal, surface in card comment + summary |
| Stale worktree lock (dead PID) | `ps -p {pid}` shows no process | git worktree unlock, then remove |
| Active worktree lock (live PID) | `ps -p {pid}` shows running process | Leave in place, report in summary |
| MCP unreachable | tasks_* call fails | Warn, continue without card updates |

---

## Idempotency Labels

| Label | Meaning | When added |
|-------|---------|------------|
| `agent:processing` | Worker is executing | Card dispatched |
| `agent:pr-created` | PR exists | Worker succeeds |
| `agent:merged` | PR merged to master | Merge queue succeeds |
| `agent:review-failed` | Code review requested changes | Review step rejects PR |
| `agent:needs-description` | Card has no description or attachments | Enrichment step can't proceed |
| `agent:failed` | Worker or merge failed | Any failure |

Cards with `agent:merged`, `agent:processing`, or `agent:failed` are skipped unless `force=true`.

---

## Comms Integration (optional)

When `comms_enabled=true`:
- Batch start: `"parallel-dev: Starting batch {n}/{total}: {card_keys}"`
- Batch done: `"parallel-dev: Batch {n} done — {ok} merged, {fail} failed"`
- Workflow done: `"parallel-dev complete: {merged}/{total} merged. PRs: {urls}"`

```bash
curl -s -X POST http://localhost:8000/api/comms/notify \
  -H 'Content-Type: application/json' \
  -d '{"message": "..."}'
```

---

## Required Actions

1. Find board and select cards
2. Create state directory and initialize state.json
3. For each batch: prepare cards → dispatch parallel workers → collect results → merge queue → clean up merged workers' worktrees + branches
4. Run the idempotent final cleanup sweep (prune + stale-lock reclaim)
5. Create summary report (including cleanup stats and any worktrees left in place)
6. Update all card states in Task Manager

## Notes

- This skill does NOT contain SDLC logic — workers run the existing `/sdlc` skill
- Workers operate in isolated git worktrees — no filesystem conflicts
- The merge queue runs in the main worktree on the current branch
- Worktrees and branches are reclaimed only after a clean, merged success; dirty/unmerged/failed work is always left in place and surfaced
- Rate limits: max_agents=3 is the practical parallel limit for Tier 2-3 API access
