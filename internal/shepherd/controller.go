package shepherd

import (
	"fmt"
	"log"
	"time"
)

// ReplicationController watches deployments and ensures the desired number of pod replicas.
type ReplicationController struct {
	store     *Store
	scheduler *Scheduler
	logger    *log.Logger
}

func NewReplicationController(store *Store, scheduler *Scheduler, logger *log.Logger) *ReplicationController {
	return &ReplicationController{
		store:     store,
		scheduler: scheduler,
		logger:    logger,
	}
}

func (rc *ReplicationController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	rc.logger.Println("replication controller started")

	for {
		select {
		case <-stopCh:
			rc.logger.Println("replication controller stopped")
			return
		case <-ticker.C:
			rc.reconcile()
		}
	}
}

func (rc *ReplicationController) reconcile() {
	deployments, err := rc.store.ListDeployments("")
	if err != nil {
		rc.logger.Printf("replication controller: list deployments error: %v", err)
		return
	}

	for _, dep := range deployments {
		rc.reconcileDeployment(dep)
	}
}

func (rc *ReplicationController) reconcileDeployment(dep *Deployment) {
	// Find pods matching this deployment's selector
	allPods, err := rc.store.ListPods(dep.Metadata.Namespace)
	if err != nil {
		return
	}

	var matchingPods []*Pod
	for _, pod := range allPods {
		if matchLabels(pod.Metadata.Labels, dep.Spec.Selector) {
			matchingPods = append(matchingPods, pod)
		}
	}

	current := len(matchingPods)
	desired := dep.Spec.Replicas

	if current < desired {
		// Scale up: find available indices that don't collide with existing pods
		existing := make(map[string]bool)
		for _, pod := range matchingPods {
			existing[pod.Metadata.Name] = true
		}
		created := 0
		for idx := 0; created < desired-current; idx++ {
			name := fmt.Sprintf("%s-%d", dep.Metadata.Name, idx)
			if !existing[name] {
				rc.createPodForDeployment(dep, idx)
				created++
			}
		}
		rc.logger.Printf("replication controller: scaled up %s from %d to %d",
			dep.Metadata.Name, current, desired)
	} else if current > desired {
		// Scale down
		for i := 0; i < current-desired; i++ {
			pod := matchingPods[len(matchingPods)-1-i]
			if err := rc.store.DeletePod(pod.Metadata.Namespace, pod.Metadata.Name); err != nil {
				rc.logger.Printf("replication controller: delete pod error: %v", err)
			}
		}
		rc.logger.Printf("replication controller: scaled down %s from %d to %d",
			dep.Metadata.Name, current, desired)
	}

	// Count ready pods
	ready := 0
	for _, pod := range matchingPods {
		if pod.Status.Phase == PodRunning {
			ready++
		}
	}

	// Update deployment status
	dep.Status.Replicas = len(matchingPods)
	dep.Status.ReadyReplicas = ready
	dep.Status.AvailableReplicas = ready
	rc.store.UpdateDeployment(dep)
}

func (rc *ReplicationController) createPodForDeployment(dep *Deployment, index int) {
	pod := &Pod{
		Kind: "Pod",
		Metadata: ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", dep.Metadata.Name, index),
			Namespace: dep.Metadata.Namespace,
			UID:       generateUID(),
			Labels:    mergeLabels(dep.Spec.Template.Metadata.Labels, dep.Spec.Selector),
			CreatedAt: time.Now(),
		},
		Spec: dep.Spec.Template.Spec,
		Status: PodStatus{
			Phase: PodPending,
		},
	}

	if err := rc.store.CreatePod(pod); err != nil {
		rc.logger.Printf("replication controller: create pod error: %v", err)
		return
	}

	rc.store.RecordEvent(Event{
		Type:      "Normal",
		Reason:    "Created",
		Message:   fmt.Sprintf("Created pod %s for deployment %s", pod.Metadata.Name, dep.Metadata.Name),
		Object:    "deployment/" + dep.Metadata.Name,
		Timestamp: time.Now(),
	})

	// Trigger scheduling
	go rc.scheduler.SchedulePod(pod)
}

func mergeLabels(sets ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, s := range sets {
		for k, v := range s {
			result[k] = v
		}
	}
	return result
}

// ServiceController watches services and updates their endpoint lists.
type ServiceController struct {
	store  *Store
	logger *log.Logger
}

func NewServiceController(store *Store, logger *log.Logger) *ServiceController {
	return &ServiceController{
		store:  store,
		logger: logger,
	}
}

func (sc *ServiceController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	sc.logger.Println("service controller started")

	for {
		select {
		case <-stopCh:
			sc.logger.Println("service controller stopped")
			return
		case <-ticker.C:
			sc.reconcile()
		}
	}
}

func (sc *ServiceController) reconcile() {
	services, err := sc.store.ListServices("")
	if err != nil {
		return
	}

	for _, svc := range services {
		sc.reconcileService(svc)
	}
}

func (sc *ServiceController) reconcileService(svc *Service) {
	pods, err := sc.store.ListPods(svc.Metadata.Namespace)
	if err != nil {
		return
	}

	var endpoints []string
	for _, pod := range pods {
		if matchLabels(pod.Metadata.Labels, svc.Spec.Selector) && pod.Status.Phase == PodRunning {
			if pod.Status.PodIP != "" {
				for _, port := range svc.Spec.Ports {
					endpoints = append(endpoints, fmt.Sprintf("%s:%d", pod.Status.PodIP, port.TargetPort))
				}
			}
		}
	}

	svc.Status.Endpoints = endpoints
	sc.store.UpdateService(svc)
}

// NodeController monitors node health via heartbeats.
type NodeController struct {
	store  *Store
	logger *log.Logger
}

func NewNodeController(store *Store, logger *log.Logger) *NodeController {
	return &NodeController{
		store:  store,
		logger: logger,
	}
}

func (nc *NodeController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	nc.logger.Println("node controller started")

	for {
		select {
		case <-stopCh:
			nc.logger.Println("node controller stopped")
			return
		case <-ticker.C:
			nc.reconcile()
		}
	}
}

func (nc *NodeController) reconcile() {
	nodes, err := nc.store.ListNodes()
	if err != nil {
		return
	}

	for _, node := range nodes {
		if time.Since(node.Status.LastHeartbeat) > 30*time.Second {
			if node.Status.Condition != NodeNotReady {
				nc.logger.Printf("node controller: node %s is not ready (last heartbeat: %s ago)",
					node.Metadata.Name, time.Since(node.Status.LastHeartbeat).Round(time.Second))
				node.Status.Condition = NodeNotReady
				nc.store.UpdateNode(node)

				nc.store.RecordEvent(Event{
					Type:      "Warning",
					Reason:    "NodeNotReady",
					Message:   fmt.Sprintf("Node %s heartbeat timeout", node.Metadata.Name),
					Object:    "node/" + node.Metadata.Name,
					Timestamp: time.Now(),
				})
			}
		}
	}
}
