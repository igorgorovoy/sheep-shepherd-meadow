package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sheep/internal/container"

	"sheep/internal/shepherd"
)

func main() {
	// Handle container init re-exec (agent starts containers by re-execing this binary)
	if len(os.Args) > 1 && os.Args[1] == "init" {
		handleInit()
		return
	}

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
		close(stopCh)
		logger.Printf("api server error: %v", err)
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
	errCh := make(chan error, 1)

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
			errCh <- fmt.Errorf("api server: %w", err)
		}
	}()

	go func() {
		if err := agent.Run(stopCh); err != nil {
			errCh <- fmt.Errorf("agent: %w", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Println("shutting down...")
	case err := <-errCh:
		logger.Printf("fatal: %v, shutting down...", err)
	}

	close(stopCh)
	api.Shutdown(context.Background())
}

func handleInit() {
	var rootfs, hostname string
	var command []string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--rootfs":
			i++
			if i < len(args) {
				rootfs = args[i]
			}
		case "--hostname":
			i++
			if i < len(args) {
				hostname = args[i]
			}
		case "--":
			command = args[i+1:]
			i = len(args)
		}
	}

	if rootfs == "" {
		fmt.Fprintln(os.Stderr, "shepherd: init: --rootfs is required")
		os.Exit(1)
	}

	if err := container.ContainerInit(rootfs, hostname, command); err != nil {
		fmt.Fprintf(os.Stderr, "shepherd: init: %v\n", err)
		os.Exit(1)
	}
}
