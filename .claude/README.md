# Claude Code setup — Sheep & Shepherd

SDLC + parallel-dev tooling adapted for this Go container platform (runtime `sheep`, orchestrator `shepherd`, registry `meadow`). Ported from `agentic-ai-landing-zone` and tuned to Go conventions (`gofmt` / `go vet` / `go build ./...` / `go test ./...`, the `_linux.go` / `_stub.go` split, and this repo's layout).

## Agents (`.claude/agents/`)

Role-based subagents you can dispatch (e.g. via the Agent tool or from the orchestration skills):

| Agent | Role | Use for |
|-------|------|---------|
| `go-architect` | Architect | System design, ADRs, C4/Mermaid diagrams, REST/OCI API contracts. Designs, doesn't code. |
| `go-engineer` | Engineer | Idiomatic Go implementation across runtime, orchestrator, registry, CLIs. |
| `devops-engineer` | DevOps | Makefile, CI/CD, cross-compile, multi-node cluster ops, registry deploy, secrets. |
| `qa-engineer` | QA | Test plans + table-driven / `-race` Go tests + CLI e2e; bug reports. |
| `frontend-engineer` | Frontend | CLI output UX + docs/diagrams. **No web UI exists yet** — scoped accordingly. |

## Skills (`.claude/skills/`)

- **Orchestration:** `parallel-dev` (dispatches N workers running `/sdlc` in isolated worktrees, then a sequential merge queue — post-merge check adapted to `go build ./... && go vet ./...`), `subagent-driven-development`, `dispatching-parallel-agents`.
- **SDLC:** `requirements`, `requirements-clarity`, `feature`, `bug`, `chore`, `patch`, `implement`, `test`, `review`, `risk-assess`.
- **Architect:** `architecture-decision-records`, `architecture-patterns`, `c4-architecture`, `api-design-principles`, `mermaid-diagrams`.
- **DevOps:** `deployment-pipeline-design`, `gitops-workflow`, `secrets-management`, `cost-optimization`.
- **QA:** `qa-test-planner`.

## Commands (`.claude/commands/`)

`commit`, `generate_branch_name`, `pull_request`, `implement` (verbatim), plus Go-adapted `test`, `review`, and `in_loop_review` (build/vet/test + run the daemons; runtime evidence needs Linux + root).

## Typical flow

1. `requirements` → clarify → classify into `feature` / `bug` / `chore` / `patch` (plan in `specs/`).
2. `go-architect` for ADRs/diagrams/contracts on anything non-trivial.
3. `implement` (or the `/sdlc` plugin skill end-to-end) → `go-engineer` builds it.
4. `qa-engineer` + `test` command validate; `review` command checks against the spec.
5. `commit` → `pull_request`. For batches of independent cards, `parallel-dev` fans out workers and merges sequentially.

## Notes & prerequisites

- `parallel-dev` chains the **`/sdlc` plugin skill** and expects a Task Manager board (the `tasks-*` MCP servers) plus, optionally, the `code-review-graph` MCP for graph-aware review. It's an advanced/optional path — the per-role agents and SDLC skills work standalone without it.
- Many source skills reference Java/Spring/React/Python; those were **not** copied. Only Go-relevant and language-agnostic skills are included here.
- Runtime reality: `sheep` requires **Linux + root** (namespaces, cgroups v2, networking). Build and unit tests run anywhere via the stub split; runtime/e2e evidence must be captured on Linux.
