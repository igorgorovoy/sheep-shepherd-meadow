# Media Plan: Sheep & Shepherd Blog Series

A series of technical articles based on the open-source Sheep project — a container runtime, orchestrator, and image registry written from scratch in Go.

**Codebase:** ~5000 lines of Go → 40 articles
**Audience:** Backend/DevOps/SRE engineers, Go developers, anyone who wants to understand how Docker and Kubernetes work under the hood
**Format:** Technical article with code fragments from the repository, diagrams, and practical examples

---

## Series 1: Container Runtime from Scratch (Sheep = Docker)

| # | Article | What it covers | Key files |
|---|--------|-------------|---------------|
| 1 | **Linux Namespaces: isolating a process in 50 lines of Go** | PID, NET, MNT, UTS, IPC namespaces, clone() with flags | `runtime_linux.go` |
| 2 | **Re-Exec Pattern: why Go and clone() don't get along** | Go threading model, the goroutines problem, self-re-exec via "init" | `runtime_linux.go`, `cmd/sheep/main.go` |
| 3 | **pivot_root: how a container gets its own filesystem** | bind mount, pivot_root(2), unmount old root, creating /dev nodes | `runtime_linux.go` |
| 4 | **Cgroups v2: limiting memory, CPU, and PIDs** | memory.max, cpu.max, pids.max, cpu.weight, subtree_control | `runtime_linux.go` |
| 5 | **OverlayFS: copy-on-write layers like in Docker** | lower/upper/work/merged, read flow, write flow (copy-up) | `manager.go`, `runtime_linux.go` |
| 6 | **Bridge Networking: giving a container an IP address** | veth pairs, Linux bridge, IP allocation, nsenter | `network_linux.go` |
| 7 | **NAT and iptables: how a container reaches the internet** | ip_forward, MASQUERADE, traffic flow container→internet | `network_linux.go` |
| 8 | **Image Management: tar archive → rootfs → container** | Import, bootstrap, manifest.json, image store layout | `image.go` |
| 9 | **Container Lifecycle: a state machine from Created to Removed** | Create→Start→Stop→Remove, PID tracking, state persistence | `manager.go`, `container.go` |
| 10 | **A Docker CLI in 500 lines of Go** | run, ps, stop, rm, inspect, logs — subcommand routing, flag parsing | `cmd/sheep/main.go` |

---

## Series 2: Orchestrator from Scratch (Shepherd = Kubernetes)

| # | Article | What it covers | Key files |
|---|--------|-------------|---------------|
| 11 | **A Kubernetes API Server in 300 lines** | REST endpoints, CRUD for pods/services/deployments/nodes, namespaced resources | `apiserver.go` |
| 12 | **BoltDB instead of etcd: an embedded state store** | Buckets, transactions, namespace/name keys, watch mechanism | `store.go` |
| 13 | **Scheduler: how to pick a node for a pod** | Filter phase (Ready? Resources? Labels?), Score phase (least-loaded, balance) | `scheduler.go` |
| 14 | **Reconciliation Loop: the heart of Kubernetes** | Desired vs actual state, observe→compare→act→record, periodic convergence | `controller.go` |
| 15 | **Replication Controller: scale up and scale down** | Deployment spec → pod count, label matching, ordinal naming (web-0, web-1) | `controller.go` |
| 16 | **Service Discovery: how a Service finds Pods** | Label selector matching, endpoint building, Running+IP filter | `controller.go` |
| 17 | **Node Health: heartbeat and failure detection** | 10s heartbeat, 30s timeout, Ready→NotReady transitions | `controller.go`, `agent.go` |
| 18 | **Node Agent: a kubelet in 350 lines** | Registration, heartbeat loop, pod reconciliation, container lifecycle | `agent.go` |
| 19 | **Pod Lifecycle: from Pending to Running** | API→Scheduler→Agent→Runtime, async scheduling, status updates | all shepherd files |
| 20 | **Event System: an audit trail for the cluster** | Event recording, type/reason/message, event sourcing pattern | `store.go`, `types.go` |

---

## Series 3: Image Registry from Scratch (Meadow = Docker Registry)

