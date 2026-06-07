# Architecture Overview

## System Context

The Sheep/Shepherd platform consists of three user-facing binaries and two core subsystems.

```mermaid
C4Context
    title System Context — Sheep & Shepherd

    Person(user, "DevOps / Developer", "Manages containers and workloads")

    System(sheep, "Sheep", "Container runtime — creates, runs, and manages isolated Linux containers")
    System(shepherd, "Shepherd", "Orchestrator — schedules and manages containers across a cluster")
    System(sheepctl, "Sheepctl", "CLI client — sends commands to the Shepherd API")

    Rel(user, sheep, "runs containers directly")
    Rel(user, sheepctl, "manages cluster resources")
    Rel(sheepctl, shepherd, "REST API", "HTTP/JSON")
    Rel(shepherd, sheep, "manages container lifecycle", "Go API")
```

## Component Architecture

### High-Level Component View

```mermaid
graph TB
    subgraph "User Interface"
        CLI_SHEEP["sheep CLI"]
        CLI_SHEEPCTL["sheepctl CLI"]
    end

    subgraph "Shepherd Control Plane"
        API["API Server<br/><i>REST endpoints</i>"]
        SCHED["Scheduler<br/><i>filter + score</i>"]
        RC["Replication<br/>Controller"]
        SC["Service<br/>Controller"]
        NC["Node<br/>Controller"]
        STORE[("BoltDB<br/>Store")]
    end

    subgraph "Node Agent"
        AGENT["Agent<br/><i>heartbeat + reconcile</i>"]
    end

    subgraph "Sheep Runtime"
        MGR["Container<br/>Manager"]
        RT["Runtime<br/><i>namespaces + cgroups</i>"]
        IMG["Image<br/>Manager"]
        NET["Network<br/><i>bridge + veth</i>"]
        FS["Filesystem<br/><i>overlayfs</i>"]
    end

    subgraph "Linux Kernel"
        NS["Namespaces<br/>PID, NET, MNT, UTS, IPC"]
        CG["Cgroups v2<br/>memory, cpu, pids"]
        OVL["OverlayFS"]
        NETK["Netfilter / iptables"]
    end

    CLI_SHEEPCTL -->|HTTP| API
    CLI_SHEEP --> MGR
    API --> STORE
    API --> SCHED
    SCHED --> STORE
    RC --> STORE
    SC --> STORE
    NC --> STORE
    AGENT -->|HTTP| API
    AGENT --> MGR
    MGR --> RT
    MGR --> IMG
    MGR --> NET
    MGR --> FS
    RT --> NS
    RT --> CG
    FS --> OVL
    NET --> NETK
```

### Sheep Runtime Components

```mermaid
graph LR
    subgraph "Container Manager"
        CREATE["Create"]
        START["Start"]
        STOP["Stop"]
        REMOVE["Remove"]
        LIST["List / Get"]
    end

    subgraph "Image Manager"
        IMPORT["Import Tarball"]
        BOOTSTRAP["Bootstrap from Host"]
        IMG_LIST["List Images"]
        IMG_STORE[("Image Store<br/>/var/lib/sheep/images/")]
    end

    subgraph "Runtime (Linux)"
        CLONE["clone() with<br/>namespace flags"]
        REEXEC["Re-exec pattern<br/>(self → init)"]
        PIVOT["pivot_root()"]
        MOUNTS["Mount proc,<br/>sys, dev, tmp"]
        DEVICES["Create device<br/>nodes"]
        CGROUP["Cgroup setup<br/>memory, cpu, pids"]
    end

    subgraph "Overlay Filesystem"
        LOWER["Lower Layer<br/>(image rootfs)"]
        UPPER["Upper Layer<br/>(container writes)"]
        WORK["Work Dir"]
        MERGED["Merged View<br/>(container rootfs)"]
    end

    subgraph "Networking"
        BRIDGE["Bridge sheep0<br/>10.20.0.1/16"]
        VETH["veth pair"]
        IPALLOC["IP Allocator"]
        NAT["iptables NAT<br/>MASQUERADE"]
    end

    CREATE --> IMG_STORE
    CREATE --> MERGED
    START --> CLONE
    CLONE --> REEXEC
    REEXEC --> PIVOT
    REEXEC --> MOUNTS
    REEXEC --> DEVICES
    START --> CGROUP
    START --> VETH

    LOWER --> MERGED
    UPPER --> MERGED
```

### Shepherd Control Plane Components

