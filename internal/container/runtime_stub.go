//go:build !linux

package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// startContainer runs the process directly on the host (no isolation).
// This is a development/demo mode for non-Linux systems.
func startContainer(c *Container) (int, error) {
	if len(c.Command) == 0 {
		return 0, fmt.Errorf("no command specified")
	}

	// Resolve binary inside the container rootfs first, then host PATH
	binary := c.Command[0]
	if c.RootFS != "" {
		candidate := c.RootFS + binary
		if _, err := os.Stat(candidate); err == nil {
			binary = candidate
		}
	}

	cmd := exec.Command(binary, c.Command[1:]...)
	cmd.Dir = c.RootFS
	cmd.Env = append(os.Environ(), c.Config.Env...)
	cmd.Env = append(cmd.Env, "SHEEP_CONTAINER_ID="+c.ID, "SHEEP_HOST_MODE=1")

	// Log output to file
	logDir := fmt.Sprintf("/var/lib/sheep/containers/%s", c.ID)
	if d := os.Getenv("SHEEP_DATA_DIR"); d != "" {
		logDir = fmt.Sprintf("%s/containers/%s", d, c.ID)
	}
	os.MkdirAll(logDir, 0755)

	logFile, err := os.Create(logDir + "/output.log")
	if err != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return 0, fmt.Errorf("start process: %w", err)
	}

	// Wait in background, close log file when done
	go func() {
		cmd.Wait()
		if logFile != nil {
			logFile.Close()
		}
	}()

	return cmd.Process.Pid, nil
}

func stopContainer(c *Container) (int, error) {
	if c.Pid <= 0 {
		return 0, nil
	}

	proc, err := os.FindProcess(c.Pid)
	if err != nil {
		return 0, nil
	}

	// Graceful then force
	proc.Signal(syscall.SIGTERM)
	proc.Signal(syscall.SIGKILL)
	state, _ := proc.Wait()

	if state != nil {
		return state.ExitCode(), nil
	}
	return 0, nil
}

func mountOverlay(lower, upper, work, merged string) error {
	return fmt.Errorf("overlayfs not available, using copy fallback")
}

func unmountOverlay(merged string) {}

// ContainerInit — not used in host mode.
func ContainerInit(rootfs, hostname string, command []string) error {
	return fmt.Errorf("container init not needed in host mode")
}
