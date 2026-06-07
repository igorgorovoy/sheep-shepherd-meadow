# Data Model

Entity relationships, storage schema, state transitions, and data flows.

## Table of Contents

- [Entity Relationship Diagram](#entity-relationship-diagram)
- [Core Entities](#core-entities)
- [BoltDB Storage Schema](#boltdb-storage-schema)
- [Data Flow Diagrams](#data-flow-diagrams)
- [State Machine Diagrams](#state-machine-diagrams)
- [Container State Schema](#container-state-schema)

## Entity Relationship Diagram

```mermaid
erDiagram
    DEPLOYMENT ||--o{ POD : "manages via replication controller"
    DEPLOYMENT {
        string name PK
        string namespace
        string uid
        int replicas
        map selector
        PodTemplate template
        DeploymentStatus status
    }

    POD ||--|{ CONTAINER_STATUS : contains
    POD {
        string name PK
        string namespace
        string uid
        map labels
        PodSpec spec
        PodPhase phase
        string node_name FK
        string pod_ip
        string host_ip
    }

    CONTAINER_STATUS {
        string name
        string container_id FK
        bool ready
        string state
        int exit_code
    }

    SERVICE ||--o{ POD : "selects via labels"
    SERVICE {
        string name PK
        string namespace
        string uid
        map selector
        ServicePort[] ports
        ServiceType type
        string cluster_ip
        string[] endpoints
    }

    NODE ||--o{ POD : "runs"
    NODE {
        string name PK
        string uid
        map labels
        string address
        NodeCondition condition
        NodeResources capacity
        NodeResources allocatable
        int pod_count
        datetime last_heartbeat
    }

    EVENT {
        string type
        string reason
        string message
        string object FK
        datetime timestamp
    }

    POD }o--|| NODE : "scheduled on"
    SERVICE }o--o{ POD : "routes to matching"
    EVENT }o--|| POD : "about"
    EVENT }o--|| NODE : "about"
    EVENT }o--|| DEPLOYMENT : "about"
```

## Core Entities

### Pod

The smallest schedulable unit. Contains one or more container specifications.

```mermaid
classDiagram
    class Pod {
        +string Kind = "Pod"
        +ObjectMeta Metadata
        +PodSpec Spec
        +PodStatus Status
    }

    class ObjectMeta {
        +string Name
        +string Namespace
        +string UID
        +map~string,string~ Labels
        +time.Time CreatedAt
    }

    class PodSpec {
        +ContainerSpec[] Containers
        +string NodeName
        +RestartPolicy RestartPolicy
        +map~string,string~ NodeSelector
    }

    class ContainerSpec {
        +string Name
        +string Image
        +string[] Command
        +map~string,string~ Env
        +PortSpec[] Ports
        +ResourceSpec Resources
    }

    class ResourceSpec {
        +int64 Memory
        +int64 CPU
    }

    class PodStatus {
        +PodPhase Phase
        +string HostIP
        +string PodIP
        +time.Time StartTime
        +ContainerStatus[] Containers
        +string Message
    }

    class ContainerStatus {
        +string Name
        +string ContainerID
        +bool Ready
        +string State
        +int ExitCode
    }

    Pod --> ObjectMeta
    Pod --> PodSpec
    Pod --> PodStatus
    PodSpec --> ContainerSpec
    ContainerSpec --> ResourceSpec
    PodStatus --> ContainerStatus
```

### Deployment

Manages a set of identical pod replicas.

```mermaid
classDiagram
    class Deployment {
        +string Kind = "Deployment"
        +ObjectMeta Metadata
        +DeploymentSpec Spec
        +DeploymentStatus Status
    }

    class DeploymentSpec {
        +int Replicas
        +map~string,string~ Selector
        +PodTemplate Template
    }

    class PodTemplate {
        +ObjectMeta Metadata
        +PodSpec Spec
    }

    class DeploymentStatus {
        +int Replicas
        +int ReadyReplicas
        +int AvailableReplicas
        +int UpdatedReplicas
    }

    Deployment --> DeploymentSpec
    Deployment --> DeploymentStatus
    DeploymentSpec --> PodTemplate
    PodTemplate --> PodSpec
```

### Service

Routes traffic to a set of pods matched by label selector.

```mermaid
classDiagram
    class Service {
        +string Kind = "Service"
        +ObjectMeta Metadata
        +ServiceSpec Spec
        +ServiceStatus Status
    }

    class ServiceSpec {
        +map~string,string~ Selector
        +ServicePort[] Ports
        +ServiceType Type
    }

    class ServicePort {
        +string Name
        +int Port
        +int TargetPort
        +int NodePort
        +string Protocol
    }

    class ServiceStatus {
        +string ClusterIP
        +string[] Endpoints
    }

    Service --> ServiceSpec
    Service --> ServiceStatus
    ServiceSpec --> ServicePort
```

### Node

Represents a worker machine in the cluster.

```mermaid
classDiagram
    class Node {
        +string Kind = "Node"
        +ObjectMeta Metadata
        +NodeSpec Spec
        +NodeStatus Status
    }

    class NodeSpec {
        +string Address
    }

    class NodeStatus {
        +NodeCondition Condition
        +NodeResources Capacity
        +NodeResources Allocatable
        +int PodCount
        +time.Time LastHeartbeat
    }

    class NodeResources {
        +int64 CPU (millicores)
        +int64 Memory (bytes)
        +int Pods (max count)
    }

    Node --> NodeSpec
    Node --> NodeStatus
    NodeStatus --> NodeResources
```

## BoltDB Storage Schema

BoltDB is an embedded key-value store organized into buckets (similar to tables).

```mermaid
graph TB
    subgraph "shepherd.db"
        subgraph "Bucket: pods"
            PK1["Key: default/web-0<br/>Value: Pod JSON"]
            PK2["Key: default/web-1<br/>Value: Pod JSON"]
            PK3["Key: staging/api-0<br/>Value: Pod JSON"]
        end

        subgraph "Bucket: services"
            SK1["Key: default/web-service<br/>Value: Service JSON"]
            SK2["Key: default/api-service<br/>Value: Service JSON"]
        end

        subgraph "Bucket: deployments"
            DK1["Key: default/web<br/>Value: Deployment JSON"]
            DK2["Key: staging/api<br/>Value: Deployment JSON"]
        end

        subgraph "Bucket: nodes"
            NK1["Key: node-1<br/>Value: Node JSON"]
            NK2["Key: node-2<br/>Value: Node JSON"]
        end

        subgraph "Bucket: events"
            EK1["Key: 1714300000-pod/web-0<br/>Value: Event JSON"]
            EK2["Key: 1714300001-node/node-1<br/>Value: Event JSON"]
        end
    end
```

### Key Schema

| Bucket | Key Format | Example |
|--------|-----------|---------|
| `pods` | `{namespace}/{name}` | `default/web-0` |
| `services` | `{namespace}/{name}` | `default/web-service` |
| `deployments` | `{namespace}/{name}` | `default/web` |
| `nodes` | `{name}` | `node-1` |
| `events` | `{timestamp_ns}-{object}` | `1714300000-pod/web-0` |

### Storage Operations

```mermaid
graph LR
    subgraph "Write Path"
        W1["API Server receives request"]
        W2["Serialize to JSON"]
        W3["BoltDB Update transaction"]
        W4["Write to bucket"]
        W1 --> W2 --> W3 --> W4
    end

    subgraph "Read Path"
        R1["API Server receives request"]
        R2["BoltDB View transaction"]
        R3["Get by key or ForEach"]
        R4["Deserialize JSON"]
        R1 --> R2 --> R3 --> R4
    end

    subgraph "List with Namespace Filter"
        L1["ForEach in bucket"]
        L2["Check key prefix matches namespace/"]
        L3["Deserialize matching entries"]
        L1 --> L2 --> L3
    end
```

## Data Flow Diagrams

### Create Deployment Data Flow

```mermaid
graph TB
    INPUT["sheepctl apply -f deployment.json"]

    subgraph "JSON Input"
        JSON["kind: Deployment<br/>name: web<br/>replicas: 3<br/>selector: app=web<br/>template: ..."]
    end

    subgraph "API Server"
        VALIDATE["Validate + defaults<br/>- Set Kind=Deployment<br/>- Set namespace=default<br/>- Generate UID<br/>- Set CreatedAt"]
    end

    subgraph "BoltDB"
        DEP_WRITE["Write to deployments bucket<br/>Key: default/web"]
    end

    subgraph "Replication Controller"
        RC_READ["Read deployment: replicas=3"]
        RC_COUNT["Count matching pods: 0"]
        RC_CREATE["Create 3 pods:<br/>web-0, web-1, web-2"]
    end

    subgraph "Scheduler"
        SCHED_READ["Read pending pods"]
        SCHED_ASSIGN["Assign to nodes:<br/>web-0 -> node-1<br/>web-1 -> node-2<br/>web-2 -> node-1"]
    end

    subgraph "Node Agents"
        A1_READ["Agent node-1: read pods"]
        A1_START["Start web-0, web-2"]
        A2_READ["Agent node-2: read pods"]
        A2_START["Start web-1"]
    end

    subgraph "Status Updates"
        STATUS["Update pod statuses:<br/>Phase=Running<br/>ContainerIDs assigned"]
        DEP_STATUS["Update deployment status:<br/>replicas=3<br/>readyReplicas=3"]
    end

    INPUT --> JSON
    JSON --> VALIDATE
    VALIDATE --> DEP_WRITE
    DEP_WRITE --> RC_READ
    RC_READ --> RC_COUNT
    RC_COUNT --> RC_CREATE
    RC_CREATE --> SCHED_READ
    SCHED_READ --> SCHED_ASSIGN
    SCHED_ASSIGN --> A1_READ
    SCHED_ASSIGN --> A2_READ
    A1_READ --> A1_START
    A2_READ --> A2_START
    A1_START --> STATUS
    A2_START --> STATUS
    STATUS --> DEP_STATUS
```

### Service Discovery Data Flow

```mermaid
graph TB
    subgraph "Service Definition"
        SVC["Service: web-service<br/>selector: app=web<br/>port: 80 -> 8080"]
    end

    subgraph "Running Pods"
        P1["web-0<br/>labels: app=web<br/>IP: 10.20.0.2<br/>Phase: Running"]
        P2["web-1<br/>labels: app=web<br/>IP: 10.20.0.3<br/>Phase: Running"]
        P3["web-2<br/>labels: app=web<br/>IP: 10.20.0.4<br/>Phase: Failed"]
        P4["api-0<br/>labels: app=api<br/>IP: 10.20.0.5<br/>Phase: Running"]
    end

    subgraph "Service Controller"
        MATCH["Match selector 'app=web'<br/>against pod labels"]
        FILTER_RUNNING["Filter: Phase=Running only"]
        BUILD["Build endpoints:<br/>10.20.0.2:8080<br/>10.20.0.3:8080"]
    end

    subgraph "Updated Service"
        SVC_OUT["Service: web-service<br/>endpoints:<br/>- 10.20.0.2:8080<br/>- 10.20.0.3:8080"]
    end

    SVC --> MATCH
    P1 --> MATCH
    P2 --> MATCH
    P3 --> MATCH
    P4 --> MATCH
    MATCH --> FILTER_RUNNING
    FILTER_RUNNING --> BUILD
    BUILD --> SVC_OUT

    P3 -.->|"excluded: not Running"| FILTER_RUNNING
    P4 -.->|"excluded: labels don't match"| MATCH
```

### Node Agent Heartbeat Data Flow

```mermaid
graph LR
    subgraph "Node Agent"
        DETECT["Detect system resources<br/>CPU: 4 cores = 4000m<br/>Memory: 8GB<br/>Max pods: 110"]
        COUNT["Count running containers"]
        BUILD_HB["Build heartbeat:<br/>condition=Ready<br/>podCount=5<br/>lastHeartbeat=now"]
    end

    subgraph "API Server"
        RECEIVE["PUT /api/v1/nodes/node-1"]
        STORE_NODE["Update node in BoltDB"]
    end

    subgraph "Node Controller"
        CHECK["Check all nodes"]
        TIMEOUT{"heartbeat<br/>age > 30s?"}
        MARK_OK["Keep Ready"]
        MARK_BAD["Mark NotReady<br/>+ Warning event"]
    end

    DETECT --> COUNT
    COUNT --> BUILD_HB
    BUILD_HB -->|"HTTP PUT"| RECEIVE
    RECEIVE --> STORE_NODE
    STORE_NODE --> CHECK
    CHECK --> TIMEOUT
    TIMEOUT -->|No| MARK_OK
    TIMEOUT -->|Yes| MARK_BAD
```

## State Machine Diagrams

### Pod Phase Transitions

```mermaid
stateDiagram-v2
    [*] --> Pending : Created via API
    Pending --> Pending : Scheduled (NodeName assigned)
    Pending --> Running : Agent starts all containers
    Pending --> Failed : Agent fails to start containers
    Running --> Failed : Container crashes
    Running --> Succeeded : Process exits 0

    note right of Pending
        No NodeName: waiting for scheduler
        Has NodeName: waiting for agent
    end note

    note right of Running
        All containers started
        Agent monitors health
    end note
```

### Node Condition Transitions

```mermaid
stateDiagram-v2
    [*] --> Ready : Agent registers
    Ready --> Ready : Heartbeat (< 30s)
    Ready --> NotReady : Timeout (> 30s)
    NotReady --> Ready : Heartbeat received
    NotReady --> [*] : Deleted

    note right of Ready
        Can receive new pods
        Scheduler considers this node
    end note

    note right of NotReady
        Scheduler skips this node
        Warning event recorded
    end note
```

### Container State Transitions (Sheep Runtime)

```mermaid
stateDiagram-v2
    [*] --> created : Manager.Create()
    created --> running : Manager.Start()
    running --> stopped : Manager.Stop()
    running --> stopped : Process exits
    stopped --> running : Manager.Start()
    stopped --> [*] : Manager.Remove()
    created --> [*] : Manager.Remove()

    note right of created
        Overlay mounted
        State persisted
        No process running
    end note

    note right of running
        Namespaces active
        Cgroups applied
        Network configured
        PID tracked
    end note

    note right of stopped
        Process terminated
        Cgroups cleaned
        Overlay still mounted
        Exit code recorded
    end note
```

## Container State Schema

### state.json (per container)

```mermaid
graph TB
    subgraph "state.json"
        ID["id: string (32 hex chars)"]
        NAME["name: string"]
        IMAGE["image: string"]
        CMD["command: string[]"]
        STATE["state: created | running | stopped"]
        PID["pid: int"]
        EXIT["exit_code: int"]
        TIMES["created_at / started_at / stopped_at"]

        subgraph "config"
            HOSTNAME["hostname: string"]
            ENV["env: string[]"]
            WORKDIR["work_dir: string"]
            MEM["memory: int64 (bytes)"]
            CPU_S["cpu_shares: int64"]
            CPU_Q["cpu_quota: int64"]
            PIDS["pids_limit: int64"]
        end

        subgraph "network"
            IP["ip_address: 10.20.x.x"]
            GW["gateway: 10.20.0.1"]
            BR["bridge: sheep0"]
            VH["veth_host: veth_xxx"]
            VG["veth_guest: eth0"]
        end

        subgraph "mounts"
            SRC["source: /host/path"]
            TGT["target: /container/path"]
            RO["readonly: bool"]
        end
    end
```

### manifest.json (per image)

```mermaid
graph TB
    subgraph "manifest.json"
        IMG_ID["id: string (32 hex chars)"]
        IMG_NAME["name: string"]
        IMG_TAG["tag: string (default: latest)"]
        IMG_SIZE["size: int64 (bytes)"]
        IMG_CREATED["created_at: RFC3339"]
        IMG_ROOTFS["rootfs: /var/lib/sheep/images/ID/rootfs"]
    end
```
