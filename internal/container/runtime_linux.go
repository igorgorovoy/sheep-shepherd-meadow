//go:build linux

package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// startContainer spawns an isolated process using Linux namespaces.
// It re-execs the current binary with a special "init" argument
// so that namespace setup happens in the child process.
func startContainer(c *Container) (int, error) {
	// Re-exec ourselves with the container init command
	cmd := reexecCommand(c)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start namespaced process: %w", err)
	}

	pid := cmd.Process.Pid

	// Set up cgroups for the container
	if err := setupCgroups(c, pid); err != nil {
		cmd.Process.Kill()
		return 0, fmt.Errorf("setup cgroups: %w", err)
	}

	// Set up networking
	if err := setupNetworkForContainer(c, pid); err != nil {
		// Non-fatal: container runs without network
		fmt.Fprintf(os.Stderr, "warning: network setup failed: %v\n", err)
	}

	// Detach: don't wait
	go cmd.Wait()

	return pid, nil
}

func stopContainer(c *Container) (int, error) {
	if c.Pid <= 0 {
		return 0, nil
	}

	proc, err := os.FindProcess(c.Pid)
	if err != nil {
		return 0, nil
	}

	// Try graceful shutdown first
	proc.Signal(syscall.SIGTERM)

	// Then force kill after brief wait
	// In production, we'd use a timeout here
	proc.Signal(syscall.SIGKILL)
	state, _ := proc.Wait()

	cleanupCgroups(c)

	if state != nil {
		return state.ExitCode(), nil
	}
	return 0, nil
}

func reexecCommand(c *Container) *exec.Cmd {
	// Pass container config via environment to the child process
	self, _ := os.Executable()
	cmd := exec.Command(self, append([]string{"init", "--rootfs", c.RootFS, "--hostname", c.Config.Hostname, "--"}, c.Command...)...)
	cmd.Env = append(c.Config.Env, fmt.Sprintf("SHEEP_CONTAINER_ID=%s", c.ID))
	return cmd
}

// ContainerInit is called inside the namespaced child process.
// It sets up the filesystem, hostname, etc., then execs the target command.
func ContainerInit(rootfs, hostname string, command []string) error {
	// Set hostname
	if hostname != "" {
		if err := syscall.Sethostname([]byte(hostname)); err != nil {
			return fmt.Errorf("sethostname: %w", err)
		}
	}

	// Mount proc inside the new rootfs
	procPath := filepath.Join(rootfs, "proc")
	os.MkdirAll(procPath, 0755)

	// Mount essential filesystems
	mounts := []struct {
		source string
		target string
		fstype string
		flags  uintptr
		data   string
	}{
		{"proc", "proc", "proc", 0, ""},
		{"sysfs", "sys", "sysfs", 0, ""},
		{"tmpfs", "tmp", "tmpfs", 0, ""},
		{"tmpfs", "dev", "tmpfs", syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755"},
	}

	for _, m := range mounts {
		target := filepath.Join(rootfs, m.target)
		os.MkdirAll(target, 0755)
		if err := syscall.Mount(m.source, target, m.fstype, m.flags, m.data); err != nil {
			fmt.Fprintf(os.Stderr, "warning: mount %s: %v\n", m.target, err)
		}
	}

	// Create essential device nodes
	createDevices(rootfs)

	// Pivot root
	if err := pivotRoot(rootfs); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	// Execute the target command
	if len(command) == 0 {
		command = []string{"/bin/sh"}
	}

	binary, err := exec.LookPath(command[0])
	if err != nil {
		binary = command[0]
	}

	return syscall.Exec(binary, command, os.Environ())
}

func pivotRoot(newRoot string) error {
	putOld := filepath.Join(newRoot, ".pivot_old")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return err
	}

	// Bind mount newRoot to itself (required for pivot_root)
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount: %w", err)
	}

	if err := unix.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	// Unmount the old root
	if err := syscall.Unmount("/.pivot_old", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old root: %w", err)
	}

	return os.RemoveAll("/.pivot_old")
}

func createDevices(rootfs string) {
	devPath := filepath.Join(rootfs, "dev")

	// Create /dev/null, /dev/zero, /dev/random, /dev/urandom
	devices := []struct {
		name  string
		major uint32
		minor uint32
		mode  uint32
	}{
		{"null", 1, 3, 0666},
		{"zero", 1, 5, 0666},
		{"random", 1, 8, 0666},
		{"urandom", 1, 9, 0666},
		{"tty", 5, 0, 0666},
	}

	for _, d := range devices {
		path := filepath.Join(devPath, d.name)
		dev := unix.Mkdev(d.major, d.minor)
		unix.Mknod(path, syscall.S_IFCHR|d.mode, int(dev))
	}

	// Symlinks
	os.Symlink("/proc/self/fd", filepath.Join(devPath, "fd"))
	os.Symlink("/proc/self/fd/0", filepath.Join(devPath, "stdin"))
	os.Symlink("/proc/self/fd/1", filepath.Join(devPath, "stdout"))
	os.Symlink("/proc/self/fd/2", filepath.Join(devPath, "stderr"))
}

// Cgroups v2 management

const cgroupBase = "/sys/fs/cgroup"

func setupCgroups(c *Container, pid int) error {
	cgroupPath := filepath.Join(cgroupBase, "sheep", c.ID)
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return err
	}

	// Add process to cgroup
	if err := writeFile(filepath.Join(cgroupPath, "cgroup.procs"), strconv.Itoa(pid)); err != nil {
		return err
	}

	// Enable controllers
	controllers := "+memory +pids +cpu"
	parentCtrl := filepath.Join(cgroupBase, "sheep", "cgroup.subtree_control")
	writeFile(parentCtrl, controllers)

	// Memory limit
	if c.Config.Memory > 0 {
		writeFile(filepath.Join(cgroupPath, "memory.max"), strconv.FormatInt(c.Config.Memory, 10))
	}

	// PID limit
	if c.Config.PidsLimit > 0 {
		writeFile(filepath.Join(cgroupPath, "pids.max"), strconv.FormatInt(c.Config.PidsLimit, 10))
	}

	// CPU quota
	if c.Config.CPUQuota > 0 {
		// cpu.max: $QUOTA $PERIOD (microseconds)
		quota := fmt.Sprintf("%d 100000", c.Config.CPUQuota)
		writeFile(filepath.Join(cgroupPath, "cpu.max"), quota)
	}

	// CPU weight (shares)
	if c.Config.CPUShares > 0 {
		// Convert Docker-style shares (2-262144) to cgroup v2 weight (1-10000)
		weight := (c.Config.CPUShares * 10000) / 262144
		if weight < 1 {
			weight = 1
		}
		writeFile(filepath.Join(cgroupPath, "cpu.weight"), strconv.FormatInt(weight, 10))
	}

	return nil
}

func cleanupCgroups(c *Container) {
	cgroupPath := filepath.Join(cgroupBase, "sheep", c.ID)
	os.RemoveAll(cgroupPath)
}

func mountOverlay(lower, upper, work, merged string) error {
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
	return syscall.Mount("overlay", merged, "overlay", 0, opts)
}

func unmountOverlay(merged string) {
	syscall.Unmount(merged, syscall.MNT_DETACH)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(strings.TrimSpace(content)), 0644)
}
