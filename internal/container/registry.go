package container

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultRegistry = "registry-1.docker.io"
	dockerAuthURL   = "https://auth.docker.io/token"
	dockerService   = "registry.docker.io"
)

// ImageRef represents a parsed container image reference.
type ImageRef struct {
	Registry string
	Repo     string
	Tag      string
}

// ParseImageRef parses an image string like "nginx", "nginx:1.25", "ghcr.io/user/repo:tag".
func ParseImageRef(s string) ImageRef {
	ref := ImageRef{Tag: "latest"}

	// Split tag
	if i := strings.LastIndex(s, ":"); i > 0 && !strings.Contains(s[i:], "/") {
		ref.Tag = s[i+1:]
		s = s[:i]
	}

	parts := strings.Split(s, "/")
	switch {
	case len(parts) == 1:
		// "nginx" -> registry-1.docker.io/library/nginx
		ref.Registry = defaultRegistry
		ref.Repo = "library/" + parts[0]
	case len(parts) == 2 && !strings.Contains(parts[0], "."):
		// "user/repo" -> registry-1.docker.io/user/repo
		ref.Registry = defaultRegistry
		ref.Repo = s
	default:
		// "ghcr.io/user/repo" or "registry.example.com/path/image"
		ref.Registry = parts[0]
		ref.Repo = strings.Join(parts[1:], "/")
	}

	return ref
}

func (r ImageRef) String() string {
	if r.Registry == defaultRegistry {
		repo := r.Repo
		if strings.HasPrefix(repo, "library/") {
			repo = strings.TrimPrefix(repo, "library/")
		}
		return repo + ":" + r.Tag
	}
	return r.Registry + "/" + r.Repo + ":" + r.Tag
}

// RegistryClient pulls OCI/Docker images from container registries.
type RegistryClient struct {
	client *http.Client
}

func NewRegistryClient() *RegistryClient {
	return &RegistryClient{
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Pull downloads an image from a registry and extracts it into the image store.
func (rc *RegistryClient) Pull(ref ImageRef, imgMgr *ImageManager, progress func(string)) (*Image, error) {
	scheme := "https"
	registryURL := scheme + "://" + ref.Registry

	// Get auth token (Docker Hub specific, other registries may differ)
	token, err := rc.getToken(ref)
	if err != nil {
		// Try without auth for registries that don't need it
		token = ""
	}

	progress(fmt.Sprintf("pulling manifest for %s", ref.String()))

	// Fetch manifest — try manifest list first (multi-arch), then single manifest
	manifest, err := rc.getManifest(registryURL, ref, token)
	if err != nil {
		return nil, fmt.Errorf("get manifest: %w", err)
	}

	// Prepare image directory
	id := GenerateID()
	rootfs := filepath.Join(imgMgr.baseDir, id, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		return nil, err
	}

	// Download and extract layers
	for i, layer := range manifest.Layers {
		progress(fmt.Sprintf("pulling layer %d/%d %s", i+1, len(manifest.Layers), shortDigest(layer.Digest)))

		if err := rc.pullLayer(registryURL, ref, token, layer.Digest, rootfs); err != nil {
			os.RemoveAll(filepath.Join(imgMgr.baseDir, id))
			return nil, fmt.Errorf("pull layer %d: %w", i+1, err)
		}
	}

	// Calculate size
	var size int64
	filepath.Walk(rootfs, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	displayName := ref.Repo
	if strings.HasPrefix(displayName, "library/") {
		displayName = strings.TrimPrefix(displayName, "library/")
	}

	img := &Image{
		ID:        id,
		Name:      displayName,
		Tag:       ref.Tag,
		Size:      size,
		CreatedAt: time.Now(),
		RootFS:    rootfs,
	}

	if err := imgMgr.saveMetadata(id, img); err != nil {
		return nil, err
	}

	progress(fmt.Sprintf("pulled %s (%s)", ref.String(), formatSize(size)))
	return img, nil
}

func (rc *RegistryClient) getToken(ref ImageRef) (string, error) {
	// Docker Hub auth
	if ref.Registry == defaultRegistry {
		url := fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", dockerAuthURL, dockerService, ref.Repo)
		resp, err := rc.client.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var tokenResp struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return "", err
		}
		return tokenResp.Token, nil
	}

	// For other registries, try anonymous access
	return "", nil
}

// manifestResponse holds the relevant fields from a Docker/OCI manifest.
type manifestResponse struct {
	MediaType string          `json:"mediaType"`
	Layers    []manifestLayer `json:"layers"`
	FSLayers  []fsLayer       `json:"fsLayers"` // v1 compat
}

type manifestLayer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type fsLayer struct {
	BlobSum string `json:"blobSum"`
}

// manifestList is a multi-arch manifest index.
type manifestList struct {
	MediaType string             `json:"mediaType"`
	Manifests []manifestPlatform `json:"manifests"`
}

type manifestPlatform struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Platform  struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
	} `json:"platform"`
}

