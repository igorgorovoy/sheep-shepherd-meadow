package container

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPullNginx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tmpDir := t.TempDir()
	imgDir := filepath.Join(tmpDir, "images")
	os.MkdirAll(imgDir, 0755)

	imgMgr := NewImageManager(imgDir)
	imgMgr.Init()

	client := NewRegistryClient()
	ref := ParseImageRef("nginx:alpine")

	img, err := client.Pull(ref, imgMgr, func(msg string) {
		t.Log(msg)
	})
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	if img.Name != "nginx" {
		t.Errorf("Name = %q, want nginx", img.Name)
	}

	// Check nginx binary exists
	nginxBin := filepath.Join(img.RootFS, "usr", "sbin", "nginx")
	if _, err := os.Stat(nginxBin); os.IsNotExist(err) {
		t.Error("expected /usr/sbin/nginx in rootfs")
	} else {
		t.Log("found /usr/sbin/nginx")
	}

	// Check nginx config exists
	nginxConf := filepath.Join(img.RootFS, "etc", "nginx", "nginx.conf")
	if _, err := os.Stat(nginxConf); os.IsNotExist(err) {
		t.Error("expected /etc/nginx/nginx.conf in rootfs")
	} else {
		t.Log("found /etc/nginx/nginx.conf")
	}

	t.Logf("Image pulled: %s:%s, size: %d bytes", img.Name, img.Tag, img.Size)
}
