package container

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PushImage pushes a local image to a registry.
func PushImage(img *Image, ref ImageRef, progress func(string)) error {
	scheme := "http" // Meadow is local, default to HTTP
	if ref.Registry == defaultRegistry {
		scheme = "https"
	}
	registryURL := scheme + "://" + ref.Registry

	// Check registry is reachable
	progress(fmt.Sprintf("pushing to %s/%s:%s", ref.Registry, ref.Repo, ref.Tag))

	client := &http.Client{Timeout: 5 * time.Minute}

	// Step 1: Create a tar.gz layer from the rootfs
	progress("creating layer from rootfs...")
	layerFile, err := os.CreateTemp("", "sheep-layer-*.tar.gz")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(layerFile.Name())
	defer layerFile.Close()

	layerDigest, layerSize, err := createLayer(img.RootFS, layerFile)
	if err != nil {
		return fmt.Errorf("create layer: %w", err)
	}
	layerFile.Close()

	// Step 2: Upload the layer blob
	progress(fmt.Sprintf("uploading layer %s (%s)...", layerDigest[:19], formatSize(layerSize)))

	if err := uploadBlob(client, registryURL, ref.Repo, layerFile.Name(), layerDigest); err != nil {
		return fmt.Errorf("upload layer: %w", err)
	}

	// Step 3: Create and upload the config blob
	config := createImageConfig(img)
	configJSON, _ := json.Marshal(config)
	configDigest := digestBytes(configJSON)

	configFile, _ := os.CreateTemp("", "sheep-config-*.json")
	configFile.Write(configJSON)
	configFile.Close()
	defer os.Remove(configFile.Name())

	progress("uploading config...")
	if err := uploadBlob(client, registryURL, ref.Repo, configFile.Name(), configDigest); err != nil {
		return fmt.Errorf("upload config: %w", err)
	}

	// Step 4: Create and upload manifest
	manifest := map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]any{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"digest":    configDigest,
			"size":      len(configJSON),
		},
		"layers": []map[string]any{
			{
				"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
				"digest":    layerDigest,
				"size":      layerSize,
			},
		},
	}

	manifestJSON, _ := json.Marshal(manifest)

	progress("uploading manifest...")
	if err := uploadManifest(client, registryURL, ref.Repo, ref.Tag, manifestJSON); err != nil {
		return fmt.Errorf("upload manifest: %w", err)
	}

	progress(fmt.Sprintf("pushed %s/%s:%s", ref.Registry, ref.Repo, ref.Tag))
	return nil
}

func createLayer(rootfs string, w io.Writer) (string, int64, error) {
	h := sha256.New()
	countWriter := &countingWriter{w: io.MultiWriter(w, h)}

	gw := gzip.NewWriter(countWriter)
	tw := tar.NewWriter(gw)

	err := filepath.Walk(rootfs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		rel, _ := filepath.Rel(rootfs, path)
		if rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil
		}
		header.Name = "./" + rel

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return nil
			}
			header.Linkname = link
			header.Typeflag = tar.TypeSymlink
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return nil
			}
			io.Copy(tw, f)
			f.Close()
		}

		return nil
	})

	tw.Close()
	gw.Close()

	digest := "sha256:" + hex.EncodeToString(h.Sum(nil))
	return digest, countWriter.n, err
}

func uploadBlob(client *http.Client, registryURL, repo, filePath, digest string) error {
	// Check if blob already exists
	headURL := fmt.Sprintf("%s/v2/%s/blobs/%s", registryURL, repo, digest)
	headReq, _ := http.NewRequest("HEAD", headURL, nil)
	headResp, err := client.Do(headReq)
	if err == nil && headResp.StatusCode == http.StatusOK {
		headResp.Body.Close()
		return nil // Already exists
	}
	if headResp != nil {
		headResp.Body.Close()
	}

	// Upload
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	url := fmt.Sprintf("%s/v2/%s/blobs/uploads?digest=%s", registryURL, repo, digest)
	req, err := http.NewRequest("POST", url, f)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("upload blob: HTTP %d", resp.StatusCode)
	}

	return nil
}

func uploadManifest(client *http.Client, registryURL, repo, tag string, manifest []byte) error {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repo, tag)
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(manifest)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload manifest: HTTP %d", resp.StatusCode)
	}

	return nil
}

type ociImageConfig struct {
	Created      string                 `json:"created"`
	Architecture string                 `json:"architecture"`
	OS           string                 `json:"os"`
	Config       map[string]interface{} `json:"config"`
	RootFS       ociRootFS              `json:"rootfs"`
}

type ociRootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

func createImageConfig(img *Image) ociImageConfig {
	return ociImageConfig{
		Created:      img.CreatedAt.Format(time.RFC3339),
		Architecture: "amd64",
		OS:           "linux",
		Config:       map[string]interface{}{},
		RootFS: ociRootFS{
			Type:    "layers",
			DiffIDs: []string{},
		},
	}
}

func digestBytes(data []byte) string {
	h := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(h[:])
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}