func (rc *RegistryClient) getManifest(registryURL string, ref ImageRef, token string) (*manifestResponse, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, ref.Repo, ref.Tag)

	// Try OCI manifest list first (for multi-arch images)
	accepts := []string{
		"application/vnd.oci.image.index.v1+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
	}

	body, mediaType, err := rc.registryGet(url, token, strings.Join(accepts, ", "))
	if err != nil {
		return nil, err
	}

	// Check if this is a manifest list (multi-arch)
	if strings.Contains(mediaType, "list") || strings.Contains(mediaType, "index") {
		var ml manifestList
		if err := json.Unmarshal(body, &ml); err == nil && len(ml.Manifests) > 0 {
			// Find linux/amd64 manifest
			digest := ""
			for _, m := range ml.Manifests {
				if m.Platform.OS == "linux" && m.Platform.Architecture == "amd64" {
					digest = m.Digest
					break
				}
			}
			if digest == "" {
				// Fall back to first manifest
				digest = ml.Manifests[0].Digest
			}

			// Fetch the actual manifest by digest
			url = fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, ref.Repo, digest)
			singleAccepts := "application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json"
			body, _, err = rc.registryGet(url, token, singleAccepts)
			if err != nil {
				return nil, fmt.Errorf("get platform manifest: %w", err)
			}
		}
	}

	var manifest manifestResponse
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	// Handle v1 manifests (fsLayers)
	if len(manifest.Layers) == 0 && len(manifest.FSLayers) > 0 {
		for _, fl := range manifest.FSLayers {
			manifest.Layers = append(manifest.Layers, manifestLayer{
				Digest: fl.BlobSum,
			})
		}
	}

	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("manifest has no layers")
	}

	return &manifest, nil
}

func (rc *RegistryClient) pullLayer(registryURL string, ref ImageRef, token, digest, rootfs string) error {
	url := fmt.Sprintf("%s/v2/%s/blobs/%s", registryURL, ref.Repo, digest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching layer", resp.StatusCode)
	}

	return extractLayer(resp.Body, rootfs)
}

func extractLayer(r io.Reader, dst string) error {
	// Try gzip decompression
	gr, err := gzip.NewReader(r)
	if err != nil {
		// Might not be gzipped, but layers from registries always are
		return fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	return extractTarStream(gr, dst)
}

func extractTarStream(r io.Reader, dst string) error {
	// Use the existing extractTar but from a reader directly
	// tar.NewReader on the decompressed stream
	tr := newTarReader(r)
	return extractTarReader(tr, dst)
}

func (rc *RegistryClient) registryGet(url, token, accept string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateStr(string(body), 200))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	ct := resp.Header.Get("Content-Type")
	return body, ct, nil
}

func shortDigest(digest string) string {
	if i := strings.Index(digest, ":"); i >= 0 {
		d := digest[i+1:]
		if len(d) > 12 {
			return d[:12]
		}
		return d
	}
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