| # | Article | What it covers | Key files |
|---|--------|-------------|---------------|
| 21 | **OCI Distribution Spec: writing your own Docker Registry** | /v2/ endpoints, blob upload, manifest push/pull | `server.go` |
| 22 | **Content-Addressable Storage: SHA256 as the key** | Digest computation, atomic writes, temp→rename pattern | `storage.go` |
| 23 | **Pulling from Docker Hub: auth, manifest lists, multi-arch** | Token auth, manifest list → platform manifest → layer download | `registry.go` |
| 24 | **Pushing to your own registry: layer creation and manifest upload** | tar.gz layer from rootfs, config blob, OCI manifest format | `push.go` |

---

## Series 4: Distributed Systems Patterns

| # | Article | What it covers | Key files |
|---|--------|-------------|---------------|
| 25 | **Desired State vs Actual State: the Kubernetes paradigm** | All controllers as examples, why this works better than imperative | `controller.go` |
| 26 | **Label Selectors: linking resources without foreign keys** | matchLabels, deployment→pods, service→pods, nodeSelector | across files |
| 27 | **Async Scheduling: why a Pod is created as Pending** | Decoupled creation/scheduling, eventual consistency | `apiserver.go`, `scheduler.go` |
| 28 | **Graceful Shutdown: signals, channels, cleanup** | SIGTERM/SIGKILL, stop channels, goroutine coordination | `cmd/shepherd/main.go` |
| 29 | **Two-Mode Architecture: server + agent** | Control plane vs worker, HTTP-based coordination, standalone mode | `cmd/shepherd/main.go` |
| 30 | **Embedded vs External DB: BoltDB vs etcd trade-offs** | Single-node simplicity vs distributed consensus | `store.go` |

---

## Series 5: Go Systems Programming

| # | Article | What it covers | Key files |
|---|--------|-------------|---------------|
| 31 | **Build Tags: one codebase for Linux and macOS** | `//go:build linux`, platform stubs, compile everywhere run on Linux | `runtime_*.go` |
| 32 | **Go's syscall package: mount, clone, pivot_root** | Direct system calls from Go, SysProcAttr, unix package | `runtime_linux.go` |
| 33 | **Goroutines for control loops** | Ticker-based reconciliation, channel shutdown, concurrent controllers | `scheduler.go`, `controller.go` |
| 34 | **A CLI without frameworks: subcommand routing** | os.Args parsing, flag package, subcommand dispatch | `cmd/sheep/main.go` |
| 35 | **JSON serialization for state persistence** | MarshalIndent, state.json, manifest.json, API request/response | across files |

---

## Series 6: Architecture & Design

| # | Article | What it covers | Key files |
|---|--------|-------------|---------------|
| 36 | **Docker vs Sheep: what a container runtime actually does** | Architecture comparison, what we simplified, what we kept | `docs/architecture.md` |
| 37 | **Kubernetes vs Shepherd: a minimal orchestrator** | API Server + Scheduler + Controllers + Agent, what can be cut | `docs/architecture.md` |
| 38 | **Filesystem Layout: where containers live** | /var/lib/sheep structure, overlay dirs, image store | `docs/data-model.md` |
| 39 | **State Machines in infrastructure software** | Pod phases, Node conditions, Container states, transitions | `docs/data-model.md` |
| 40 | **Entity Relationships: how Pod, Service, Deployment, and Node connect** | ER diagram, label-based coupling, namespace scoping | `docs/data-model.md` |

---

## Summary

| Series | Count | Topics |
|-------|-----------|----------|
| 1. Container Runtime | 10 | Linux primitives, isolation, networking |
| 2. Orchestrator | 10 | Scheduling, controllers, reconciliation |
| 3. Image Registry | 4 | OCI spec, content-addressable storage |
| 4. Distributed Systems | 6 | Patterns, trade-offs, architecture |
| 5. Go Systems Programming | 5 | Syscalls, concurrency, build tags |
| 6. Architecture & Design | 5 | Comparisons, data models, state machines |
| **Total** | **40** | **~5000 lines of Go code** |

---

## Metrics

- **Source:** 26 source files, 4932 lines of Go
- **Articles:** 40
- **Average depth:** ~120 lines of code per article
- **Each article:** self-contained, with code fragments from the repository and Mermaid diagrams
- **Repository:** open-source, readers can run it and experiment
