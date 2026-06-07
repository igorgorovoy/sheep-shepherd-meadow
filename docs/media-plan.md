# Media Plan: Sheep & Shepherd Blog Series

Серія технічних статей на базі open-source проекту Sheep — контейнерний рантайм, оркестратор та реєстр образів написані з нуля на Go.

**Кодова база:** ~5000 рядків Go → 40 статей
**Аудиторія:** Backend/DevOps/SRE інженери, Go розробники, всі хто хоче зрозуміти як працює Docker і Kubernetes зсередини
**Формат:** Технічна стаття з фрагментами коду з репозиторію, діаграмами, практичними прикладами

---

## Серія 1: Container Runtime з нуля (Sheep = Docker)

| # | Стаття | Що покриває | Ключові файли |
|---|--------|-------------|---------------|
| 1 | **Linux Namespaces: ізолюємо процес за 50 рядків Go** | PID, NET, MNT, UTS, IPC namespaces, clone() з прапорцями | `runtime_linux.go` |
| 2 | **Re-Exec Pattern: чому Go і clone() не дружать** | Go threading model, проблема з goroutines, self-re-exec через "init" | `runtime_linux.go`, `cmd/sheep/main.go` |
| 3 | **pivot_root: як контейнер отримує свою файлову систему** | bind mount, pivot_root(2), unmount old root, створення /dev nodes | `runtime_linux.go` |
| 4 | **Cgroups v2: обмежуємо пам'ять, CPU та PIDs** | memory.max, cpu.max, pids.max, cpu.weight, subtree_control | `runtime_linux.go` |
| 5 | **OverlayFS: copy-on-write шари як у Docker** | lower/upper/work/merged, read flow, write flow (copy-up) | `manager.go`, `runtime_linux.go` |
| 6 | **Bridge Networking: даємо контейнеру IP-адресу** | veth pairs, Linux bridge, IP allocation, nsenter | `network_linux.go` |
| 7 | **NAT і iptables: як контейнер бачить інтернет** | ip_forward, MASQUERADE, traffic flow container→internet | `network_linux.go` |
| 8 | **Image Management: tar-архів → rootfs → контейнер** | Import, bootstrap, manifest.json, image store layout | `image.go` |
| 9 | **Container Lifecycle: state machine від Created до Removed** | Create→Start→Stop→Remove, PID tracking, state persistence | `manager.go`, `container.go` |
| 10 | **Docker CLI за 500 рядків Go** | run, ps, stop, rm, inspect, logs — subcommand routing, flag parsing | `cmd/sheep/main.go` |

---

## Серія 2: Оркестратор з нуля (Shepherd = Kubernetes)

| # | Стаття | Що покриває | Ключові файли |
|---|--------|-------------|---------------|
| 11 | **Kubernetes API Server за 300 рядків** | REST endpoints, CRUD для pods/services/deployments/nodes, namespaced resources | `apiserver.go` |
| 12 | **BoltDB замість etcd: embedded state store** | Buckets, transactions, namespace/name ключі, watch mechanism | `store.go` |
| 13 | **Scheduler: як вибрати ноду для поду** | Filter phase (Ready? Resources? Labels?), Score phase (least-loaded, balance) | `scheduler.go` |
| 14 | **Reconciliation Loop: серце Kubernetes** | Desired vs actual state, observe→compare→act→record, periodic convergence | `controller.go` |
| 15 | **Replication Controller: scale up і scale down** | Deployment spec → pod count, label matching, ordinal naming (web-0, web-1) | `controller.go` |
| 16 | **Service Discovery: як Service знаходить Pods** | Label selector matching, endpoint building, Running+IP filter | `controller.go` |
| 17 | **Node Health: heartbeat і failure detection** | 10s heartbeat, 30s timeout, Ready→NotReady transitions | `controller.go`, `agent.go` |
| 18 | **Node Agent: kubelet за 350 рядків** | Registration, heartbeat loop, pod reconciliation, container lifecycle | `agent.go` |
| 19 | **Pod Lifecycle: від Pending до Running** | API→Scheduler→Agent→Runtime, async scheduling, status updates | all shepherd files |
| 20 | **Event System: audit trail для кластера** | Event recording, type/reason/message, event sourcing pattern | `store.go`, `types.go` |

---

## Серія 3: Image Registry з нуля (Meadow = Docker Registry)

