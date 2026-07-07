package shepherd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"sheep/internal/container"
)

// Agent runs on each node and manages containers via the sheep runtime.
// It registers with the API server, sends heartbeats, and reconciles pods.
type Agent struct {
	nodeName  string
	apiAddr   string // shepherd API server address
	mgr       *container.Manager
	logger    *log.Logger
	capacity  NodeResources
}

func NewAgent(nodeName, apiAddr string, logger *log.Logger) *Agent {
	return &Agent{
		nodeName: nodeName,
		apiAddr:  apiAddr,
		mgr:      container.NewManager(os.Getenv("SHEEP_DATA_DIR")),
		logger:   logger,
		capacity: detectCapacity(),
	}
}

func (a *Agent) Run(stopCh <-chan struct{}) error {
	if err := a.mgr.Init(); err != nil {
		return fmt.Errorf("init container manager: %w", err)
	}

	// Register node with API server (retry for standalone mode race condition)
	for attempt := 0; attempt < 10; attempt++ {
		if err := a.register(); err == nil {
			break
		} else if attempt == 9 {
			return fmt.Errorf("register node after 10 attempts: %w", err)
		} else {
			a.logger.Printf("agent: register attempt %d failed, retrying...", attempt+1)
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}

	a.logger.Printf("agent %s started, reporting to %s", a.nodeName, a.apiAddr)

	// Start heartbeat
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()

	// Start pod reconciliation
	reconcile := time.NewTicker(3 * time.Second)
	defer reconcile.Stop()

	for {
		select {
		case <-stopCh:
			a.logger.Println("agent stopped")
			return nil
		case <-heartbeat.C:
			a.sendHeartbeat()
		case <-reconcile.C:
			a.reconcilePods()
		}
	}
}

func (a *Agent) register() error {
	hostname, _ := os.Hostname()
	if a.nodeName == "" {
		a.nodeName = hostname
	}

	node := Node{
		Kind: "Node",
		Metadata: ObjectMeta{
			Name:      a.nodeName,
			CreatedAt: time.Now(),
			Labels: map[string]string{
				"hostname": hostname,
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
			},
		},
		Spec: NodeSpec{
			Address: a.nodeName,
		},
		Status: NodeStatus{
			Condition:     NodeReady,
			Capacity:      a.capacity,
			Allocatable:   a.capacity,
			LastHeartbeat: time.Now(),
		},
	}

	return a.post("/api/v1/nodes", node)
}

func (a *Agent) sendHeartbeat() {
	node, err := a.getNode()
	if err != nil {
		a.logger.Printf("agent: get node error: %v", err)
		return
	}

	// Update pod count
	containers := a.mgr.List(false)
	node.Status.PodCount = len(containers)
	node.Status.LastHeartbeat = time.Now()
	node.Status.Condition = NodeReady

	if err := a.put("/api/v1/nodes/"+a.nodeName, node); err != nil {
		a.logger.Printf("agent: heartbeat error: %v", err)
	}
}

func (a *Agent) reconcilePods() {
	// Get pods assigned to this node
	pods, err := a.listMyPods()
	if err != nil {
		return
	}

	podNamesOnNode := make(map[string]struct{})

	for _, pod := range pods {
		if pod.Spec.NodeName != a.nodeName {
			continue
		}

		podNamesOnNode[pod.Metadata.Name] = struct{}{}

		switch pod.Status.Phase {
		case PodPending:
			// Start containers for pending pods
			a.startPod(pod)
		case PodRunning:
			// Check health of running pods
			a.checkPod(pod)
		}
	}

	a.stopOrphanPodContainers(podNamesOnNode)
}

// stopOrphanPodContainers stops containers whose pod was removed from the API
// (e.g. deployment cascade delete). Agent-managed pods set Config.Hostname to
// the pod name; manually created containers use a random hostname.
func (a *Agent) stopOrphanPodContainers(activePods map[string]struct{}) {
	for _, c := range a.mgr.List(true) {
		podName := c.Config.Hostname
		if podName == "" {
			continue
		}
		if _, ok := activePods[podName]; ok {
			continue
		}

		a.logger.Printf("agent: stopping orphan container %s (pod %s removed)", c.Name, podName)
		if err := a.mgr.Stop(c.ID); err != nil {
			a.logger.Printf("agent: stop orphan container %s: %v", c.Name, err)
		}
		if err := a.mgr.Remove(c.ID); err != nil {
			a.logger.Printf("agent: remove orphan container %s: %v", c.Name, err)
		}
	}
}

func (a *Agent) startPod(pod *Pod) {
	a.logger.Printf("agent: starting pod %s", pod.Metadata.Name)

	var statuses []ContainerStatus

	for _, cs := range pod.Spec.Containers {
		// Convert env map to slice
		var env []string
		for k, v := range cs.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}

		opts := container.RunOpts{
			Name:    fmt.Sprintf("%s-%s", pod.Metadata.Name, cs.Name),
			Image:   cs.Image,
			Command: cs.Command,
			Config: container.Config{
				Hostname: pod.Metadata.Name,
				Env:      env,
				Memory:   cs.Resources.Memory,
			},
		}

		c, err := a.mgr.Create(opts)
		if err != nil {
			a.logger.Printf("agent: create container error: %v", err)
			statuses = append(statuses, ContainerStatus{
				Name:  cs.Name,
				State: "failed",
			})
			continue
		}

		if err := a.mgr.Start(c.ID); err != nil {
			a.logger.Printf("agent: start container error: %v", err)
			statuses = append(statuses, ContainerStatus{
				Name:        cs.Name,
				ContainerID: c.ID,
				State:       "failed",
			})
			continue
		}

		statuses = append(statuses, ContainerStatus{
			Name:        cs.Name,
			ContainerID: c.ID,
			Ready:       true,
			State:       "running",
		})
	}

	// Update pod status
	allReady := true
	for _, s := range statuses {
		if !s.Ready {
			allReady = false
			break
		}
	}

	pod.Status.Containers = statuses
	pod.Status.StartTime = time.Now()

	// Set PodIP from container network, or 127.0.0.1 for host-mode
	if pod.Status.PodIP == "" {
		for _, cs := range statuses {
			if cs.ContainerID != "" {
				c, err := a.mgr.Get(cs.ContainerID)
				if err == nil && c.Network != nil && c.Network.IPAddress != "" {
					pod.Status.PodIP = c.Network.IPAddress
					break
				}
			}
		}
		if pod.Status.PodIP == "" {
			pod.Status.PodIP = "127.0.0.1"
		}
	}

	if allReady {
		pod.Status.Phase = PodRunning
		pod.Status.Message = "all containers running"
	} else {
		pod.Status.Phase = PodFailed
		pod.Status.Message = "one or more containers failed to start"
	}

	a.updatePodStatus(pod)
}

