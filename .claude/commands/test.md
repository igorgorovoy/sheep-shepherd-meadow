# Application Validation Test Suite (Go)

Execute comprehensive validation tests for the Sheep & Shepherd Go codebase, returning results in a standardized JSON format for automated processing.

## Purpose

Proactively identify and fix issues before they impact users or developers:

- Detect compile errors, type mismatches, and unused imports
- Catch suspicious constructs via `go vet`
- Verify formatting, tests, and the build of all three binaries (`sheep`, `shepherd`, `sheepctl`)

## Variables

TEST_COMMAND_TIMEOUT: 5 minutes
affected_packages: $1 - Comma-separated Go import paths to test (e.g. "./internal/shepherd,./internal/container"). If not provided, run all tests with `./...`.

## Instructions

- **Determine scope**:
  - If `affected_packages` is provided, restrict `go test` / `go vet` to those packages.
  - If not provided, use `./...` (all packages).
- Execute each test in the sequence below, from the repo root.
- Capture the result (passed/failed) and any error messages (stderr).
- IMPORTANT: Return ONLY the JSON array with test results — no extra text or markdown. We run `JSON.parse()` on the output.
- If a test passes, omit the `error` field. If it fails, include stderr in `error`.
- Error Handling:
  - If a command returns a non-zero exit code, mark as failed and **stop processing further tests**, returning results so far.
  - Timeout commands after `TEST_COMMAND_TIMEOUT`.
- The project cross-compiles on any OS, but the `sheep` runtime only *runs* on Linux (namespaces/cgroups). Building and testing is fine on macOS thanks to the `_linux.go` / `_stub.go` split — never gate the suite on being on Linux.
- Always `pwd` before running to ensure you are at the repo root (where `go.mod` and the `Makefile` live).

## Test Execution Sequence

Prefer the `Makefile` targets when present; the raw `go` commands below are the fallback.

1. **Format check** — code must be gofmt-clean
   - Command: `test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)`
   - test_name: "gofmt"
   - test_purpose: "Verifies all Go source is gofmt-formatted"

2. **Vet** — static analysis for suspicious constructs
   - Command: `go vet ./...` (or the provided packages)
   - test_name: "go_vet"
   - test_purpose: "Catches common mistakes go build won't (printf mismatches, struct tags, etc.)"

3. **Build** — all binaries compile
   - Command: `make build` (falls back to `go build ./...`)
   - test_name: "build"
   - test_purpose: "Compiles sheep, shepherd, and sheepctl for the host platform"

4. **Unit / integration tests**
   - Command: `make test` (falls back to `go test ./...`) — scope to `affected_packages` when provided
   - test_name: "go_test"
   - test_purpose: "Runs table-driven unit tests and package integration tests (store, scheduler, controllers, image)"

Add `-race` (`go test -race ./...`) for the scheduler/controller/store packages, which have concurrent reconcilers.

## Report

- Return results exclusively as a JSON array based on the `Output Structure` below.
- Sort the array with failed tests (`passed: false`) at the top.
- Include all tests, both passed and failed.
- `execution_command` must be the exact command to reproduce the test.

### Output Structure

```json
[
  {
    "test_name": "string",
    "passed": true,
    "execution_command": "string",
    "test_purpose": "string",
    "error": "optional string"
  }
]
```

### Example Output

```json
[
  {
    "test_name": "go_test",
    "passed": false,
    "execution_command": "go test ./internal/shepherd/...",
    "test_purpose": "Runs scheduler and controller unit tests",
    "error": "--- FAIL: TestScheduler_ScoreNodes (0.00s)\n    scheduler_test.go:42: expected node-1, got node-2"
  },
  {
    "test_name": "build",
    "passed": true,
    "execution_command": "make build",
    "test_purpose": "Compiles sheep, shepherd, and sheepctl for the host platform"
  }
]
```
