package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := flag.Int("port", 8888, "listen port")
	flag.Parse()

	hostname, _ := os.Hostname()
	containerID := os.Getenv("SHEEP_CONTAINER_ID")
	if containerID == "" {
		containerID = "unknown"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Server", "tinynginx/1.0 (sheep)")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>tinynginx on sheep</title>
  <style>
    body { font-family: monospace; background: #1a1a2e; color: #e0e0e0; padding: 40px; }
    .box { background: #16213e; border: 1px solid #0f3460; border-radius: 8px; padding: 30px; max-width: 600px; margin: 0 auto; }
    h1 { color: #e94560; margin-top: 0; }
    .info { color: #a0a0a0; }
    .val  { color: #53d8fb; }
    hr { border: none; border-top: 1px solid #0f3460; margin: 20px 0; }
    .sheep { font-size: 48px; text-align: center; }
  </style>
</head>
<body>
  <div class="box">
    <div class="sheep">🐑</div>
    <h1>tinynginx on sheep</h1>
    <hr>
    <p><span class="info">Container ID:</span> <span class="val">%s</span></p>
    <p><span class="info">Hostname:</span>     <span class="val">%s</span></p>
    <p><span class="info">Port:</span>         <span class="val">%d</span></p>
    <p><span class="info">PID:</span>          <span class="val">%d</span></p>
    <p><span class="info">Time:</span>         <span class="val">%s</span></p>
    <p><span class="info">Request:</span>      <span class="val">%s %s</span></p>
    <p><span class="info">Remote:</span>       <span class="val">%s</span></p>
    <hr>
    <p class="info">Served by sheep container runtime</p>
  </div>
</body>
</html>`,
			containerID[:min(12, len(containerID))],
			hostname,
			*port,
			os.Getpid(),
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method, r.URL.Path,
			r.RemoteAddr,
		)
		log.Printf("%s %s %s from %s", r.Method, r.URL.Path, r.Proto, r.RemoteAddr)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","container":"%s"}`, containerID[:min(12, len(containerID))])
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("tinynginx starting on %s (container: %s)", addr, containerID[:min(12, len(containerID))])
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
