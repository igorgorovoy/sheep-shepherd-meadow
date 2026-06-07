package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Server is the Meadow registry HTTP server.
type Server struct {
	storage *Storage
	logger  *log.Logger
	server  *http.Server
}

func NewServer(addr string, storage *Storage, logger *log.Logger) *Server {
	s := &Server{
		storage: storage,
		logger:  logger,
	}

	mux := http.NewServeMux()

	// OCI Distribution Spec endpoints
	mux.HandleFunc("/v2/", s.handler)

	// Meadow-specific endpoints
	mux.HandleFunc("/meadow/stats", s.handleStats)
	mux.HandleFunc("/meadow/stats/", s.handleRepoStats)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.logging(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Printf("meadow registry listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/")

	// GET /v2/ — version check
	if path == "" || path == "/" {
		w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
		w.WriteHeader(http.StatusOK)
		return
	}

	// GET /v2/_catalog
	if path == "_catalog" {
		s.handleCatalog(w, r)
		return
	}

	// Parse: {name}/blobs/{digest} or {name}/manifests/{ref} or {name}/tags/list or {name}/blobs/uploads
	// Name can contain slashes, so we match from the right
	if i := strings.LastIndex(path, "/blobs/uploads"); i > 0 {
		repo := path[:i]
		s.handleBlobUpload(w, r, repo)
		return
	}
	if i := strings.LastIndex(path, "/blobs/"); i > 0 {
		repo := path[:i]
		digest := path[i+len("/blobs/"):]
		s.handleBlob(w, r, repo, digest)
		return
	}
	if i := strings.LastIndex(path, "/manifests/"); i > 0 {
		repo := path[:i]
		ref := path[i+len("/manifests/"):]
		s.handleManifest(w, r, repo, ref)
		return
	}
	if strings.HasSuffix(path, "/tags/list") {
		repo := strings.TrimSuffix(path, "/tags/list")
		s.handleTags(w, r, repo)
		return
	}

	http.NotFound(w, r)
}

// --- Version ---

// --- Catalog ---

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	repos, _ := s.storage.ListRepositories()
	if repos == nil {
		repos = []string{}
	}

	respondJSON2(w, http.StatusOK, map[string]any{
		"repositories": repos,
	})
}

// --- Tags ---

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request, repo string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tags, err := s.storage.ListTags(repo)
	if err != nil {
		registryError(w, http.StatusNotFound, "NAME_UNKNOWN", "repository not found")
		return
	}
	if tags == nil {
		tags = []string{}
	}

	respondJSON2(w, http.StatusOK, map[string]any{
		"name": repo,
		"tags": tags,
	})
}

// --- Blobs ---

