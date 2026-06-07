package shepherd

import (
	"time"
)

// --- Object Metadata ---

type ObjectMeta struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	UID       string            `json:"uid"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// --- Pod ---

type Pod struct {
	Kind     string    `json:"kind"`
	Metadata ObjectMeta `json:"metadata"`
	Spec     PodSpec    `json:"spec"`
	Status   PodStatus  `json:"status"`
}

type PodSpec struct {
	Containers    []ContainerSpec   `json:"containers"`
	NodeName      string            `json:"node_name,omitempty"`
	RestartPolicy RestartPolicy     `json:"restart_policy"`
	NodeSelector  map[string]string `json:"node_selector,omitempty"`
}

type ContainerSpec struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Command []string          `json:"command,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Ports   []PortSpec        `json:"ports,omitempty"`
	Resources ResourceSpec    `json:"resources,omitempty"`
}

type PortSpec struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"` // tcp, udp
}

type ResourceSpec struct {
	Memory int64 `json:"memory,omitempty"` // bytes
	CPU    int64 `json:"cpu,omitempty"`    // millicores
}

type RestartPolicy string

const (
	RestartAlways    RestartPolicy = "Always"
	RestartOnFailure RestartPolicy = "OnFailure"
	RestartNever     RestartPolicy = "Never"
)

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
)

type PodStatus struct {
	Phase      PodPhase          `json:"phase"`
	HostIP     string            `json:"host_ip,omitempty"`
	PodIP      string            `json:"pod_ip,omitempty"`
	StartTime  time.Time         `json:"start_time,omitempty"`
	Containers []ContainerStatus `json:"containers,omitempty"`
	Message    string            `json:"message,omitempty"`
}

type ContainerStatus struct {
	Name        string `json:"name"`
	ContainerID string `json:"container_id"`
	Ready       bool   `json:"ready"`
	State       string `json:"state"`
	ExitCode    int    `json:"exit_code,omitempty"`
}

// --- Service ---

type Service struct {
	Kind     string      `json:"kind"`
	Metadata ObjectMeta  `json:"metadata"`
	Spec     ServiceSpec `json:"spec"`
	Status   ServiceStatus `json:"status"`
}

type ServiceSpec struct {
	Selector map[string]string `json:"selector"`
	Ports    []ServicePort     `json:"ports"`
	Type     ServiceType       `json:"type"`
}

type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int    `json:"port"`
	TargetPort int    `json:"target_port"`
	NodePort   int    `json:"node_port,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
}

type ServiceType string

const (
	ServiceClusterIP ServiceType = "ClusterIP"
	ServiceNodePort  ServiceType = "NodePort"
)

type ServiceStatus struct {
	ClusterIP string   `json:"cluster_ip,omitempty"`
	Endpoints []string `json:"endpoints,omitempty"`
}

// --- Deployment ---

type Deployment struct {
	Kind     string         `json:"kind"`
	Metadata ObjectMeta     `json:"metadata"`
	Spec     DeploymentSpec `json:"spec"`
	Status   DeploymentStatus `json:"status"`
}

type DeploymentSpec struct {
	Replicas int               `json:"replicas"`
	Selector map[string]string `json:"selector"`
	Template PodTemplate       `json:"template"`
}

type PodTemplate struct {
	Metadata ObjectMeta `json:"metadata"`
	Spec     PodSpec    `json:"spec"`
}

type DeploymentStatus struct {
	Replicas          int `json:"replicas"`
	ReadyReplicas     int `json:"ready_replicas"`
	AvailableReplicas int `json:"available_replicas"`
	UpdatedReplicas   int `json:"updated_replicas"`
}

// --- Node ---

type Node struct {
	Kind     string     `json:"kind"`
	Metadata ObjectMeta `json:"metadata"`
	Spec     NodeSpec   `json:"spec"`
	Status   NodeStatus `json:"status"`
}

type NodeSpec struct {
	Address string `json:"address"` // host:port of the node agent
}

type NodeCondition string

const (
	NodeReady    NodeCondition = "Ready"
	NodeNotReady NodeCondition = "NotReady"
)

type NodeStatus struct {
	Condition    NodeCondition `json:"condition"`
	Capacity     NodeResources `json:"capacity"`
	Allocatable  NodeResources `json:"allocatable"`
	PodCount     int           `json:"pod_count"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
}

type NodeResources struct {
	CPU    int64 `json:"cpu"`    // millicores
	Memory int64 `json:"memory"` // bytes
	Pods   int   `json:"pods"`   // max pods
}

// --- Event ---

type Event struct {
	Type      string    `json:"type"` // Normal, Warning
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Object    string    `json:"object"` // "pod/name", "node/name"
	Timestamp time.Time `json:"timestamp"`
}
