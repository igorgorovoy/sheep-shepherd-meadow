package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	BaseDir       = "/var/lib/sheep"
	ContainersDir = "/var/lib/sheep/containers"
	OverlayDir    = "/var/lib/sheep/overlay"
)

type Manager struct {
	mu         sync.RWMutex
	containers map[string]*Container
	baseDir    string
	images     *ImageManager
}

func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		baseDir = BaseDir
	}
	return &Manager{
		containers: make(map[string]*Container),
		baseDir:    baseDir,
		images:     NewImageManager(filepath.Join(baseDir, "images")),
	}
}

func (m *Manager) Init() error {
	dirs := []string{
		filepath.Join(m.baseDir, "containers"),
		filepath.Join(m.baseDir, "overlay"),
		filepath.Join(m.baseDir, "images"),
		filepath.Join(m.baseDir, "network"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	if err := m.images.Init(); err != nil {
		return err
	}

	return m.loadExisting()
}

func (m *Manager) Images() *ImageManager {
	return m.images
}

func (m *Manager) Create(opts RunOpts) (*Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := GenerateID()
	if opts.Name == "" {
		opts.Name = "sheep-" + ShortID(id)
	}

	// Check name uniqueness
	for _, c := range m.containers {
		if c.Name == opts.Name {
			return nil, fmt.Errorf("container name %q already in use", opts.Name)
		}
	}

	// Resolve image
	img, err := m.images.Get(opts.Image, "latest")
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Set up overlay filesystem
	rootfs, err := m.setupOverlay(id, img.RootFS)
	if err != nil {
		return nil, fmt.Errorf("setup overlay: %w", err)
	}

	hostname := opts.Config.Hostname
	if hostname == "" {
		hostname = ShortID(id)
	}

	c := &Container{
		ID:        id,
		Name:      opts.Name,
		Image:     opts.Image,
		Command:   opts.Command,
		State:     StateCreated,
		CreatedAt: time.Now(),
		Config: Config{
			Hostname:  hostname,
			Env:       opts.Config.Env,
			WorkDir:   opts.Config.WorkDir,
			Memory:    opts.Config.Memory,
			CPUShares: opts.Config.CPUShares,
			CPUQuota:  opts.Config.CPUQuota,
			PidsLimit: opts.Config.PidsLimit,
		},
		RootFS: rootfs,
		Mounts: opts.Mounts,
	}

	m.containers[id] = c
	if err := m.saveState(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (m *Manager) Start(id string) error {
	m.mu.Lock()
	c, ok := m.containers[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("container %s not found", ShortID(id))
	}
	if c.State == StateRunning {
		m.mu.Unlock()
		return fmt.Errorf("container %s already running", ShortID(id))
	}
	m.mu.Unlock()

	pid, err := startContainer(c)
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	m.mu.Lock()
	c.Pid = pid
	c.State = StateRunning
	c.StartedAt = time.Now()
	m.mu.Unlock()

	return m.saveState(c)
}

func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	c, ok := m.containers[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("container %s not found", ShortID(id))
	}
	if c.State != StateRunning {
		m.mu.Unlock()
		return fmt.Errorf("container %s not running", ShortID(id))
	}
	m.mu.Unlock()

	exitCode, err := stopContainer(c)
	if err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	m.mu.Lock()
	c.State = StateStopped
	c.ExitCode = exitCode
	c.StoppedAt = time.Now()
	c.Pid = 0
	m.mu.Unlock()

	return m.saveState(c)
}

func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	c, ok := m.containers[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("container %s not found", ShortID(id))
	}
	if c.State == StateRunning {
		m.mu.Unlock()
		return fmt.Errorf("container %s is running, stop it first", ShortID(id))
	}
	delete(m.containers, id)
	m.mu.Unlock()

	m.cleanupOverlay(id)
	os.RemoveAll(filepath.Join(m.baseDir, "containers", id))
	return nil
}

func (m *Manager) Get(id string) (*Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try exact match
	if c, ok := m.containers[id]; ok {
		return c, nil
	}

	// Try prefix match
	for cid, c := range m.containers {
		if len(id) >= 4 && len(cid) >= len(id) && cid[:len(id)] == id {
			return c, nil
		}
	}

	// Try name match
	for _, c := range m.containers {
		if c.Name == id {
			return c, nil
		}
	}

	return nil, fmt.Errorf("container %s not found", id)
}

func (m *Manager) List(all bool) []*Container {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Container
	for _, c := range m.containers {
		if all || c.State == StateRunning {
			result = append(result, c)
		}
	}
	return result
}

func (m *Manager) setupOverlay(id, lowerDir string) (string, error) {
	overlayBase := filepath.Join(m.baseDir, "overlay", id)
	upper := filepath.Join(overlayBase, "upper")
	work := filepath.Join(overlayBase, "work")
	merged := filepath.Join(overlayBase, "merged")

	for _, d := range []string{upper, work, merged} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", err
		}
	}

	if err := mountOverlay(lowerDir, upper, work, merged); err != nil {
		// Fallback: just copy rootfs if overlay not supported
		return copyRootFS(lowerDir, merged)
	}

	return merged, nil
}

func (m *Manager) cleanupOverlay(id string) {
	overlayBase := filepath.Join(m.baseDir, "overlay", id)
	merged := filepath.Join(overlayBase, "merged")
	unmountOverlay(merged)
	os.RemoveAll(overlayBase)
}

func copyRootFS(src, dst string) (string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		info, err := e.Info()
		if err != nil {
			continue
		}

		if info.IsDir() {
			os.MkdirAll(dstPath, info.Mode())
			copyRootFS(srcPath, dstPath)
		} else if info.Mode().IsRegular() {
			copyFile(srcPath, dstPath)
		}
	}
	return dst, nil
}

func (m *Manager) saveState(c *Container) error {
	dir := filepath.Join(m.baseDir, "containers", c.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "state.json"), data, 0644)
}

func (m *Manager) loadExisting() error {
	dir := filepath.Join(m.baseDir, "containers")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name(), "state.json"))
		if err != nil {
			continue
		}
		var c Container
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		// Check if previously running containers are still alive
		if c.State == StateRunning && c.Pid > 0 {
			if !isProcessAlive(c.Pid) {
				c.State = StateStopped
				c.Pid = 0
			}
		}
		m.containers[c.ID] = &c
	}

	return nil
}

func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if process exists without actually sending a signal
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
