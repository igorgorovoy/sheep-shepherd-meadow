package shepherd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"sheep/internal/dashboard"
)

type APIServer struct {
	store     *Store
	scheduler *Scheduler
	server    *http.Server
	logger    *log.Logger
}

func NewAPIServer(addr string, store *Store, scheduler *Scheduler, logger *log.Logger) *APIServer {
	api := &APIServer{
		store:     store,
		scheduler: scheduler,
		logger:    logger,
	}

	mux := http.NewServeMux()

	// Pods
	mux.HandleFunc("/api/v1/pods", api.handlePods)
	mux.HandleFunc("/api/v1/namespaces/", api.handleNamespacedResource)

	// Nodes
	mux.HandleFunc("/api/v1/nodes", api.handleNodes)
	mux.HandleFunc("/api/v1/nodes/", api.handleNode)

	// Services
	mux.HandleFunc("/api/v1/services", api.handleServices)

	// Deployments
	mux.HandleFunc("/api/v1/deployments", api.handleDeployments)

	// Events
	mux.HandleFunc("/api/v1/events", api.handleEvents)

	// Health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Cluster info
	mux.HandleFunc("/api/v1/info", api.handleInfo)

	// Aggregate cluster summary (convenience endpoint for the dashboard)
	mux.HandleFunc("/api/v1/cluster/summary", api.handleClusterSummary)

	// Dashboard SPA: fallback handler for any non-API path. Registered on the
	// root pattern, which the mux only matches when no more specific pattern
	// (e.g. /api/v1/..., /healthz) applies, so API routes take precedence.
	dashboardHandler := dashboard.Handler()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}
		dashboardHandler.ServeHTTP(w, r)
	})

	api.server = &http.Server{
		Addr:         addr,
		Handler:      api.logging(api.cors(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return api
}

func (api *APIServer) Start() error {
	api.logger.Printf("API server listening on %s", api.server.Addr)
	return api.server.ListenAndServe()
}

func (api *APIServer) Shutdown(ctx context.Context) error {
	return api.server.Shutdown(ctx)
}

func (api *APIServer) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		api.logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// cors adds permissive CORS headers so a cross-origin dev SPA (e.g. the Vite
// dev server on http://localhost:5173) can call the API on
// http://localhost:9876. It answers preflight OPTIONS requests with 204.
func (api *APIServer) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Pod Handlers ---

func (api *APIServer) handlePods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listPods(w, r, "default")
	case http.MethodPost:
		api.createPod(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *APIServer) handleNamespacedResource(w http.ResponseWriter, r *http.Request) {
	// /api/v1/namespaces/{ns}/pods/{name}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/namespaces/"), "/")
	if len(parts) < 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	ns := parts[0]
	resource := parts[1]

	switch resource {
	case "pods":
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				api.listPods(w, r, ns)
			case http.MethodPost:
				api.createPod(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			name := parts[2]
			switch r.Method {
			case http.MethodGet:
				api.getPod(w, r, ns, name)
			case http.MethodPut:
				api.updatePod(w, r, ns, name)
			case http.MethodDelete:
				api.deletePod(w, r, ns, name)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	case "services":
		if len(parts) == 2 {
			api.listServicesNS(w, r, ns)
		} else {
			name := parts[2]
			switch r.Method {
			case http.MethodGet:
				api.getService(w, r, ns, name)
			case http.MethodDelete:
				api.deleteService(w, r, ns, name)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	case "deployments":
		if len(parts) == 2 {
			api.listDeploymentsNS(w, r, ns)
		} else {
			name := parts[2]
			switch r.Method {
			case http.MethodGet:
				api.getDeployment(w, r, ns, name)
			case http.MethodPut:
				api.updateDeployment(w, r, ns, name)
			case http.MethodDelete:
				api.deleteDeployment(w, r, ns, name)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	default:
		http.Error(w, "unknown resource: "+resource, http.StatusNotFound)
	}
}

func (api *APIServer) createPod(w http.ResponseWriter, r *http.Request) {
	var pod Pod
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		httpError(w, http.StatusBadRequest, "invalid pod: %v", err)
		return
	}

	pod.Kind = "Pod"
	if pod.Metadata.Namespace == "" {
		pod.Metadata.Namespace = "default"
	}
	if pod.Metadata.UID == "" {
		pod.Metadata.UID = generateUID()
	}
	pod.Metadata.CreatedAt = time.Now()
	pod.Status.Phase = PodPending

	if err := api.store.CreatePod(&pod); err != nil {
		httpError(w, http.StatusInternalServerError, "create pod: %v", err)
		return
	}

	api.store.RecordEvent(Event{
		Type:      "Normal",
		Reason:    "Created",
		Message:   fmt.Sprintf("Pod %s created", pod.Metadata.Name),
		Object:    "pod/" + pod.Metadata.Name,
		Timestamp: time.Now(),
	})

	// Trigger scheduling
	go api.scheduler.SchedulePod(&pod)

	respondJSON(w, http.StatusCreated, pod)
}

func (api *APIServer) getPod(w http.ResponseWriter, _ *http.Request, ns, name string) {
	pod, err := api.store.GetPod(ns, name)
	if err != nil {
		httpError(w, http.StatusNotFound, "pod %s not found", name)
		return
	}
	respondJSON(w, http.StatusOK, pod)
}

func (api *APIServer) listPods(w http.ResponseWriter, _ *http.Request, ns string) {
	pods, err := api.store.ListPods(ns)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list pods: %v", err)
		return
	}
	respondJSON(w, http.StatusOK, pods)
}

func (api *APIServer) updatePod(w http.ResponseWriter, r *http.Request, ns, name string) {
	var pod Pod
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		httpError(w, http.StatusBadRequest, "invalid pod: %v", err)
		return
	}

	pod.Metadata.Namespace = ns
	pod.Metadata.Name = name

	if err := api.store.UpdatePod(&pod); err != nil {
		httpError(w, http.StatusInternalServerError, "update pod: %v", err)
		return
	}

	respondJSON(w, http.StatusOK, pod)
}

func (api *APIServer) deletePod(w http.ResponseWriter, _ *http.Request, ns, name string) {
	if err := api.store.DeletePod(ns, name); err != nil {
		httpError(w, http.StatusNotFound, "pod %s not found", name)
		return
	}

	api.store.RecordEvent(Event{
		Type:      "Normal",
		Reason:    "Deleted",
		Message:   fmt.Sprintf("Pod %s deleted", name),
		Object:    "pod/" + name,
		Timestamp: time.Now(),
	})

	w.WriteHeader(http.StatusNoContent)
}

// --- Node Handlers ---

func (api *APIServer) handleNodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		nodes, err := api.store.ListNodes()
		if err != nil {
			httpError(w, http.StatusInternalServerError, "list nodes: %v", err)
			return
		}
		respondJSON(w, http.StatusOK, nodes)
	case http.MethodPost:
		var node Node
		if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
			httpError(w, http.StatusBadRequest, "invalid node: %v", err)
			return
		}
		node.Kind = "Node"
		if node.Metadata.UID == "" {
			node.Metadata.UID = generateUID()
		}
		node.Metadata.CreatedAt = time.Now()
		node.Status.Condition = NodeReady
		node.Status.LastHeartbeat = time.Now()

		if err := api.store.RegisterNode(&node); err != nil {
			httpError(w, http.StatusInternalServerError, "register node: %v", err)
			return
		}
		respondJSON(w, http.StatusCreated, node)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *APIServer) handleNode(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/nodes/")

	switch r.Method {
	case http.MethodGet:
		node, err := api.store.GetNode(name)
		if err != nil {
			httpError(w, http.StatusNotFound, "node %s not found", name)
			return
		}
		respondJSON(w, http.StatusOK, node)
	case http.MethodPut:
		var node Node
		if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
			httpError(w, http.StatusBadRequest, "invalid node: %v", err)
			return
		}
		node.Metadata.Name = name
		if err := api.store.UpdateNode(&node); err != nil {
			httpError(w, http.StatusInternalServerError, "update node: %v", err)
			return
		}
		respondJSON(w, http.StatusOK, node)
	case http.MethodDelete:
		if err := api.store.DeleteNode(name); err != nil {
			httpError(w, http.StatusNotFound, "node %s not found", name)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Service Handlers ---

func (api *APIServer) handleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listServicesNS(w, r, "default")
	case http.MethodPost:
		api.createService(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *APIServer) createService(w http.ResponseWriter, r *http.Request) {
	var svc Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		httpError(w, http.StatusBadRequest, "invalid service: %v", err)
		return
	}

	svc.Kind = "Service"
	if svc.Metadata.Namespace == "" {
		svc.Metadata.Namespace = "default"
	}
	if svc.Metadata.UID == "" {
		svc.Metadata.UID = generateUID()
	}
	svc.Metadata.CreatedAt = time.Now()

	if err := api.store.CreateService(&svc); err != nil {
		httpError(w, http.StatusInternalServerError, "create service: %v", err)
		return
	}

	respondJSON(w, http.StatusCreated, svc)
}

func (api *APIServer) listServicesNS(w http.ResponseWriter, _ *http.Request, ns string) {
	svcs, err := api.store.ListServices(ns)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list services: %v", err)
		return
	}
	respondJSON(w, http.StatusOK, svcs)
}

func (api *APIServer) getService(w http.ResponseWriter, _ *http.Request, ns, name string) {
	svc, err := api.store.GetService(ns, name)
	if err != nil {
		httpError(w, http.StatusNotFound, "service %s not found", name)
		return
	}
	respondJSON(w, http.StatusOK, svc)
}

func (api *APIServer) deleteService(w http.ResponseWriter, _ *http.Request, ns, name string) {
	if err := api.store.DeleteService(ns, name); err != nil {
		httpError(w, http.StatusNotFound, "service %s not found", name)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Deployment Handlers ---

func (api *APIServer) handleDeployments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listDeploymentsNS(w, r, "default")
	case http.MethodPost:
		api.createDeployment(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *APIServer) createDeployment(w http.ResponseWriter, r *http.Request) {
	var dep Deployment
	if err := json.NewDecoder(r.Body).Decode(&dep); err != nil {
		httpError(w, http.StatusBadRequest, "invalid deployment: %v", err)
		return
	}

	dep.Kind = "Deployment"
	if dep.Metadata.Namespace == "" {
		dep.Metadata.Namespace = "default"
	}
	if dep.Metadata.UID == "" {
		dep.Metadata.UID = generateUID()
	}
	dep.Metadata.CreatedAt = time.Now()

	if dep.Spec.Replicas <= 0 {
		dep.Spec.Replicas = 1
	}

	if err := api.store.CreateDeployment(&dep); err != nil {
		httpError(w, http.StatusInternalServerError, "create deployment: %v", err)
		return
	}

	respondJSON(w, http.StatusCreated, dep)
}

func (api *APIServer) listDeploymentsNS(w http.ResponseWriter, _ *http.Request, ns string) {
	deps, err := api.store.ListDeployments(ns)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list deployments: %v", err)
		return
	}
	respondJSON(w, http.StatusOK, deps)
}

func (api *APIServer) getDeployment(w http.ResponseWriter, _ *http.Request, ns, name string) {
	dep, err := api.store.GetDeployment(ns, name)
	if err != nil {
		httpError(w, http.StatusNotFound, "deployment %s not found", name)
		return
	}
	respondJSON(w, http.StatusOK, dep)
}

func (api *APIServer) updateDeployment(w http.ResponseWriter, r *http.Request, ns, name string) {
	var dep Deployment
	if err := json.NewDecoder(r.Body).Decode(&dep); err != nil {
		httpError(w, http.StatusBadRequest, "invalid deployment: %v", err)
		return
	}

	dep.Metadata.Namespace = ns
	dep.Metadata.Name = name

	if err := api.store.UpdateDeployment(&dep); err != nil {
		httpError(w, http.StatusInternalServerError, "update deployment: %v", err)
		return
	}

	respondJSON(w, http.StatusOK, dep)
}

func (api *APIServer) deleteDeployment(w http.ResponseWriter, _ *http.Request, ns, name string) {
	if err := api.store.DeleteDeployment(ns, name); err != nil {
		httpError(w, http.StatusNotFound, "deployment %s not found", name)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Events ---

func (api *APIServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events, err := api.store.ListEvents(100)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list events: %v", err)
		return
	}
	respondJSON(w, http.StatusOK, events)
}

// --- Info ---

func (api *APIServer) handleInfo(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, api.clusterInfo())
}

// clusterInfo builds the map returned by /api/v1/info. It is shared with the
// cluster summary endpoint.
func (api *APIServer) clusterInfo() map[string]any {
	nodes, _ := api.store.ListNodes()
	pods, _ := api.store.ListPods("")

	return map[string]any{
		"version":    "v0.1.0",
		"name":       "shepherd",
		"node_count": len(nodes),
		"pod_count":  len(pods),
	}
}

// --- Cluster Summary ---

func (api *APIServer) handleClusterSummary(w http.ResponseWriter, _ *http.Request) {
	nodes, err := api.store.ListNodes()
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list nodes: %v", err)
		return
	}
	pods, err := api.store.ListPods("")
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list pods: %v", err)
		return
	}
	services, err := api.store.ListServices("default")
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list services: %v", err)
		return
	}
	deployments, err := api.store.ListDeployments("default")
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list deployments: %v", err)
		return
	}
	events, err := api.store.ListEvents(100)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "list events: %v", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"info":        api.clusterInfo(),
		"nodes":       nodes,
		"pods":        pods,
		"deployments": deployments,
		"services":    services,
		"events":      events,
	})
}

// --- Helpers ---

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, status int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	http.Error(w, msg, status)
}

func generateUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