| # | Стаття | Що покриває | Ключові файли |
|---|--------|-------------|---------------|
| 21 | **OCI Distribution Spec: пишемо свій Docker Registry** | /v2/ endpoints, blob upload, manifest push/pull | `server.go` |
| 22 | **Content-Addressable Storage: SHA256 як ключ** | Digest computation, atomic writes, temp→rename pattern | `storage.go` |
| 23 | **Pull з Docker Hub: auth, manifest lists, multi-arch** | Token auth, manifest list → platform manifest → layer download | `registry.go` |
| 24 | **Push в свій реєстр: layer creation і manifest upload** | tar.gz layer from rootfs, config blob, OCI manifest format | `push.go` |

---

## Серія 4: Distributed Systems Patterns

| # | Стаття | Що покриває | Ключові файли |
|---|--------|-------------|---------------|
| 25 | **Desired State vs Actual State: парадигма Kubernetes** | Всі контролери як приклад, чому це працює краще ніж imperative | `controller.go` |
| 26 | **Label Selectors: як зв'язати ресурси без foreign keys** | matchLabels, deployment→pods, service→pods, nodeSelector | across files |
| 27 | **Async Scheduling: чому Pod створюється як Pending** | Decoupled creation/scheduling, eventual consistency | `apiserver.go`, `scheduler.go` |
| 28 | **Graceful Shutdown: signals, channels, cleanup** | SIGTERM/SIGKILL, stop channels, goroutine coordination | `cmd/shepherd/main.go` |
| 29 | **Two-Mode Architecture: server + agent** | Control plane vs worker, HTTP-based coordination, standalone mode | `cmd/shepherd/main.go` |
| 30 | **Embedded vs External DB: BoltDB vs etcd trade-offs** | Single-node simplicity vs distributed consensus | `store.go` |

---

## Серія 5: Go Systems Programming

| # | Стаття | Що покриває | Ключові файли |
|---|--------|-------------|---------------|
| 31 | **Build Tags: один код для Linux і macOS** | `//go:build linux`, platform stubs, compile everywhere run on Linux | `runtime_*.go` |
| 32 | **syscall пакет Go: mount, clone, pivot_root** | Прямі системні виклики з Go, SysProcAttr, unix package | `runtime_linux.go` |
| 33 | **Goroutines для control loops** | Ticker-based reconciliation, channel shutdown, concurrent controllers | `scheduler.go`, `controller.go` |
| 34 | **CLI без фреймворків: subcommand routing** | os.Args parsing, flag package, subcommand dispatch | `cmd/sheep/main.go` |
| 35 | **JSON serialization для state persistence** | MarshalIndent, state.json, manifest.json, API request/response | across files |

---

## Серія 6: Architecture & Design

| # | Стаття | Що покриває | Ключові файли |
|---|--------|-------------|---------------|
| 36 | **Docker vs Sheep: що насправді робить контейнерний рантайм** | Порівняння архітектур, що ми спростили, що зберегли | `docs/architecture.md` |
| 37 | **Kubernetes vs Shepherd: мінімальний оркестратор** | API Server + Scheduler + Controllers + Agent, що можна вирізати | `docs/architecture.md` |
| 38 | **Filesystem Layout: де живуть контейнери** | /var/lib/sheep structure, overlay dirs, image store | `docs/data-model.md` |
| 39 | **State Machines в інфраструктурному софті** | Pod phases, Node conditions, Container states, transitions | `docs/data-model.md` |
| 40 | **Entity Relationships: як пов'язані Pod, Service, Deployment, Node** | ER diagram, label-based coupling, namespace scoping | `docs/data-model.md` |

---

## Зведення

| Серія | Кількість | Тематика |
|-------|-----------|----------|
| 1. Container Runtime | 10 | Linux primitives, isolation, networking |
| 2. Orchestrator | 10 | Scheduling, controllers, reconciliation |
| 3. Image Registry | 4 | OCI spec, content-addressable storage |
| 4. Distributed Systems | 6 | Patterns, trade-offs, architecture |
| 5. Go Systems Programming | 5 | Syscalls, concurrency, build tags |
| 6. Architecture & Design | 5 | Comparisons, data models, state machines |
| **Разом** | **40** | **~5000 рядків Go коду** |

---

## Метрики

- **Джерело:** 26 source files, 4932 lines of Go
- **Статей:** 40
- **Середня глибина:** ~120 рядків коду на статтю
- **Кожна стаття:** самодостатня, з фрагментами коду з репозиторію, Mermaid діаграмами
- **Репозиторій:** open-source, читач може запустити і експериментувати