func (s *Server) handleBlob(w http.ResponseWriter, r *http.Request, repo, digest string) {
	switch r.Method {
	case http.MethodHead:
		if !s.storage.HasBlob(digest) {
			registryError(w, http.StatusNotFound, "BLOB_UNKNOWN", "blob not found")
			return
		}
		size, _ := s.storage.BlobSize(digest)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Set("Docker-Content-Digest", digest)
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:
		rc, size, err := s.storage.GetBlob(digest)
		if err != nil {
			registryError(w, http.StatusNotFound, "BLOB_UNKNOWN", "blob not found")
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Set("Docker-Content-Digest", digest)
		w.Header().Set("Content-Type", "application/octet-stream")
		io.Copy(w, rc)

	case http.MethodDelete:
		if err := s.storage.DeleteBlob(digest); err != nil {
			registryError(w, http.StatusNotFound, "BLOB_UNKNOWN", "blob not found")
			return
		}
		w.WriteHeader(http.StatusAccepted)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBlobUpload(w http.ResponseWriter, r *http.Request, repo string) {
	switch r.Method {
	case http.MethodPost:
		// Monolithic upload: client sends everything in one request
		if r.Body == nil {
			// Initiate upload — return location for PUT
			// For simplicity, we support monolithic uploads via PUT to /v2/{name}/blobs/uploads?digest=...
			w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads", repo))
			w.Header().Set("Docker-Upload-UUID", "single")
			w.WriteHeader(http.StatusAccepted)
			return
		}

		// Check if digest is provided (monolithic upload)
		digest := r.URL.Query().Get("digest")
		if digest != "" {
			s.doBlobUpload(w, r, repo, digest)
			return
		}

		// Start chunked upload
		w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads", repo))
		w.Header().Set("Docker-Upload-UUID", "single")
		w.WriteHeader(http.StatusAccepted)

	case http.MethodPut:
		digest := r.URL.Query().Get("digest")
		if digest == "" {
			registryError(w, http.StatusBadRequest, "DIGEST_INVALID", "digest required")
			return
		}
		s.doBlobUpload(w, r, repo, digest)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) doBlobUpload(w http.ResponseWriter, r *http.Request, repo, expectedDigest string) {
	actualDigest, size, err := s.storage.PutBlob(r.Body)
	if err != nil {
		registryError(w, http.StatusInternalServerError, "BLOB_UPLOAD_INVALID", err.Error())
		return
	}

	if expectedDigest != "" && actualDigest != expectedDigest {
		s.storage.DeleteBlob(actualDigest)
		registryError(w, http.StatusBadRequest, "DIGEST_INVALID",
			fmt.Sprintf("expected %s, got %s", expectedDigest, actualDigest))
		return
	}

	w.Header().Set("Docker-Content-Digest", actualDigest)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", repo, actualDigest))
	w.WriteHeader(http.StatusCreated)

	s.logger.Printf("blob uploaded: %s (%d bytes)", shortDigest2(actualDigest), size)
}

// --- Manifests ---

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request, repo, ref string) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		data, ct, err := s.storage.GetManifest(repo, ref)
		if err != nil {
			registryError(w, http.StatusNotFound, "MANIFEST_UNKNOWN", "manifest not found")
			return
		}

		w.Header().Set("Content-Type", ct)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

		if r.Method == http.MethodGet {
			w.Write(data)
		}

	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			registryError(w, http.StatusBadRequest, "MANIFEST_INVALID", "failed to read body")
			return
		}

		ct := r.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/vnd.oci.image.manifest.v1+json"
		}

		digest, err := s.storage.PutManifest(repo, ref, body, ct)
		if err != nil {
			registryError(w, http.StatusInternalServerError, "MANIFEST_INVALID", err.Error())
			return
		}

		w.Header().Set("Docker-Content-Digest", digest)
		w.Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s", repo, ref))
		w.WriteHeader(http.StatusCreated)

		s.logger.Printf("manifest pushed: %s:%s (%s)", repo, ref, shortDigest2(digest))

	case http.MethodDelete:
		if err := s.storage.DeleteManifest(repo, ref); err != nil {
			registryError(w, http.StatusNotFound, "MANIFEST_UNKNOWN", "manifest not found")
			return
		}
		w.WriteHeader(http.StatusAccepted)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Meadow stats ---

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	repos, _ := s.storage.ListRepositories()

	var stats []RepoStats
	for _, repo := range repos {
		rs, err := s.storage.GetRepoStats(repo)
		if err != nil {
			continue
		}
		stats = append(stats, *rs)
	}

	respondJSON2(w, http.StatusOK, map[string]any{
		"registry":     "meadow",
		"version":      "v0.1.0",
		"repositories": len(repos),
		"details":      stats,
	})
}

func (s *Server) handleRepoStats(w http.ResponseWriter, r *http.Request) {
	repo := strings.TrimPrefix(r.URL.Path, "/meadow/stats/")
	if repo == "" {
		http.NotFound(w, r)
		return
	}

	stats, err := s.storage.GetRepoStats(repo)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	respondJSON2(w, http.StatusOK, stats)
}

// --- Helpers ---

func respondJSON2(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func registryError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"errors": []map[string]string{
			{"code": code, "message": message},
		},
	})
}

func shortDigest2(digest string) string {
	if i := strings.Index(digest, ":"); i >= 0 {
		d := digest[i+1:]
		if len(d) > 12 {
			return d[:12]
		}
	}
	return digest
}
