package container

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		input    string
		registry string
		repo     string
		tag      string
	}{
		{"nginx", "registry-1.docker.io", "library/nginx", "latest"},
		{"nginx:1.25", "registry-1.docker.io", "library/nginx", "1.25"},
		{"alpine:3.19", "registry-1.docker.io", "library/alpine", "3.19"},
		{"myuser/myapp", "registry-1.docker.io", "myuser/myapp", "latest"},
		{"myuser/myapp:v2", "registry-1.docker.io", "myuser/myapp", "v2"},
		{"ghcr.io/user/repo:latest", "ghcr.io", "user/repo", "latest"},
		{"registry.example.com/org/app:v1", "registry.example.com", "org/app", "v1"},
	}

	for _, tt := range tests {
		ref := ParseImageRef(tt.input)
		if ref.Registry != tt.registry {
			t.Errorf("ParseImageRef(%q).Registry = %q, want %q", tt.input, ref.Registry, tt.registry)
		}
		if ref.Repo != tt.repo {
			t.Errorf("ParseImageRef(%q).Repo = %q, want %q", tt.input, ref.Repo, tt.repo)
		}
		if ref.Tag != tt.tag {
			t.Errorf("ParseImageRef(%q).Tag = %q, want %q", tt.input, ref.Tag, tt.tag)
		}
	}
}

func TestPullAlpine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tmpDir := t.TempDir()
	imgDir := filepath.Join(tmpDir, "images")
	os.MkdirAll(imgDir, 0755)

	imgMgr := NewImageManager(imgDir)
	imgMgr.Init()

	client := NewRegistryClient()
	ref := ParseImageRef("alpine:3.19")

	img, err := client.Pull(ref, imgMgr, func(msg string) {
		t.Log(msg)
	})
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	if img.Name != "alpine" {
		t.Errorf("Name = %q, want alpine", img.Name)
	}
	if img.Tag != "3.19" {
		t.Errorf("Tag = %q, want 3.19", img.Tag)
	}

	// Check rootfs has expected directories
	for _, dir := range []string{"bin", "etc", "usr"} {
		path := filepath.Join(img.RootFS, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s in rootfs", dir)
		}
	}

	// Check that /bin has executables (alpine uses busybox symlinks)
	binDir := filepath.Join(img.RootFS, "bin")
	entries, err := os.ReadDir(binDir)
	if err != nil {
		t.Fatalf("read /bin: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected files in /bin")
	}
	t.Logf("/bin has %d entries", len(entries))

	t.Logf("Image pulled: %s:%s, size: %d bytes, rootfs: %s", img.Name, img.Tag, img.Size, img.RootFS)
}
