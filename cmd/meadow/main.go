package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sheep/internal/registry"
)

func main() {
	var (
		addr    = flag.String("addr", ":5555", "Listen address")
		dataDir = flag.String("data-dir", "/var/lib/meadow", "Data directory for blobs and manifests")
	)
	flag.Parse()

	logger := log.New(os.Stdout, "[meadow] ", log.LstdFlags|log.Lshortfile)

	os.MkdirAll(*dataDir, 0755)

	storage := registry.NewStorage(*dataDir)
	if err := storage.Init(); err != nil {
		logger.Fatalf("init storage: %v", err)
	}

	server := registry.NewServer(*addr, storage, logger)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Println("shutting down...")
		server.Shutdown(context.Background())
	}()

	fmt.Printf(`
  ┌──────────────────────────────────────┐
  │         meadow registry %s        │
  │   Sheep's own image registry         │
  │                                      │
  │   Listening on %s             │
  │   Storage: %s  │
  └──────────────────────────────────────┘

`, "v0.1.0", pad(*addr, 16), pad(*dataDir, 22))

	if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
		logger.Fatalf("server: %v", err)
	}
}

func pad(s string, width int) string {
	for len(s) < width {
		s = s + " "
	}
	return s
}