func (a *Agent) checkPod(pod *Pod) {
	for i, cs := range pod.Status.Containers {
		if cs.ContainerID == "" {
			continue
		}
		c, err := a.mgr.Get(cs.ContainerID)
		if err != nil {
			pod.Status.Containers[i].Ready = false
			pod.Status.Containers[i].State = "unknown"
			continue
		}
		pod.Status.Containers[i].Ready = c.State == container.StateRunning
		pod.Status.Containers[i].State = string(c.State)
	}

	// Check if pod is still healthy
	allRunning := true
	for _, cs := range pod.Status.Containers {
		if !cs.Ready {
			allRunning = false
			break
		}
	}

	if !allRunning && pod.Status.Phase == PodRunning {
		pod.Status.Phase = PodFailed
		pod.Status.Message = "one or more containers stopped"
		a.updatePodStatus(pod)
	}
}

// HTTP helpers for communicating with the API server

func (a *Agent) post(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://"+a.apiAddr+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (a *Agent) put(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, "http://"+a.apiAddr+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (a *Agent) getNode() (*Node, error) {
	resp, err := http.Get("http://" + a.apiAddr + "/api/v1/nodes/" + a.nodeName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}
	return &node, nil
}

func (a *Agent) listMyPods() ([]*Pod, error) {
	resp, err := http.Get("http://" + a.apiAddr + "/api/v1/pods")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pods []*Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		return nil, err
	}
	return pods, nil
}

func (a *Agent) updatePodStatus(pod *Pod) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", pod.Metadata.Namespace, pod.Metadata.Name)
	if err := a.put(path, pod); err != nil {
		a.logger.Printf("agent: update pod status error: %v", err)
	}
}

func detectCapacity() NodeResources {
	// Detect system resources
	var memTotal int64
	// Simple: read from /proc/meminfo or use a default
	data, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		fmt.Sscanf(string(data), "MemTotal: %d kB", &memTotal)
		memTotal *= 1024 // Convert to bytes
	}
	if memTotal == 0 {
		memTotal = 2 * 1024 * 1024 * 1024 // Default: 2GB
	}

	cpuCores := runtime.NumCPU()

	return NodeResources{
		CPU:    int64(cpuCores) * 1000, // millicores
		Memory: memTotal,
		Pods:   110, // Default max pods per node
	}
}
