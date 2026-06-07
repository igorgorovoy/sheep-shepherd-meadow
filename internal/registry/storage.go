package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Storage manages blobs and manifests on the filesystem.
type Storage struct {
	baseDir string
	mu      sync.RWMutex
}

func NewStorage(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

func (s *Storage) Init() error {
	dirs := []string{
		filepath.Join(s.baseDir, "blobs", "sha256"),
		filepath.Join(s.baseDir, "repositories"),
		filepath.Join(s.baseDir, "uploads"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

// --- Blobs ---

func (s *Storage) HasBlob(digest string) bool {
	_, alg, hex := parseDigest(digest)
	if alg == "" {
		return false
	}
	path := filepath.Join(s.baseDir, "blobs", alg, hex)
	_, err := os.Stat(path)
	return err == nil
}

func (s *Storage) GetBlob(digest string) (io.ReadCloser, int64, error) {
	_, alg, hex := parseDigest(digest)
	path := filepath.Join(s.baseDir, "blobs", alg, hex)

	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("blob not found: %s", digest)
	}

	info, _ := f.Stat()
	return f, info.Size(), nil
}

func (s *Storage) PutBlob(r io.Reader) (string, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write to temp file, compute digest
	tmpFile, err := os.CreateTemp(filepath.Join(s.baseDir, "uploads"), "blob-*")
	if err != nil {
		return "", 0, err
	}
	tmpPath := tmpFile.Name()

	h := sha256.New()
	size, err := io.Copy(io.MultiWriter(tmpFile, h), r)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return "", 0, err
	}

	digest := "sha256:" + hex.EncodeToString(h.Sum(nil))
	_, alg, hexStr := parseDigest(digest)
	blobPath := filepath.Join(s.baseDir, "blobs", alg, hexStr)

	os.MkdirAll(filepath.Dir(blobPath), 0755)
	if err := os.Rename(tmpPath, blobPath); err != nil {
		// Cross-device: copy
		copyFilePath(tmpPath, blobPath)
		os.Remove(tmpPath)
	}

	return digest, size, nil
}

func (s *Storage) DeleteBlob(digest string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, alg, hex := parseDigest(digest)
	return os.Remove(filepath.Join(s.baseDir, "blobs", alg, hex))
}

func (s *Storage) BlobSize(digest string) (int64, error) {
	_, alg, hex := parseDigest(digest)
	info, err := os.Stat(filepath.Join(s.baseDir, "blobs", alg, hex))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// --- Manifests ---

func (s *Storage) GetManifest(repo, ref string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.manifestPath(repo, ref)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("manifest not found: %s:%s", repo, ref)
	}

	// Read content type from sidecar
	ct, _ := os.ReadFile(path + ".content-type")
	if len(ct) == 0 {
		ct = []byte("application/vnd.oci.image.manifest.v1+json")
	}

	return data, string(ct), nil
}

func (s *Storage) PutManifest(repo, ref string, data []byte, contentType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Compute digest
	h := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(h[:])

	// Store by tag
	path := s.manifestPath(repo, ref)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	os.WriteFile(path+".content-type", []byte(contentType), 0644)

	// Also store by digest
	digestPath := s.manifestPath(repo, digest)
	os.MkdirAll(filepath.Dir(digestPath), 0755)
	os.WriteFile(digestPath, data, 0644)
	os.WriteFile(digestPath+".content-type", []byte(contentType), 0644)

	return digest, nil
}

func (s *Storage) DeleteManifest(repo, ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.manifestPath(repo, ref)
	os.Remove(path + ".content-type")
	return os.Remove(path)
}

func (s *Storage) manifestPath(repo, ref string) string {
	// Sanitize ref for filesystem (digests contain ":")
	safe := strings.ReplaceAll(ref, ":", "_")
	return filepath.Join(s.baseDir, "repositories", repo, "manifests", safe)
}

// --- Catalog ---

func (s *Storage) ListRepositories() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoDir := filepath.Join(s.baseDir, "repositories")
	var repos []string

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == "manifests" {
			rel, _ := filepath.Rel(repoDir, filepath.Dir(path))
			repos = append(repos, rel)
		}
		return nil
	})

	sort.Strings(repos)
	return repos, err
}

func (s *Storage) ListTags(repo string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	manifestDir := filepath.Join(s.baseDir, "repositories", repo, "manifests")
	entries, err := os.ReadDir(manifestDir)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %s", repo)
	}

	var tags []string
	for _, e := range entries {
		name := e.Name()
		// Skip digest references and content-type sidecars
		if strings.HasPrefix(name, "sha256_") || strings.HasSuffix(name, ".content-type") {
			continue
		}
		tags = append(tags, name)
	}

	sort.Strings(tags)
	return tags, nil
}

// --- Helpers ---

func parseDigest(digest string) (full, alg, hex string) {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return digest, "", ""
	}
	return digest, parts[0], parts[1]
}

func copyFilePath(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	os.MkdirAll(filepath.Dir(dst), 0755)
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}

// RepoStats returns summary info about a repository.
type RepoStats struct {
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	TotalSize int64    `json:"total_size"`
}

func (s *Storage) GetRepoStats(repo string) (*RepoStats, error) {
	tags, err := s.ListTags(repo)
	if err != nil {
		return nil, err
	}

	var totalSize int64
	seen := make(map[string]bool)

	for _, tag := range tags {
		data, _, err := s.GetManifest(repo, tag)
		if err != nil {
			continue
		}
		var manifest struct {
			Layers []struct {
				Digest string `json:"digest"`
				Size   int64  `json:"size"`
			} `json:"layers"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		for _, l := range manifest.Layers {
			if !seen[l.Digest] {
				totalSize += l.Size
				seen[l.Digest] = true
			}
		}
	}

	return &RepoStats{
		Name:      repo,
		Tags:      tags,
		TotalSize: totalSize,
	}, nil
}
