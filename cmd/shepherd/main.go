package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sheep/internal/shepherd"
)

func main() {
	var (
		addr     = flag.String("addr", ":9876", "API server listen address")
		dataDir  = flag.String("data-dir", "/var/lib/shepherd", "Data directory")
		nodeName = flag.String("node-name", "", "Node name (for agent mode)")
		apiAddr  = flag.String("api-addr", "", "API server address (for agent mode)")
		mode     = flag.String("mode", "server", "Run mode: server, agent, or standalone")
	)
	flag.Parse()

	logger := log.New(os.Stdout, "[shepherd] ", log.LstdFlags|log.Lshortfile)

	switch *mode {
	case "server":
		runServer(*addr, *dataDir, logger)
	case "agent":
		if *apiAddr == "" {
			fmt.Fprintln(os.Stderr, "agent mode requires --api-addr")
			os.Exit(1)
		}
		runAgent(*nodeName, *apiAddr, logger)
	case "standalone":
		// Runs both server and agent in a single process
		runStandalone(*addr, *dataDir, *nodeName, logger)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}

func runServer(addr, dataDir string, logger *log.Logger) {
	os.MkdirAll(dataDir, 0755)

	store, err := shepherd.NewStore(dataDir + "/shepherd.db")
	if err != nil {
		logger.Fatalf("open store: %v", err)
	}
	defer store.Close()

	stopCh := make(chan struct{})
	scheduler := shepherd.NewScheduler(store, logger)
	go scheduler.Run(stopCh)

	replicationCtrl := shepherd.NewReplicationController(store, scheduler, logger)
	go replicationCtrl.Run(stopCh)

	serviceCtrl := shepherd.NewServiceController(store, logger)
	go serviceCtrl.Run(stopCh)

	nodeCtrl := shepherd.NewNodeController(store, logger)
	go nodeCtrl.Run(stopCh)

	api := shepherd.NewAPIServer(addr, store, scheduler, logger)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Println("shutting down...")
		close(stopCh)
		api.Shutdown(context.Background())
	}()

	logger.Printf("shepherd server starting on %s", addr)
	if err := api.Start(); err != nil && err.Error() != "http: Server closed" {
		logger.Fatalf("api server: %v", err)
	}
}

func runAgent(nodeName, apiAddr string, logger *log.Logger) {
	agent := shepherd.NewAgent(nodeName, apiAddr, logger)

	stopCh := make(chan struct{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		close(stopCh)
	}()

	if err := agent.Run(stopCh); err != nil {
		logger.Fatalf("agent: %v", err)
	}
}

func runStandalone(addr, dataDir, nodeName string, logger *log.Logger) {
	os.MkdirAll(dataDir, 0755)

	store, err := shepherd.NewStore(dataDir + "/shepherd.db")
	if err != nil {
		logger.Fatalf("open store: %v", err)
	}
	defer store.Close()

	stopCh := make(chan struct{})

	scheduler := shepherd.NewScheduler(store, logger)
	go scheduler.Run(stopCh)

	replicationCtrl := shepherd.NewReplicationController(store, scheduler, logger)
	go replicationCtrl.Run(stopCh)

	serviceCtrl := shepherd.NewServiceController(store, logger)
	go serviceCtrl.Run(stopCh)

	nodeCtrl := shepherd.NewNodeController(store, logger)
	go nodeCtrl.Run(stopCh)

	api := shepherd.NewAPIServer(addr, store, scheduler, logger)

	// Start agent pointing at ourselves
	if nodeName == "" {
		host, _ := os.Hostname()
		nodeName = host
	}
	actualAddr := addr
	if actualAddr[0] == ':' {
		actualAddr = "localhost" + actualAddr
	}
	agent := shepherd.NewAgent(nodeName, actualAddr, logger)

	// Start API server first, then agent
	go func() {
		if err := api.Start(); err != nil && err.Error() != "http: Server closed" {
			logger.Fatalf("api server: %v", err)
		}
	}()

	// Small delay for API server to be ready
	go func() {
		// Agent will retry on connection failure
		agent.Run(stopCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Println("shutting down...")
	close(stopCh)
	api.Shutdown(context.Background())
}
