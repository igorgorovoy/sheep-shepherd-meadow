package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"sheep/internal/cli"
	"sheep/internal/shepherd"
)

var apiServer = "localhost:9876"

func main() {
	// Check SHEPHERD_API env
	if addr := os.Getenv("SHEPHERD_API"); addr != "" {
		apiServer = addr
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "apply":
		cmdApply()
	case "get":
		cmdGet()
	case "delete":
		cmdDelete()
	case "describe":
		cmdDescribe()
	case "scale":
		cmdScale()
	case "logs":
		cmdLogs()
	case "nodes":
		listNodes()
	case "events":
		cmdEvents()
	case "info":
		cmdInfo()
	case "version":
		fmt.Println("sheepctl v0.1.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func cmdApply() {
	if len(os.Args) < 4 || os.Args[2] != "-f" {
		fatal("usage: sheepctl apply -f <file>")
	}

	data, err := os.ReadFile(os.Args[3])
	if err != nil {
		fatal("read file: %v", err)
	}

	// Detect kind
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		fatal("invalid JSON: %v", err)
	}

	kind, _ := raw["kind"].(string)
	switch strings.ToLower(kind) {
	case "pod":
		resp, err := apiPost("/api/v1/pods", data)
		if err != nil {
			fatal("create pod: %v", err)
		}
		var pod shepherd.Pod
		json.Unmarshal(resp, &pod)
		fmt.Printf("pod/%s created\n", pod.Metadata.Name)

	case "service":
		resp, err := apiPost("/api/v1/services", data)
		if err != nil {
			fatal("create service: %v", err)
		}
		var svc shepherd.Service
		json.Unmarshal(resp, &svc)
		fmt.Printf("service/%s created\n", svc.Metadata.Name)

	case "deployment":
		resp, err := apiPost("/api/v1/deployments", data)
		if err != nil {
			fatal("create deployment: %v", err)
		}
		var dep shepherd.Deployment
		json.Unmarshal(resp, &dep)
		fmt.Printf("deployment/%s created\n", dep.Metadata.Name)

	default:
		fatal("unknown kind: %s", kind)
	}
}

func cmdGet() {
	if len(os.Args) < 3 {
		fatal("usage: sheepctl get <resource> [name]")
	}

	resource := os.Args[2]
	ns := "default"

	// Parse --namespace/-n flag
	for i := 3; i < len(os.Args); i++ {
		if (os.Args[i] == "-n" || os.Args[i] == "--namespace") && i+1 < len(os.Args) {
			ns = os.Args[i+1]
		}
	}

	switch resource {
	case "pods", "pod", "po":
		if len(os.Args) > 3 && !strings.HasPrefix(os.Args[3], "-") {
			getPod(ns, os.Args[3])
		} else {
			listPods(ns)
		}
	case "services", "service", "svc":
		if len(os.Args) > 3 && !strings.HasPrefix(os.Args[3], "-") {
			getService(ns, os.Args[3])
		} else {
			listServices(ns)
		}
	case "deployments", "deployment", "deploy":
		if len(os.Args) > 3 && !strings.HasPrefix(os.Args[3], "-") {
			getDeployment(ns, os.Args[3])
		} else {
			listDeployments(ns)
		}
	case "nodes", "node", "no":
		listNodes()
	default:
		fatal("unknown resource: %s", resource)
	}
}

func listPods(ns string) {
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/pods", ns))
	if err != nil {
		fatal("list pods: %v", err)
	}

	var pods []shepherd.Pod
	json.Unmarshal(data, &pods)

	tbl := cli.NewTable("NAME", "STATUS", "NODE", "IP", "AGE")
	for _, p := range pods {
		age := timeAgo(p.Metadata.CreatedAt)
		tbl.AddRow(p.Metadata.Name, string(p.Status.Phase), p.Spec.NodeName, p.Status.PodIP, age)
	}
	tbl.Render(os.Stdout)
}

func getPod(ns, name string) {
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", ns, name))
	if err != nil {
		fatal("get pod: %v", err)
	}

	var pod shepherd.Pod
	json.Unmarshal(data, &pod)
	printJSON(pod)
}

func listServices(ns string) {
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/services", ns))
	if err != nil {
		fatal("list services: %v", err)
	}

	var svcs []shepherd.Service
	json.Unmarshal(data, &svcs)

	tbl := cli.NewTable("NAME", "TYPE", "CLUSTER-IP", "PORTS", "ENDPOINTS", "AGE")
	for _, s := range svcs {
		ports := formatPorts(s.Spec.Ports)
		eps := strconv.Itoa(len(s.Status.Endpoints))
		tbl.AddRow(s.Metadata.Name, string(s.Spec.Type), s.Status.ClusterIP, ports, eps, timeAgo(s.Metadata.CreatedAt))
	}
	tbl.Render(os.Stdout)
}

func getService(ns, name string) {
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/services/%s", ns, name))
	if err != nil {
		fatal("get service: %v", err)
	}

	var svc shepherd.Service
	json.Unmarshal(data, &svc)
	printJSON(svc)
}

func listDeployments(ns string) {
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/deployments", ns))
	if err != nil {
		fatal("list deployments: %v", err)
	}

	var deps []shepherd.Deployment
	json.Unmarshal(data, &deps)

	tbl := cli.NewTable("NAME", "READY", "AVAILABLE", "AGE")
	for _, d := range deps {
		ready := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Spec.Replicas)
		tbl.AddRow(d.Metadata.Name, ready, strconv.Itoa(d.Status.AvailableReplicas), timeAgo(d.Metadata.CreatedAt))
	}
	tbl.Render(os.Stdout)
}

