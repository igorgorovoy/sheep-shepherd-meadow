# In-Loop Review (Go)

Quick checkout, build, and run workflow for validating agent work on the Sheep & Shepherd stack.

## Variables

branch: $ARGUMENT

## Workflow

IMPORTANT: If no branch is provided, stop and report that a branch argument is required.

### Step 1: Pull and Checkout Branch
- `git fetch origin`
- `git checkout {branch}`

### Step 2: Build
- `make build` (produces `bin/sheep`, `bin/shepherd`, `bin/sheepctl`); fallback `go build ./...`
- `go vet ./...` and `go test ./...` for a fast sanity check

### Step 3: Run the stack for manual inspection
- Orchestrator (single-node dev): `sudo ./bin/shepherd --mode standalone` (run in background)
- In another shell: `export SHEPHERD_API=localhost:9876`
- Registry (if the change touches images/pull/push): `./bin/meadow --addr :5555` (run in background)
- Give daemons a moment to come up (e.g. `sleep 2`)

### Step 4: Exercise the change
- Orchestration: `./bin/sheepctl apply -f examples/deployment.json`, `./bin/sheepctl get pods`, `./bin/sheepctl nodes`, `./bin/sheepctl events`
- Runtime: `sudo ./bin/sheep run --name shell minimal /bin/sh`, `sudo ./bin/sheep ps -a`, `sudo ./bin/sheep inspect <id>`
- NOTE: the `sheep` runtime requires **Linux + root** (namespaces, cgroups v2, networking). On macOS you can build and run unit tests, but runtime commands must be exercised on a Linux host.

## Report

Report the steps taken (checkout, build result, vet/test result, which daemons are running and on which ports) so a human can inspect the running stack.
