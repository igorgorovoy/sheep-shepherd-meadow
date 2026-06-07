# Sheep Internals

Container runtime architecture, Linux isolation primitives, and networking.

## Table of Contents

- [Container Lifecycle](#container-lifecycle)
- [Linux Isolation Stack](#linux-isolation-stack)
- [Re-Exec Pattern](#re-exec-pattern)
- [Namespace Configuration](#namespace-configuration)
- [Cgroups v2 Resource Control](#cgroups-v2-resource-control)
- [Overlay Filesystem](#overlay-filesystem)
- [Container Networking](#container-networking)
- [Image Management](#image-management)
- [Filesystem Layout](#filesystem-layout)

## Container Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Created : sheep create
    Created --> Running : sheep start
    Running --> Stopped : sheep stop
    Running --> Stopped : process exits
    Stopped --> Running : sheep start
    Stopped --> [*] : sheep rm
    Created --> [*] : sheep rm
```

### Detailed Lifecycle Sequence

```mermaid
sequenceDiagram
    participant User
    participant CLI as sheep CLI
    participant Mgr as Container Manager
    participant Img as Image Manager
    participant OVL as OverlayFS
    participant RT as Runtime (Linux)
    participant Kernel

    User->>CLI: sheep run --name web minimal /bin/sh

    rect rgb(240, 248, 255)
        Note over CLI,OVL: Phase 1 - Create
        CLI->>Mgr: Create(RunOpts)
        Mgr->>Img: Get("minimal", "latest")
        Img-->>Mgr: Image with RootFS path
        Mgr->>OVL: setupOverlay(id, image.RootFS)
        OVL->>Kernel: mount overlay with lowerdir, upperdir, workdir
        OVL-->>Mgr: merged path
        Mgr->>Mgr: Save state.json
        Mgr-->>CLI: Container with State=created
    end

    rect rgb(240, 255, 240)
        Note over CLI,Kernel: Phase 2 - Start
        CLI->>Mgr: Start(containerID)
        Mgr->>RT: startContainer(container)
        RT->>RT: Build re-exec command
        RT->>Kernel: clone with NEWUTS, NEWPID, NEWNS, NEWIPC, NEWNET
        Kernel-->>RT: child PID

        par Cgroups Setup
            RT->>Kernel: write cgroup.procs, memory.max, pids.max, cpu.max
        and Network Setup
            RT->>Kernel: create veth pair, attach to bridge, move to netns
        end

        Note over RT,Kernel: Inside child namespace (init process)
        RT->>Kernel: sethostname, mount proc/sys/dev/tmp
        RT->>Kernel: create device nodes
        RT->>Kernel: pivot_root, unmount old root
        RT->>Kernel: exec(command)

        Mgr->>Mgr: State = Running, save PID
    end

    rect rgb(255, 240, 240)
        Note over CLI,Kernel: Phase 3 - Stop
        User->>CLI: sheep stop web
        CLI->>Mgr: Stop(containerID)
        Mgr->>RT: stopContainer(container)
        RT->>Kernel: SIGTERM then SIGKILL
        RT->>Kernel: Remove cgroup
        Mgr->>Mgr: State = Stopped, save exit code
    end
```

## Linux Isolation Stack

Each container runs in its own set of Linux namespaces with cgroup resource limits applied.

```mermaid
graph TB
    subgraph "Host"
        HOST_PID["Host PID namespace"]
        HOST_NET["Host network - eth0: 192.168.1.x"]
        HOST_MNT["Host mount namespace"]
        HOST_FS["Host filesystem /"]

        subgraph "Container - isolated"
            subgraph "Namespaces"
                PID["PID namespace - PID 1 = container process"]
                NET["NET namespace - eth0: 10.20.0.x"]
                MNT["MNT namespace - private mount tree"]
                UTS["UTS namespace - own hostname"]
                IPC["IPC namespace - isolated semaphores, shm"]
            end

            subgraph "Cgroups v2"
                CGMEM["memory.max = 268435456 (256M)"]
                CGCPU["cpu.max = 50000 100000 (50%)"]
                CGPID["pids.max = 100"]
            end

            subgraph "Filesystem"
                ROOT["/ (merged overlay)"]
                PROC["/proc (procfs)"]
                SYS["/sys (sysfs)"]
                DEV["/dev (tmpfs + nodes)"]
                TMP["/tmp (tmpfs)"]
            end
        end
    end

    HOST_PID -.->|"clone CLONE_NEWPID"| PID
    HOST_NET -.->|"clone CLONE_NEWNET"| NET
    HOST_MNT -.->|"clone CLONE_NEWNS"| MNT
    HOST_FS -.->|"pivot_root"| ROOT
```

### Namespace Isolation Matrix

| Namespace | Flag | Isolates | Effect |
|-----------|------|----------|--------|
| PID | `CLONE_NEWPID` | Process IDs | Container process is PID 1 |
| NET | `CLONE_NEWNET` | Network stack | Own interfaces, IPs, routing table |
| MNT | `CLONE_NEWNS` | Mount points | Private filesystem mount tree |
| UTS | `CLONE_NEWUTS` | Hostname | Own hostname and domain name |
| IPC | `CLONE_NEWIPC` | IPC primitives | Isolated semaphores, message queues, shared memory |

## Re-Exec Pattern

Go uses threads internally (goroutines mapped to OS threads). Linux `clone()` only copies the calling thread into the new namespace, leaving Go's runtime in an inconsistent state. The re-exec pattern solves this:

```mermaid
sequenceDiagram
    participant Parent as sheep (parent)
    participant Child as sheep init (child)
    participant Target as User Command

    Parent->>Parent: Build command: self "init" --rootfs /path -- /bin/sh
    Parent->>Child: clone with namespace flags

    Note over Child: Fresh process in new namespaces<br/>Go runtime initializes cleanly

    Child->>Child: Parse args (rootfs, hostname, command)
    Child->>Child: sethostname()
    Child->>Child: Mount /proc, /sys, /dev, /tmp
    Child->>Child: Create device nodes
    Child->>Child: pivot_root(rootfs, .pivot_old)
    Child->>Child: chdir("/")
    Child->>Child: umount(.pivot_old)
    Child->>Child: rmdir(.pivot_old)

    Child->>Target: syscall.Exec("/bin/sh")
    Note over Target: Now PID 1 inside the container
```

## Cgroups v2 Resource Control

```mermaid
graph TB
    subgraph "/sys/fs/cgroup"
        ROOT_CG["/ (root cgroup)"]

        subgraph "/sys/fs/cgroup/sheep"
            SHEEP_CG["sheep/<br/>cgroup.subtree_control:<br/>+memory +pids +cpu"]

            subgraph "Per-Container Cgroup"
                C1["sheep/CONTAINER_ID/"]
                C1_PROCS["cgroup.procs = PID"]
                C1_MEM["memory.max = 268435456"]
                C1_CPU["cpu.max = 50000 100000"]
                C1_PIDS["pids.max = 100"]
            end
        end
    end

    ROOT_CG --> SHEEP_CG
    SHEEP_CG --> C1
    C1 --> C1_PROCS
    C1 --> C1_MEM
    C1 --> C1_CPU
    C1 --> C1_PIDS
```

### Resource Limit Mapping

| Sheep Flag | Cgroup v2 File | Value Format | Example |
|-----------|---------------|-------------|---------|
| `-m 256m` | `memory.max` | bytes | `268435456` |
| `--cpu-quota 50000` | `cpu.max` | `$QUOTA $PERIOD` | `50000 100000` (50%) |
| `--cpu-shares 512` | `cpu.weight` | 1-10000 | `19` (mapped from Docker-style) |
| `--pids-limit 100` | `pids.max` | integer | `100` |

## Overlay Filesystem

OverlayFS provides copy-on-write layering: the image rootfs is read-only, and each container gets a writable layer on top.

```mermaid
graph TB
    subgraph "OverlayFS Mount"
        MERGED["Merged (container sees this)<br/>/var/lib/sheep/overlay/ID/merged"]
    end

    subgraph "Layers"
        UPPER["Upper - writable, container changes<br/>/var/lib/sheep/overlay/ID/upper"]
        LOWER["Lower - read-only, image rootfs<br/>/var/lib/sheep/images/IMG_ID/rootfs"]
        WORK["Work - overlay internals<br/>/var/lib/sheep/overlay/ID/work"]
    end

    UPPER -->|"modified and new files"| MERGED
    LOWER -->|"original files"| MERGED
    WORK -.->|"used by kernel"| MERGED
```

### Read/Write Flow

```mermaid
sequenceDiagram
    participant App as Container Process
    participant OVL as OverlayFS
    participant Upper as Upper Layer (writable)
    participant Lower as Lower Layer (image)

    Note over App,Lower: Reading an existing file
    App->>OVL: open("/etc/hosts")
    OVL->>Upper: Check upper layer
    Upper-->>OVL: Not found
    OVL->>Lower: Check lower layer
    Lower-->>OVL: Found - return file
    OVL-->>App: File contents

    Note over App,Lower: Writing to an existing file (copy-up)
    App->>OVL: write("/etc/hosts", data)
    OVL->>Lower: Copy file to upper layer
    OVL->>Upper: Write modified data
    Upper-->>OVL: OK
    OVL-->>App: Written

    Note over App,Lower: Creating a new file
    App->>OVL: create("/tmp/new.txt")
    OVL->>Upper: Create in upper layer
    Upper-->>OVL: OK
    OVL-->>App: Created
```

## Container Networking

### Bridge Network Architecture

```mermaid
graph TB
    subgraph "Host"
        ETH["eth0<br/>192.168.1.100"]
        IPTABLES["iptables NAT<br/>MASQUERADE"]
        BRIDGE["sheep0 (bridge)<br/>10.20.0.1/16"]
        IP_FWD["ip_forward = 1"]

        subgraph "Container 1"
            VETH1_G["eth0<br/>10.20.0.2/16"]
        end
        VETH1_H["veth_abc12345"]

        subgraph "Container 2"
            VETH2_G["eth0<br/>10.20.0.3/16"]
        end
        VETH2_H["veth_def67890"]

        subgraph "Container 3"
            VETH3_G["eth0<br/>10.20.0.4/16"]
        end
        VETH3_H["veth_ghi13579"]
    end

    INTERNET(("Internet"))

    VETH1_G ---|"veth pair"| VETH1_H
    VETH2_G ---|"veth pair"| VETH2_H
    VETH3_G ---|"veth pair"| VETH3_H

    VETH1_H --- BRIDGE
    VETH2_H --- BRIDGE
    VETH3_H --- BRIDGE

    BRIDGE --- IP_FWD
    IP_FWD --- IPTABLES
    IPTABLES --- ETH
    ETH --- INTERNET
```

### Network Setup Sequence

```mermaid
sequenceDiagram
    participant RT as Runtime
    participant Host as Host Kernel
    participant Bridge as sheep0 Bridge
    participant NS as Container Netns

    Note over RT,NS: First container: create bridge
    RT->>Host: ip link add sheep0 type bridge
    RT->>Host: ip addr add 10.20.0.1/16 dev sheep0
    RT->>Host: ip link set sheep0 up
    RT->>Host: enable ip_forward
    RT->>Host: iptables MASQUERADE for 10.20.0.0/16

    Note over RT,NS: Per container
    RT->>RT: Allocate IP: 10.20.0.2
    RT->>Host: ip link add veth_abc type veth peer name eth0
    RT->>Bridge: ip link set veth_abc master sheep0
    RT->>NS: ip link set eth0 netns PID
    RT->>Host: ip link set veth_abc up
    RT->>NS: nsenter: ip addr add 10.20.0.2/16 dev eth0
    RT->>NS: nsenter: ip link set eth0 up
    RT->>NS: nsenter: ip link set lo up
    RT->>NS: nsenter: ip route add default via 10.20.0.1
```

### Traffic Flow: Container to Internet

```mermaid
graph LR
    C_APP["App in Container<br/>10.20.0.2"] --> C_ETH0["eth0 (container)"]
    C_ETH0 --> V_HOST["veth (host end)"]
    V_HOST --> BR["sheep0 bridge"]
    BR --> NAT["MASQUERADE<br/>10.20.0.2 to 192.168.1.100"]
    NAT --> H_ETH0["eth0 (host)"]
    H_ETH0 --> INET["Internet"]
```

### Traffic Flow: Container to Container (same host)

```mermaid
graph LR
    C1_APP["Container 1<br/>10.20.0.2"] --> C1_ETH["eth0"]
    C1_ETH --> V1["veth1"]
    V1 --> BR2["sheep0 bridge"]
    BR2 --> V2["veth2"]
    V2 --> C2_ETH["eth0"]
    C2_ETH --> C2_APP["Container 2<br/>10.20.0.3"]
```

## Image Management

### Image Operations

```mermaid
graph TB
    subgraph "Image Sources"
        TAR["Tarball .tar.gz<br/>rootfs archive"]
        HOST["Host OS<br/>binary bootstrap"]
    end

    subgraph "Image Manager"
        IMPORT["Import<br/>extract tar to rootfs"]
        BOOT["Bootstrap<br/>copy /bin/sh, ls, etc."]
        LIST["List"]
        GET["Get by name:tag"]
        RM["Remove"]
    end

    subgraph "Image Store"
        subgraph "Image abc123"
            MANIFEST1["manifest.json"]
            ROOTFS1["rootfs/"]
        end
        subgraph "Image def456"
            MANIFEST2["manifest.json"]
            ROOTFS2["rootfs/"]
        end
    end

    TAR --> IMPORT
    HOST --> BOOT
    IMPORT --> MANIFEST1
    IMPORT --> ROOTFS1
    BOOT --> MANIFEST2
    BOOT --> ROOTFS2
```

## Filesystem Layout

```
/var/lib/sheep/
|-- containers/
|   +-- {container_id}/
|       +-- state.json              # Container metadata and state
|-- overlay/
|   +-- {container_id}/
|       |-- upper/                  # Writable layer
|       |-- work/                   # OverlayFS workdir
|       +-- merged/                 # Merged view (container rootfs)
|-- images/
|   +-- {image_id}/
|       |-- manifest.json           # Image metadata
|       +-- rootfs/                 # Image filesystem
|           |-- bin/
|           |-- etc/
|           |-- lib/
|           |-- proc/               # Mount point
|           |-- sys/                # Mount point
|           +-- ...
+-- network/
    +-- ip_counter                  # IP allocation state
```