func getDeployment(ns, name string) {
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/deployments/%s", ns, name))
	if err != nil {
		fatal("get deployment: %v", err)
	}

	var dep shepherd.Deployment
	json.Unmarshal(data, &dep)
	printJSON(dep)
}

func listNodes() {
	data, err := apiGet("/api/v1/nodes")
	if err != nil {
		fatal("list nodes: %v", err)
	}

	var nodes []shepherd.Node
	json.Unmarshal(data, &nodes)

	tbl := cli.NewTable("NAME", "STATUS", "PODS", "CPU", "MEMORY", "AGE")
	for _, n := range nodes {
		status := string(n.Status.Condition)
		cpu := fmt.Sprintf("%dm", n.Status.Allocatable.CPU)
		mem := cli.FormatBytes(n.Status.Allocatable.Memory)
		tbl.AddRow(n.Metadata.Name, status, strconv.Itoa(n.Status.PodCount), cpu, mem, timeAgo(n.Metadata.CreatedAt))
	}
	tbl.Render(os.Stdout)
}

func cmdDelete() {
	if len(os.Args) < 4 {
		fatal("usage: sheepctl delete <resource> <name>")
	}

	resource := os.Args[2]
	name := os.Args[3]
	ns := "default"

	for i := 4; i < len(os.Args); i++ {
		if (os.Args[i] == "-n" || os.Args[i] == "--namespace") && i+1 < len(os.Args) {
			ns = os.Args[i+1]
		}
	}

	var path string
	switch resource {
	case "pod", "pods", "po":
		path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", ns, name)
	case "service", "services", "svc":
		path = fmt.Sprintf("/api/v1/namespaces/%s/services/%s", ns, name)
	case "deployment", "deployments", "deploy":
		path = fmt.Sprintf("/api/v1/namespaces/%s/deployments/%s", ns, name)
	case "node", "nodes":
		path = fmt.Sprintf("/api/v1/nodes/%s", name)
	default:
		fatal("unknown resource: %s", resource)
	}

	if err := apiDelete(path); err != nil {
		fatal("delete: %v", err)
	}
	fmt.Printf("%s/%s deleted\n", resource, name)
}

func cmdDescribe() {
	if len(os.Args) < 4 {
		fatal("usage: sheepctl describe <resource> <name>")
	}

	// Describe just does a get with formatted output
	resource := os.Args[2]
	name := os.Args[3]
	ns := "default"

	var path string
	switch resource {
	case "pod", "pods":
		path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", ns, name)
	case "service", "services":
		path = fmt.Sprintf("/api/v1/namespaces/%s/services/%s", ns, name)
	case "deployment", "deployments":
		path = fmt.Sprintf("/api/v1/namespaces/%s/deployments/%s", ns, name)
	case "node", "nodes":
		path = fmt.Sprintf("/api/v1/nodes/%s", name)
	default:
		fatal("unknown resource: %s", resource)
	}

	data, err := apiGet(path)
	if err != nil {
		fatal("describe: %v", err)
	}

	// Pretty print
	var raw any
	json.Unmarshal(data, &raw)
	printJSON(raw)
}

