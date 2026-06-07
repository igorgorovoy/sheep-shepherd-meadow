package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestServer(t *testing.T) (*httptest.Server, *Storage) {
	t.Helper()
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)
	if err := storage.Init(); err != nil {
		t.Fatal(err)
	}

	logger := log.New(io.Discard, "", 0)
	srv := NewServer(":0", storage, logger)

	// Use httptest to get a real server
	ts := httptest.NewServer(srv.server.Handler)
	return ts, storage
}

func TestVersionCheck(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if v := resp.Header.Get("Docker-Distribution-API-Version"); v != "registry/2.0" {
		t.Errorf("expected registry/2.0, got %s", v)
	}
}

func TestPushPullFlow(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	repo := "myapp"
	tag := "v1"

	// 1. Upload a blob
	blobContent := []byte("hello this is a layer")
	h := sha256.Sum256(blobContent)
	blobDigest := "sha256:" + hex.EncodeToString(h[:])

	url := fmt.Sprintf("%s/v2/%s/blobs/uploads?digest=%s", ts.URL, repo, blobDigest)
	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(blobContent))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("blob upload: expected 201, got %d", resp.StatusCode)
	}
	t.Logf("blob uploaded: %s", blobDigest)

	// 2. Check blob exists
	headURL := fmt.Sprintf("%s/v2/%s/blobs/%s", ts.URL, repo, blobDigest)
	headReq, _ := http.NewRequest("HEAD", headURL, nil)
	headResp, err := http.DefaultClient.Do(headReq)
	if err != nil {
		t.Fatal(err)
	}
	headResp.Body.Close()
	if headResp.StatusCode != http.StatusOK {
		t.Fatalf("blob HEAD: expected 200, got %d", headResp.StatusCode)
	}

	// 3. Upload manifest
	manifest := map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]any{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"digest":    blobDigest,
			"size":      len(blobContent),
		},
		"layers": []map[string]any{
			{
				"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
				"digest":    blobDigest,
				"size":      len(blobContent),
			},
		},
	}
	manifestJSON, _ := json.Marshal(manifest)

	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", ts.URL, repo, tag)
	putReq, _ := http.NewRequest("PUT", manifestURL, bytes.NewReader(manifestJSON))
	putReq.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatal(err)
	}
	putResp.Body.Close()
	if putResp.StatusCode != http.StatusCreated {
		t.Fatalf("manifest PUT: expected 201, got %d", putResp.StatusCode)
	}
	t.Logf("manifest pushed: %s:%s", repo, tag)

	// 4. Pull manifest back
	getResp, err := http.Get(manifestURL)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("manifest GET: expected 200, got %d", getResp.StatusCode)
	}
	body, _ := io.ReadAll(getResp.Body)
	if !bytes.Equal(body, manifestJSON) {
		t.Error("manifest content mismatch")
	}
	t.Log("manifest pulled successfully")

	// 5. Pull blob back
	blobResp, err := http.Get(fmt.Sprintf("%s/v2/%s/blobs/%s", ts.URL, repo, blobDigest))
	if err != nil {
		t.Fatal(err)
	}
	defer blobResp.Body.Close()
	blobBody, _ := io.ReadAll(blobResp.Body)
	if !bytes.Equal(blobBody, blobContent) {
		t.Error("blob content mismatch")
	}
	t.Log("blob pulled successfully")

	// 6. Check catalog
	catalogResp, err := http.Get(ts.URL + "/v2/_catalog")
	if err != nil {
		t.Fatal(err)
	}
	defer catalogResp.Body.Close()
	var catalog map[string]any
	json.NewDecoder(catalogResp.Body).Decode(&catalog)
	repos := catalog["repositories"].([]any)
	if len(repos) != 1 || repos[0] != "myapp" {
		t.Errorf("unexpected catalog: %v", repos)
	}
	t.Logf("catalog: %v", repos)

	// 7. Check tags
	tagsResp, err := http.Get(fmt.Sprintf("%s/v2/%s/tags/list", ts.URL, repo))
	if err != nil {
		t.Fatal(err)
	}
	defer tagsResp.Body.Close()
	var tagsResult map[string]any
	json.NewDecoder(tagsResp.Body).Decode(&tagsResult)
	tags := tagsResult["tags"].([]any)
	if len(tags) != 1 || tags[0] != "v1" {
		t.Errorf("unexpected tags: %v", tags)
	}
	t.Logf("tags: %v", tags)

	// 8. Check meadow stats
	statsResp, err := http.Get(ts.URL + "/meadow/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer statsResp.Body.Close()
	var stats map[string]any
	json.NewDecoder(statsResp.Body).Decode(&stats)
	if stats["registry"] != "meadow" {
		t.Errorf("expected meadow, got %v", stats["registry"])
	}
	t.Logf("stats: %v", stats)
}

func TestEmptyCatalog(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v2/_catalog")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var catalog map[string]any
	json.NewDecoder(resp.Body).Decode(&catalog)
	repos := catalog["repositories"].([]any)
	if len(repos) != 0 {
		t.Errorf("expected empty catalog, got %v", repos)
	}
}

func TestBlobNotFound(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v2/myapp/blobs/sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestManifestNotFound(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v2/nonexistent/manifests/latest")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestMultipleTags(t *testing.T) {
	ts, storage := setupTestServer(t)
	defer ts.Close()

	manifest := []byte(`{"schemaVersion":2}`)

	// Push same manifest with two tags
	storage.PutManifest("myapp", "v1", manifest, "application/json")
	storage.PutManifest("myapp", "v2", manifest, "application/json")
	storage.PutManifest("myapp", "latest", manifest, "application/json")

	tags, err := storage.ListTags("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(tags), tags)
	}

	// Verify via API
	resp, _ := http.Get(fmt.Sprintf("%s/v2/myapp/tags/list", ts.URL))
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	apiTags := result["tags"].([]any)
	if len(apiTags) != 3 {
		t.Errorf("API: expected 3 tags, got %d", len(apiTags))
	}
}

// Ensure the storage dir environment is clean
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
