package container

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ImagesDir = "/var/lib/sheep/images"
)

type Image struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tag       string    `json:"tag"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	RootFS    string    `json:"rootfs"`
}

type ImageManager struct {
	baseDir string
}

func NewImageManager(baseDir string) *ImageManager {
	if baseDir == "" {
		baseDir = ImagesDir
	}
	return &ImageManager{baseDir: baseDir}
}

func (im *ImageManager) Init() error {
	return os.MkdirAll(im.baseDir, 0755)
}

// Import imports a rootfs tarball as an image.
func (im *ImageManager) Import(name, tag, tarPath string) (*Image, error) {
	if tag == "" {
		tag = "latest"
	}

	id := GenerateID()
	rootfs := filepath.Join(im.baseDir, id, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		return nil, fmt.Errorf("create image dir: %w", err)
	}

	f, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("open tar: %w", err)
	}
	defer f.Close()

	if err := extractTar(f, rootfs); err != nil {
		os.RemoveAll(filepath.Join(im.baseDir, id))
		return nil, fmt.Errorf("extract tar: %w", err)
	}

	img := &Image{
		ID:        id,
		Name:      name,
		Tag:       tag,
		CreatedAt: time.Now(),
		RootFS:    rootfs,
	}

	// Calculate size
	var size int64
	filepath.Walk(rootfs, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	img.Size = size

	if err := im.saveMetadata(id, img); err != nil {
		return nil, err
	}

	return img, nil
}

// Bootstrap creates a minimal image from the host's filesystem for testing.
func (im *ImageManager) Bootstrap(name string) (*Image, error) {
	id := GenerateID()
	rootfs := filepath.Join(im.baseDir, id, "rootfs")

	dirs := []string{
		"bin", "sbin", "usr/bin", "usr/sbin", "usr/lib",
		"lib", "lib64", "etc", "dev", "proc", "sys",
		"tmp", "var", "run", "home", "root",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(rootfs, d), 0755); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	// Copy essential binaries
	binaries := []string{"/bin/sh", "/bin/ls", "/bin/cat", "/bin/echo", "/bin/ps", "/bin/mkdir", "/bin/sleep"}
	for _, bin := range binaries {
		if _, err := os.Stat(bin); err == nil {
			dst := filepath.Join(rootfs, bin)
			if err := copyFile(bin, dst); err != nil {
				// Non-fatal: skip missing binaries
				continue
			}
		}
	}

	// Minimal /etc files
	os.WriteFile(filepath.Join(rootfs, "etc/hostname"), []byte("sheep\n"), 0644)
	os.WriteFile(filepath.Join(rootfs, "etc/hosts"), []byte("127.0.0.1 localhost\n"), 0644)
	os.WriteFile(filepath.Join(rootfs, "etc/resolv.conf"), []byte("nameserver 8.8.8.8\n"), 0644)

	img := &Image{
		ID:        id,
		Name:      name,
		Tag:       "latest",
		CreatedAt: time.Now(),
		RootFS:    rootfs,
	}

	if err := im.saveMetadata(id, img); err != nil {
		return nil, err
	}

	return img, nil
}

func (im *ImageManager) Get(name, tag string) (*Image, error) {
	if tag == "" {
		tag = "latest"
	}

	entries, err := os.ReadDir(im.baseDir)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		img, err := im.loadMetadata(e.Name())
		if err != nil {
			continue
		}
		if img.Name == name && img.Tag == tag {
			return img, nil
		}
	}

	return nil, fmt.Errorf("image %s:%s not found", name, tag)
}

func (im *ImageManager) List() ([]*Image, error) {
	entries, err := os.ReadDir(im.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var images []*Image
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		img, err := im.loadMetadata(e.Name())
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	return images, nil
}

func (im *ImageManager) Tag(srcID, newName, newTag string) (*Image, error) {
	src, err := im.loadMetadata(srcID)
	if err != nil {
		return nil, fmt.Errorf("source image not found: %w", err)
	}

	// Create a new image entry pointing to the same rootfs
	newID := GenerateID()
	newDir := filepath.Join(im.baseDir, newID)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return nil, err
	}

	// Symlink rootfs to save space
	os.Symlink(src.RootFS, filepath.Join(newDir, "rootfs"))

	tagged := &Image{
		ID:        newID,
		Name:      newName,
		Tag:       newTag,
		Size:      src.Size,
		CreatedAt: time.Now(),
		RootFS:    src.RootFS,
	}

	if err := im.saveMetadata(newID, tagged); err != nil {
		return nil, err
	}

	return tagged, nil
}

func (im *ImageManager) Remove(id string) error {
	return os.RemoveAll(filepath.Join(im.baseDir, id))
}

func (im *ImageManager) saveMetadata(id string, img *Image) error {
	data, err := json.Marshal(img)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(im.baseDir, id, "manifest.json"), data, 0644)
}

func (im *ImageManager) loadMetadata(id string) (*Image, error) {
	data, err := os.ReadFile(filepath.Join(im.baseDir, id, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var img Image
	if err := json.Unmarshal(data, &img); err != nil {
		return nil, err
	}
	return &img, nil
}

func extractTar(r io.Reader, dst string) error {
	// Try gzip first, fall back to plain tar
	gr, err := gzip.NewReader(r)
	var tr *tar.Reader
	if err != nil {
		// Not gzip, try as plain tar
		if rs, ok := r.(io.ReadSeeker); ok {
			rs.Seek(0, io.SeekStart)
		}
		tr = tar.NewReader(r)
	} else {
		defer gr.Close()
		tr = tar.NewReader(gr)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			os.Symlink(hdr.Linkname, target)
		case tar.TypeLink:
			os.Link(filepath.Join(dst, hdr.Linkname), target)
		}
	}

	return nil
}

func newTarReader(r io.Reader) *tar.Reader {
	return tar.NewReader(r)
}

func extractTarReader(tr *tar.Reader, dst string) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Handle whiteout files (OCI layer deletions)
		name := hdr.Name
		if strings.HasPrefix(filepath.Base(name), ".wh.") {
			// Whiteout: delete the corresponding file
			target := filepath.Join(dst, filepath.Dir(name), strings.TrimPrefix(filepath.Base(name), ".wh."))
			os.RemoveAll(target)
			continue
		}

		target := filepath.Join(dst, name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, os.FileMode(hdr.Mode))
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				continue
			}
			io.Copy(f, tr)
			f.Close()
		case tar.TypeSymlink:
			os.Remove(target)
			os.Symlink(hdr.Linkname, target)
		case tar.TypeLink:
			os.Remove(target)
			os.Link(filepath.Join(dst, hdr.Linkname), target)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	info, err := sf.Stat()
	if err != nil {
		return err
	}

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}