func cmdScale() {
	if len(os.Args) < 4 {
		fatal("usage: sheepctl scale deployment/<name> --replicas=<N>")
	}

	parts := strings.SplitN(os.Args[2], "/", 2)
	if len(parts) != 2 || parts[0] != "deployment" {
		fatal("usage: sheepctl scale deployment/<name> --replicas=<N>")
	}

	name := parts[1]
	ns := "default"
	replicas := 0

	for _, arg := range os.Args[3:] {
		if strings.HasPrefix(arg, "--replicas=") {
			v, _ := strconv.Atoi(strings.TrimPrefix(arg, "--replicas="))
			replicas = v
		}
		if (arg == "-n" || arg == "--namespace") {
			// handled in next iteration
		}
	}

	if replicas <= 0 {
		fatal("--replicas must be positive")
	}

	// Get current deployment
	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/deployments/%s", ns, name))
	if err != nil {
		fatal("get deployment: %v", err)
	}

	var dep shepherd.Deployment
	json.Unmarshal(data, &dep)

	old := dep.Spec.Replicas
	dep.Spec.Replicas = replicas

	depData, _ := json.Marshal(dep)
	if err := apiPut(fmt.Sprintf("/api/v1/namespaces/%s/deployments/%s", ns, name), depData); err != nil {
		fatal("scale: %v", err)
	}

	fmt.Printf("deployment/%s scaled from %d to %d\n", name, old, replicas)
}

func cmdLogs() {
	if len(os.Args) < 3 {
		fatal("usage: sheepctl logs <pod>")
	}

	name := os.Args[2]
	ns := "default"

	data, err := apiGet(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", ns, name))
	if err != nil {
		fatal("get pod: %v", err)
	}

	var pod shepherd.Pod
	json.Unmarshal(data, &pod)

	fmt.Printf("--- logs for pod/%s (node: %s) ---\n", pod.Metadata.Name, pod.Spec.NodeName)
	for _, cs := range pod.Status.Containers {
		fmt.Printf("[%s] container_id=%s state=%s\n", cs.Name, cs.ContainerID, cs.State)
	}
}

func cmdEvents() {
	data, err := apiGet("/api/v1/events")
	if err != nil {
		fatal("list events: %v", err)
	}

	var events []shepherd.Event
	json.Unmarshal(data, &events)

	tbl := cli.NewTable("TYPE", "REASON", "OBJECT", "MESSAGE", "AGE")
	for _, e := range events {
		tbl.AddRow(e.Type, e.Reason, e.Object, cli.Truncate(e.Message, 50), timeAgo(e.Timestamp))
	}
	tbl.Render(os.Stdout)
}

func cmdInfo() {
	data, err := apiGet("/api/v1/info")
	if err != nil {
		fatal("info: %v", err)
	}

	var info map[string]any
	json.Unmarshal(data, &info)

	fmt.Println("Shepherd Cluster Info")
	fmt.Println("---------------------")
	for k, v := range info {
		fmt.Printf("%-15s %v\n", k+":", v)
	}
}

// --- API helpers ---

func apiGet(path string) ([]byte, error) {
	resp, err := http.Get("http://" + apiServer + path)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", apiServer, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return io.ReadAll(resp.Body)
}

func apiPost(path string, data []byte) ([]byte, error) {
	resp, err := http.Post("http://"+apiServer+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func apiPut(path string, data []byte) error {
	req, err := http.NewRequest(http.MethodPut, "http://"+apiServer+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func apiDelete(path string) error {
	req, err := http.NewRequest(http.MethodDelete, "http://"+apiServer+path, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// --- Helpers ---

func formatPorts(ports []shepherd.ServicePort) string {
	var parts []string
	for _, p := range ports {
		s := fmt.Sprintf("%d->%d", p.Port, p.TargetPort)
		if p.NodePort > 0 {
			s += fmt.Sprintf(":%d", p.NodePort)
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ",")
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func timeAgo(t time.Time) string {
	if t.IsZero() {
		return "<unknown>"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "sheepctl: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`sheepctl - shepherd cluster CLI

Usage: sheepctl <command> [options]

Commands:
  apply       Create a resource from a JSON file
  get         List or get resources (pods, services, deployments, nodes)
  delete      Delete a resource
  describe    Show detailed info about a resource
  scale       Scale a deployment
  logs        Show pod logs
  nodes       List cluster nodes
  events      Show cluster events
  info        Show cluster info

Flags:
  -n, --namespace   Namespace (default: "default")

Environment:
  SHEPHERD_API      API server address (default: localhost:9876)

Examples:
  sheepctl apply -f pod.json
  sheepctl get pods
  sheepctl get pod my-pod
  sheepctl get deployments
  sheepctl scale deployment/web --replicas=3
  sheepctl delete pod my-pod
  sheepctl events
  sheepctl nodes`)
}
