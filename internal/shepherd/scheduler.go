package shepherd

import (
	"fmt"
	"log"
	"sort"
	"time"
)

type Scheduler struct {
	store  *Store
	logger *log.Logger
}

func NewScheduler(store *Store, logger *log.Logger) *Scheduler {
	return &Scheduler{
		store:  store,
		logger: logger,
	}
}

// Run starts the scheduler loop that watches for unscheduled pods.
func (s *Scheduler) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	s.logger.Println("scheduler started")

	for {
		select {
		case <-stopCh:
			s.logger.Println("scheduler stopped")
			return
		case <-ticker.C:
			s.reconcile()
		}
	}
}

func (s *Scheduler) reconcile() {
	pods, err := s.store.ListPods("")
	if err != nil {
		s.logger.Printf("scheduler: list pods error: %v", err)
		return
	}

	for _, pod := range pods {
		if pod.Status.Phase == PodPending && pod.Spec.NodeName == "" {
			s.SchedulePod(pod)
		}
	}
}

// SchedulePod assigns a pod to the best available node.
func (s *Scheduler) SchedulePod(pod *Pod) {
	nodes, err := s.store.ListNodes()
	if err != nil {
		s.logger.Printf("scheduler: list nodes error: %v", err)
		return
	}

	if len(nodes) == 0 {
		s.logger.Printf("scheduler: no nodes available for pod %s", pod.Metadata.Name)
		pod.Status.Phase = PodPending
		pod.Status.Message = "no nodes available"
		s.store.UpdatePod(pod)
		return
	}

	// Filter feasible nodes
	feasible := s.filterNodes(nodes, pod)
	if len(feasible) == 0 {
		s.logger.Printf("scheduler: no feasible nodes for pod %s", pod.Metadata.Name)
		pod.Status.Message = "no feasible nodes: insufficient resources"
		s.store.UpdatePod(pod)
		return
	}

	// Score and rank nodes
	scored := s.scoreNodes(feasible, pod)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	selected := scored[0].node
	s.logger.Printf("scheduler: pod %s -> node %s (score: %d)",
		pod.Metadata.Name, selected.Metadata.Name, scored[0].score)

	pod.Spec.NodeName = selected.Metadata.Name
	pod.Status.Phase = PodPending
	pod.Status.Message = fmt.Sprintf("scheduled to node %s", selected.Metadata.Name)
	pod.Status.HostIP = selected.Spec.Address

	if err := s.store.UpdatePod(pod); err != nil {
		s.logger.Printf("scheduler: update pod error: %v", err)
		return
	}

	s.store.RecordEvent(Event{
		Type:      "Normal",
		Reason:    "Scheduled",
		Message:   fmt.Sprintf("Pod %s scheduled to node %s", pod.Metadata.Name, selected.Metadata.Name),
		Object:    "pod/" + pod.Metadata.Name,
		Timestamp: time.Now(),
	})
}

type scoredNode struct {
	node  *Node
	score int
}

func (s *Scheduler) filterNodes(nodes []*Node, pod *Pod) []*Node {
	var feasible []*Node

	for _, node := range nodes {
		// Check node is ready
		if node.Status.Condition != NodeReady {
			continue
		}

		// Check heartbeat freshness (30 seconds)
		if time.Since(node.Status.LastHeartbeat) > 30*time.Second {
			continue
		}

		// Check node selector
		if !matchLabels(node.Metadata.Labels, pod.Spec.NodeSelector) {
			continue
		}

		// Check resource capacity
		totalReq := podResourceRequests(pod)
		alloc := node.Status.Allocatable
		if totalReq.CPU > 0 && totalReq.CPU > alloc.CPU {
			continue
		}
		if totalReq.Memory > 0 && totalReq.Memory > alloc.Memory {
			continue
		}
		if node.Status.PodCount >= alloc.Pods {
			continue
		}

		feasible = append(feasible, node)
	}

	return feasible
}

func (s *Scheduler) scoreNodes(nodes []*Node, pod *Pod) []scoredNode {
	scored := make([]scoredNode, len(nodes))

	for i, node := range nodes {
		score := 0

		// Least-loaded: prefer nodes with fewer pods
		score += (node.Status.Allocatable.Pods - node.Status.PodCount) * 10

		// Resource balance: prefer nodes with more available resources
		if node.Status.Allocatable.CPU > 0 {
			usedRatio := float64(node.Status.Allocatable.CPU-podResourceRequests(pod).CPU) / float64(node.Status.Allocatable.CPU)
			score += int(usedRatio * 50)
		}

		if node.Status.Allocatable.Memory > 0 {
			usedRatio := float64(node.Status.Allocatable.Memory-podResourceRequests(pod).Memory) / float64(node.Status.Allocatable.Memory)
			score += int(usedRatio * 50)
		}

		scored[i] = scoredNode{node: node, score: score}
	}

	return scored
}

func podResourceRequests(pod *Pod) ResourceSpec {
	var total ResourceSpec
	for _, c := range pod.Spec.Containers {
		total.CPU += c.Resources.CPU
		total.Memory += c.Resources.Memory
	}
	return total
}

func matchLabels(nodeLabels, selector map[string]string) bool {
	if len(selector) == 0 {
		return true
	}
	for k, v := range selector {
		if nodeLabels[k] != v {
			return false
		}
	}
	return true
}
