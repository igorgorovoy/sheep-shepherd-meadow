# Review (Go)

Review work done against a specification file (`specs/*.md`) to ensure the implemented changes match requirements. Use the spec to understand requirements, then use the git diff to understand the changes. This is a **review**, not a test run — confirm the implementation matches what was requested and surface issues. Report issues if any; otherwise report success.

## Variables

feature_id: $1
spec_file: $2
agent_name: $3 if provided, otherwise use 'review_agent'
review_evidence_dir: `<absolute path to codebase>/reviews/<feature_id>/<agent_name>/evidence/`

## Instructions

- Check the current branch with `git branch` to understand context.
- Run `git diff origin/main` to see all changes on the current branch. Continue even if there are no matching changes.
- Find the spec file among `specs/*.md` that matches the current branch/feature; read it to understand requirements.
- **Verify the implementation against the spec** (this is a Go runtime/orchestrator/registry — there is no browser UI):
  - Read the changed Go files and confirm the behaviour matches the spec's acceptance criteria.
  - Prefer **behavioural evidence** over screenshots. Capture it as text/log files in `review_evidence_dir` (use full absolute paths):
    - `go build ./...`, `go vet ./...`, and relevant `go test ./...` output.
    - Where the change is user-facing, exercise the CLI and capture output, e.g.:
      - `sudo ./bin/sheep run --name hello -m 128m minimal /bin/echo hi`, `./bin/sheep ps -a`, `./bin/sheep inspect <id>`
      - `./bin/sheepctl apply -f examples/deployment.json`, `./bin/sheepctl get pods`, `./bin/sheepctl scale deployment/web --replicas=5`
      - registry: `curl -s localhost:5555/v2/_catalog`, `sheep push`/`sheep pull` round-trips
    - Save 1–5 evidence files demonstrating the critical path works as specified. Number them `01_<name>.log`, `02_<name>.log`, etc.
    - The `sheep` runtime only executes on Linux (namespaces/cgroups). If you are not on Linux, review the code + unit tests and note that runtime evidence must be captured on a Linux host — this is not a blocker for the review itself.
  - For any issue found, capture the relevant output/diff snippet into an evidence file and add it to `review_issues`.
- Issue Severity Guidelines — think hard about impact on the feature and the user:
  - `skippable` — non-blocker, but still a problem.
  - `tech_debt` — non-blocker that creates debt to address later.
  - `blocker` — must be fixed before release; breaks correctness, safety (e.g. namespace/cgroup/pivot_root handling), data integrity (BoltDB store), or the reconciliation loop.
- IMPORTANT: Return ONLY the JSON object described in `Report` — no extra text or markdown. We run `JSON.parse()` on the output.
- Think hard. Focus on critical functionality and correctness. Don't report non-critical issues as blockers.

## Report

- `success` is `true` if there are NO BLOCKING issues; `false` only if there ARE blocking issues.
- `review_issues` may contain issues of any severity.
- `evidence` should always contain absolute paths to the evidence files, regardless of success.

### Output Structure

```json
{
  "success": true,
  "review_summary": "2-4 sentences describing what was built and whether it matches the spec, written as a standup update.",
  "review_issues": [
    {
      "review_issue_number": 1,
      "evidence_path": "/absolute/path/to/evidence.log",
      "issue_description": "string",
      "issue_resolution": "string",
      "issue_severity": "skippable | tech_debt | blocker"
    }
  ],
  "evidence": [
    "/absolute/path/to/01_build.log",
    "/absolute/path/to/02_sheepctl_scale.log"
  ]
}
```
