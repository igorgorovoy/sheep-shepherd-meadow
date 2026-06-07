# Shepherd Internals

Orchestrator architecture, scheduling algorithms, controller reconciliation loops, and cluster management.

## Table of Contents

- [Control Plane Architecture](#control-plane-architecture)
- [API Server](#api-server)
- [Scheduler](#scheduler)
- [Controller Manager](#controller-manager)
- [Node Agent](#node-agent)
- [Pod Lifecycle in Cluster](#pod-lifecycle-in-cluster)
- [Deployment Scaling Flow](#deployment-scaling-flow)
- [Health Monitoring](#health-monitoring)

## Control Plane Architecture

```mermaid
graph TB
    subgraph "Control Plane"
        API["API Server<br/>:9876<br/>REST/JSON"]
        SCHED["Scheduler<br/>2s reconcile loop"]

        subgraph "Controller Manager"
            RC["Replication Controller<br/>5s reconcile loop"]
            SC["Service Controller<br/>5s reconcile loop"]
            NC["Node Controller<br/>10s reconcile loop"]
        end

        STORE[("BoltDB Store<br/>shepherd.db")]
    end

    subgraph "Node Pool"
        A1["Agent (node-1)"]
        A2["Agent (node-2)"]
        A3["Agent (node-3)"]
    end

    CLI["sheepctl"] -->|HTTP| API
    API -->|read/write| STORE
    SCHED -->|read/write| STORE
    RC -->|read/write| STORE
    SC -->|read/write| STORE
    NC -->|read/write| STORE

    A1 -->|"heartbeat (10s)<br/>pod sync (3s)"| API
    A2 -->|"heartbeat (10s)<br/>pod sync (3s)"| API
    A3 -->|"heartbeat (10s)<br/>pod sync (3s)"| API
```

### Component Responsibilities

```mermaid
graph LR
    subgraph "API Server"
        direction TB
        AS1["CRUD for all resources"]
        AS2["Input validation"]
        AS3["Event recording"]
        AS4["Trigger scheduling"]
    end

    subgraph "Scheduler"
        direction TB
        S1["Watch pending pods"]
        S2["Filter feasible nodes"]
        S3["Score and rank nodes"]
        S4["Bind pod to node"]
    end

    subgraph "Replication Controller"
        direction TB
        R1["Watch deployments"]
        R2["Count matching pods"]
        R3["Scale up: create pods"]
        R4["Scale down: delete pods"]
        R5["Update deployment status"]
    end

    subgraph "Service Controller"
        direction TB
        SC1["Watch services"]
        SC2["Find matching pods"]
        SC3["Update endpoint list"]
    end

    subgraph "Node Controller"
        direction TB
        NC1["Monitor heartbeats"]
        NC2["Mark nodes NotReady"]
        NC3["Record warning events"]
    end
```

## API Server

### Endpoint Map

```mermaid
graph TB
    subgraph "API Routes"
        subgraph "Pods"
            P1["GET  /api/v1/pods"]
            P2["POST /api/v1/pods"]
            P3["GET  /api/v1/namespaces/NS/pods"]
            P4["GET  /api/v1/namespaces/NS/pods/NAME"]
            P5["PUT  /api/v1/namespaces/NS/pods/NAME"]
            P6["DELETE /api/v1/namespaces/NS/pods/NAME"]
        end

        subgraph "Services"
            S1["GET  /api/v1/services"]
            S2["POST /api/v1/services"]
            S3["GET  /api/v1/namespaces/NS/services/NAME"]
            S4["DELETE /api/v1/namespaces/NS/services/NAME"]
        end

        subgraph "Deployments"
            D1["GET  /api/v1/deployments"]
            D2["POST /api/v1/deployments"]
            D3["GET  /api/v1/namespaces/NS/deployments/NAME"]
            D4["PUT  /api/v1/namespaces/NS/deployments/NAME"]
            D5["DELETE /api/v1/namespaces/NS/deployments/NAME"]
        end

        subgraph "Nodes"
            N1["GET  /api/v1/nodes"]
            N2["POST /api/v1/nodes"]
            N3["GET  /api/v1/nodes/NAME"]
            N4["PUT  /api/v1/nodes/NAME"]
            N5["DELETE /api/v1/nodes/NAME"]
        end

        subgraph "System"
            H1["GET /healthz"]
            I1["GET /api/v1/info"]
            E1["GET /api/v1/events"]
        end
    end
```

### Request Flow

```mermaid
sequenceDiagram
    participant Client as sheepctl
    participant API as API Server
    participant Store as BoltDB
    participant Sched as Scheduler

    Client->>API: POST /api/v1/pods (JSON body)
    API->>API: Validate and set defaults
    API->>API: Generate UID, set namespace, timestamp
    API->>Store: CreatePod(pod)
    Store-->>API: OK
    API->>Store: RecordEvent("Created")
    API->>Sched: go SchedulePod(pod)
    API-->>Client: 201 Created (pod JSON)

    Note over Sched,Store: Async scheduling
    Sched->>Store: ListNodes()
    Sched->>Sched: Filter and score
    Sched->>Store: UpdatePod(pod with NodeName)
    Sched->>Store: RecordEvent("Scheduled")
```

## Scheduler

### Scheduling Algorithm

```mermaid
flowchart TB
    START["Pod enters Pending<br/>(NodeName empty)"] --> FILTER

    subgraph "Phase 1: Filter"
        FILTER["Filter feasible nodes"]
        F1{"Node Ready?"}
        F2{"Heartbeat < 30s?"}
        F3{"Labels match<br/>NodeSelector?"}
        F4{"Sufficient CPU?"}
        F5{"Sufficient Memory?"}
        F6{"Pod slots<br/>available?"}

        FILTER --> F1
        F1 -->|No| REJECT1["Reject node"]
        F1 -->|Yes| F2
        F2 -->|No| REJECT2["Reject node"]
        F2 -->|Yes| F3
        F3 -->|No| REJECT3["Reject node"]
        F3 -->|Yes| F4
        F4 -->|No| REJECT4["Reject node"]
        F4 -->|Yes| F5
        F5 -->|No| REJECT5["Reject node"]
        F5 -->|Yes| F6
        F6 -->|No| REJECT6["Reject node"]
        F6 -->|Yes| FEASIBLE["Add to feasible set"]
    end

    FEASIBLE --> SCORE

    subgraph "Phase 2: Score"
        SCORE["Score each feasible node"]
        S1["Least Loaded<br/>(allocatable.pods - pod_count) * 10"]
        S2["CPU Balance<br/>available_ratio * 50"]
        S3["Memory Balance<br/>available_ratio * 50"]
        SUM["Total Score = S1 + S2 + S3"]

        SCORE --> S1
        SCORE --> S2
        SCORE --> S3
        S1 --> SUM
        S2 --> SUM
        S3 --> SUM
    end

    SUM --> RANK["Rank by score<br/>(highest first)"]
    RANK --> BIND["Bind pod to top-ranked node"]
    BIND --> UPDATE["Update pod:<br/>spec.nodeName = selected node<br/>status.hostIP = node address"]
```

### Scheduling Example

```mermaid
graph TB
    POD["Pod: web-server<br/>Requests: CPU 250m, Mem 128MB"]

    subgraph "Node Filtering"
        N1["node-1<br/>Ready, 8 pods<br/>CPU: 4000m, Mem: 8GB"]
        N2["node-2<br/>NotReady<br/>CPU: 2000m, Mem: 4GB"]
        N3["node-3<br/>Ready, 2 pods<br/>CPU: 4000m, Mem: 8GB"]
        N4["node-4<br/>Ready, 110 pods (full)<br/>CPU: 8000m, Mem: 16GB"]
    end

    POD --> N1
    POD --> N2
    POD --> N3
    POD --> N4

    N2 -.->|"REJECTED: NotReady"| X2["X"]
    N4 -.->|"REJECTED: Pod limit"| X4["X"]

    N1 -->|"Score: 1020 + 47 + 49 = 1116"| RANK
    N3 -->|"Score: 1080 + 49 + 49 = 1178"| RANK

    RANK["Ranking"] --> WINNER["Winner: node-3<br/>Score 1178 > 1116"]

    style X2 fill:#FF6B6B
    style X4 fill:#FF6B6B
    style WINNER fill:#90EE90
```

## Controller Manager

### Reconciliation Loop Pattern

All controllers follow the same pattern: observe desired state, compare with actual state, take action to converge.

```mermaid
graph TB
    subgraph "Controller Loop (runs every N seconds)"
        OBSERVE["1. Observe<br/>Read desired state from store"]
        COMPARE["2. Compare<br/>Desired vs actual"]
        ACT["3. Act<br/>Create, update, or delete resources"]
        RECORD["4. Record<br/>Update status, emit events"]
    end

    OBSERVE --> COMPARE
    COMPARE --> ACT
    ACT --> RECORD
    RECORD -.->|"next tick"| OBSERVE
```

### Replication Controller Flow

```mermaid
sequenceDiagram
    participant RC as Replication Controller
    participant Store as BoltDB
    participant Sched as Scheduler

    loop Every 5 seconds
        RC->>Store: ListDeployments("")
        Store-->>RC: [deployment-web (replicas=3)]

        RC->>Store: ListPods("default")
        Store-->>RC: [web-0, web-1] (2 pods match selector)

        Note over RC: desired=3, actual=2, need 1 more

        RC->>Store: CreatePod("web-2")
        RC->>Store: RecordEvent("Created pod web-2")
        RC->>Sched: go SchedulePod(web-2)

        RC->>Store: UpdateDeployment(status: replicas=3, ready=2)
    end
```

### Scale Down Flow

```mermaid
sequenceDiagram
    participant User as sheepctl
    participant API as API Server
    participant Store as BoltDB
    participant RC as Replication Controller

    User->>API: PUT deployment/web replicas=1
    API->>Store: UpdateDeployment(replicas=1)

    Note over RC: Next reconcile tick (5s)
    RC->>Store: ListDeployments
    RC->>Store: ListPods matching selector
    Note over RC: desired=1, actual=3, need to remove 2

    RC->>Store: DeletePod("web-2")
    RC->>Store: DeletePod("web-1")
    RC->>Store: RecordEvent("Scaled down web from 3 to 1")
    RC->>Store: UpdateDeployment(status: replicas=1, ready=1)
```

### Service Controller Flow

```mermaid
sequenceDiagram
    participant SC as Service Controller
    participant Store as BoltDB

    loop Every 5 seconds
        SC->>Store: ListServices("")
        Store-->>SC: [web-service (selector: app=web)]

        SC->>Store: ListPods("default")
        Store-->>SC: [web-0 (Running, IP=10.20.0.2), web-1 (Running, IP=10.20.0.3)]

        Note over SC: Match selector "app=web" against pod labels

        SC->>Store: UpdateService(endpoints: ["10.20.0.2:8080", "10.20.0.3:8080"])
    end
```

## Node Agent

### Agent Lifecycle

```mermaid
sequenceDiagram
    participant Agent as Node Agent
    participant API as API Server
    participant Sheep as Sheep Runtime

    Note over Agent: Startup
    Agent->>Agent: Detect capacity (CPU, Memory)
    Agent->>API: POST /api/v1/nodes (register)
    API-->>Agent: 201 Created

    par Heartbeat Loop (10s)
        loop Every 10 seconds
            Agent->>Sheep: List running containers
            Agent->>API: PUT /api/v1/nodes/NAME<br/>(condition=Ready, podCount=N, lastHeartbeat=now)
        end
    and Pod Reconciliation Loop (3s)
        loop Every 3 seconds
            Agent->>API: GET /api/v1/pods
            API-->>Agent: Pods assigned to this node

            alt Pod is Pending
                Agent->>Sheep: Create container
                Agent->>Sheep: Start container
                Agent->>API: PUT pod status (Phase=Running)
            else Pod is Running
                Agent->>Sheep: Check container health
                alt Container stopped
                    Agent->>API: PUT pod status (Phase=Failed)
                end
            end
        end
    end
```

### Agent Pod Start Sequence

```mermaid
sequenceDiagram
    participant Agent
    participant Sheep as Container Manager
    participant API as API Server

    Agent->>Agent: Found pending pod "web-0" on my node

    loop For each container in pod spec
        Agent->>Agent: Convert pod spec to RunOpts
        Agent->>Sheep: Create(RunOpts)
        Sheep-->>Agent: Container ID

        Agent->>Sheep: Start(containerID)
        Sheep-->>Agent: OK / Error
    end

    alt All containers started
        Agent->>API: PUT pod (Phase=Running, containers=[...])
    else Some failed
        Agent->>API: PUT pod (Phase=Failed, message="...")
    end
```

## Pod Lifecycle in Cluster

End-to-end flow from `sheepctl apply` to running container.

```mermaid
sequenceDiagram
    participant User as sheepctl
    participant API as API Server
    participant Store as BoltDB
    participant Sched as Scheduler
    participant Agent as Node Agent
    participant Sheep as Sheep Runtime

    User->>API: POST /api/v1/pods
    API->>Store: CreatePod (Phase=Pending)
    API-->>User: 201 Created

    API->>Sched: SchedulePod(pod)
    Sched->>Store: ListNodes()
    Sched->>Sched: Filter: Ready + heartbeat + resources
    Sched->>Sched: Score: least-loaded + resource balance
    Sched->>Store: UpdatePod(nodeName=node-2)
    Sched->>Store: RecordEvent("Scheduled to node-2")

    Note over Agent: Reconcile loop detects new pod

    Agent->>API: GET /api/v1/pods
    Agent->>Agent: Found pending pod on my node

    Agent->>Sheep: Create(name, image, command, resources)
    Sheep->>Sheep: Setup overlay filesystem
    Sheep-->>Agent: Container created

    Agent->>Sheep: Start(containerID)
    Sheep->>Sheep: clone() with namespaces
    Sheep->>Sheep: Setup cgroups
    Sheep->>Sheep: Setup networking
    Sheep->>Sheep: pivot_root, exec command
    Sheep-->>Agent: PID

    Agent->>API: PUT pod (Phase=Running, containerID=xxx)
    API->>Store: UpdatePod

    Note over User: sheepctl get pods -> Phase=Running
```

## Deployment Scaling Flow

```mermaid
sequenceDiagram
    participant User as sheepctl
    participant API as API Server
    participant Store as BoltDB
    participant RC as Replication Controller
    participant Sched as Scheduler
    participant Agent1 as Agent (node-1)
    participant Agent2 as Agent (node-2)

    User->>API: sheepctl apply -f deployment.json (replicas=3)
    API->>Store: CreateDeployment

    Note over RC: Reconcile: desired=3, actual=0

    RC->>Store: CreatePod("web-0")
    RC->>Sched: SchedulePod("web-0")
    RC->>Store: CreatePod("web-1")
    RC->>Sched: SchedulePod("web-1")
    RC->>Store: CreatePod("web-2")
    RC->>Sched: SchedulePod("web-2")

    Sched->>Store: web-0 -> node-1
    Sched->>Store: web-1 -> node-2
    Sched->>Store: web-2 -> node-1

    Agent1->>Agent1: Start web-0, web-2
    Agent2->>Agent2: Start web-1

    Agent1->>API: web-0 Running
    Agent2->>API: web-1 Running
    Agent1->>API: web-2 Running

    Note over RC: Update deployment status: 3/3 ready

    User->>API: sheepctl scale deployment/web --replicas=5
    API->>Store: UpdateDeployment(replicas=5)

    Note over RC: Reconcile: desired=5, actual=3

    RC->>Store: CreatePod("web-3")
    RC->>Store: CreatePod("web-4")
    RC->>Sched: SchedulePod("web-3")
    RC->>Sched: SchedulePod("web-4")
```

## Health Monitoring

### Heartbeat and Node Health

```mermaid
stateDiagram-v2
    [*] --> Ready : Agent registers
    Ready --> Ready : Heartbeat received (< 30s)
    Ready --> NotReady : Heartbeat timeout (> 30s)
    NotReady --> Ready : Heartbeat received
    NotReady --> [*] : Node deleted

    note right of Ready
        Scheduler can assign pods
    end note
    note right of NotReady
        Scheduler skips this node
        Warning event emitted
    end note
```

### Health Check Timeline

```mermaid
gantt
    title Node Health Monitoring Timeline
    dateFormat ss
    axisFormat %Ss

    section Agent
    Heartbeat 1          :h1, 00, 1s
    Heartbeat 2          :h2, 10, 1s
    Heartbeat 3          :h3, 20, 1s
    Agent crashes        :crit, crash, 25, 1s

    section Node Controller
    Check (all ok)       :c1, 10, 1s
    Check (all ok)       :c2, 20, 1s
    Check (all ok)       :c3, 30, 1s
    Check (all ok)       :c4, 40, 1s
    Check (node NotReady):crit, c5, 56, 1s
```

### Event Flow

```mermaid
graph TB
    subgraph "Event Sources"
        API_EVT["API Server<br/>Pod created, deleted"]
        SCHED_EVT["Scheduler<br/>Pod scheduled"]
        RC_EVT["Replication Controller<br/>Pod created for deployment"]
        NC_EVT["Node Controller<br/>Node not ready"]
    end

    STORE_EVT[("Events Bucket<br/>BoltDB")]

    API_EVT --> STORE_EVT
    SCHED_EVT --> STORE_EVT
    RC_EVT --> STORE_EVT
    NC_EVT --> STORE_EVT

    STORE_EVT --> LIST_EVT["sheepctl events<br/>GET /api/v1/events"]
```

### Event Types

| Source | Type | Reason | Example Message |
|--------|------|--------|-----------------|
| API Server | Normal | Created | Pod web-0 created |
| API Server | Normal | Deleted | Pod web-0 deleted |
| Scheduler | Normal | Scheduled | Pod web-0 scheduled to node-1 |
| Replication Controller | Normal | Created | Created pod web-0 for deployment web |
| Node Controller | Warning | NodeNotReady | Node worker-2 heartbeat timeout |