```mermaid
graph TB
    subgraph "API Server"
        PODS_EP["/api/v1/pods"]
        SVC_EP["/api/v1/services"]
        DEP_EP["/api/v1/deployments"]
        NODE_EP["/api/v1/nodes"]
        EVT_EP["/api/v1/events"]
    end

    subgraph "Scheduler"
        WATCH_P["Watch pending pods"]
        FILTER["Filter feasible nodes<br/><i>ready? heartbeat? resources?<br/>node selector?</i>"]
        SCORE["Score nodes<br/><i>least-loaded, resource balance</i>"]
        BIND["Bind pod → node"]
    end

    subgraph "Controllers"
        subgraph "Replication Controller"
            RC_WATCH["Watch deployments"]
            RC_COMPARE["Compare desired vs actual"]
            RC_SCALE["Create / delete pods"]
        end
        subgraph "Service Controller"
            SC_WATCH["Watch services"]
            SC_ENDPOINTS["Update endpoint list"]
        end
        subgraph "Node Controller"
            NC_WATCH["Monitor heartbeats"]
            NC_MARK["Mark NotReady"]
        end
    end

    subgraph "Store (BoltDB)"
        B_PODS[("pods")]
        B_SVC[("services")]
        B_DEP[("deployments")]
        B_NODES[("nodes")]
        B_EVT[("events")]
    end

    PODS_EP --> B_PODS
    SVC_EP --> B_SVC
    DEP_EP --> B_DEP
    NODE_EP --> B_NODES
    EVT_EP --> B_EVT

    WATCH_P --> B_PODS
    FILTER --> B_NODES
    BIND --> B_PODS

    RC_WATCH --> B_DEP
    RC_COMPARE --> B_PODS
    RC_SCALE --> B_PODS

    SC_WATCH --> B_SVC
    SC_ENDPOINTS --> B_PODS

    NC_WATCH --> B_NODES
```

## Deployment Topologies

### Standalone (Single Node)

```mermaid
graph TB
    subgraph "Single Host"
        SHEP["Shepherd<br/>(--mode standalone)"]
        subgraph "Embedded"
            API2["API Server"]
            SCHED2["Scheduler"]
            CTRL2["Controllers"]
            AGENT2["Agent"]
            SHEEP2["Sheep Runtime"]
        end
        SHEP --> API2
        SHEP --> SCHED2
        SHEP --> CTRL2
        SHEP --> AGENT2
        AGENT2 --> SHEEP2
    end

    SHEEPCTL2["sheepctl"] -->|":9876"| API2
```

### Multi-Node Cluster

```mermaid
graph TB
    subgraph "Control Plane Node"
        CP_API["API Server :9876"]
        CP_SCHED["Scheduler"]
        CP_CTRL["Controllers"]
        CP_DB[("shepherd.db")]
    end

    subgraph "Worker Node 1"
        W1_AGENT["Agent"]
        W1_SHEEP["Sheep Runtime"]
        W1_C1["Container A"]
        W1_C2["Container B"]
        W1_SHEEP --> W1_C1
        W1_SHEEP --> W1_C2
    end

    subgraph "Worker Node 2"
        W2_AGENT["Agent"]
        W2_SHEEP["Sheep Runtime"]
        W2_C1["Container C"]
        W2_C2["Container D"]
        W2_SHEEP --> W2_C1
        W2_SHEEP --> W2_C2
    end

    subgraph "Worker Node 3"
        W3_AGENT["Agent"]
        W3_SHEEP["Sheep Runtime"]
        W3_C1["Container E"]
        W3_SHEEP --> W3_C1
    end

    SHEEPCTL3["sheepctl"] -->|HTTP| CP_API
    CP_API --> CP_DB
    CP_SCHED --> CP_DB
    CP_CTRL --> CP_DB

    W1_AGENT -->|"heartbeat<br/>pod sync"| CP_API
    W2_AGENT -->|"heartbeat<br/>pod sync"| CP_API
    W3_AGENT -->|"heartbeat<br/>pod sync"| CP_API

    W1_AGENT --> W1_SHEEP
    W2_AGENT --> W2_SHEEP
    W3_AGENT --> W3_SHEEP
```

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go 1.23 | System programming, concurrency, static binaries |
| Container isolation | Linux namespaces + cgroups v2 | Direct kernel primitives, no shim |
| Filesystem isolation | OverlayFS | Copy-on-write, standard in container runtimes |
| State store | BoltDB | Embedded, no external dependencies, transactional |
| API protocol | REST/JSON | Simple, debuggable, standard tooling |
| Build constraints | `//go:build linux` | Compile everywhere, run on Linux |
| Process isolation | Re-exec pattern | Required for Go's threading model with clone() |
| Networking | Bridge + veth + NAT | Standard container networking model |

## Comparison with Industry Standards

| Component | Sheep/Shepherd | Docker/Kubernetes |
|-----------|----------------|-------------------|
| Container runtime | `sheep` | `containerd` / `runc` |
| Image format | tar/gzip rootfs | OCI image spec |
| Orchestrator API | REST on `:9876` | kube-apiserver on `:6443` |
| State store | BoltDB (embedded) | etcd (distributed) |
| Scheduler | filter + score | predicates + priorities |
| Controllers | Replication, Service, Node | ReplicaSet, Service, Node, ... |
| Node agent | Shepherd agent | kubelet |
| CLI | `sheepctl` | `kubectl` |
| Networking | bridge + veth | CNI plugins |
| Service mesh | — | Istio, Linkerd, ... |
