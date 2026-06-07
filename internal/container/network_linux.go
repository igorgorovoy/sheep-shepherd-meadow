//go:build linux

package container

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	BridgeName    = "sheep0"
	BridgeSubnet  = "10.20.0.0/16"
	BridgeGateway = "10.20.0.1"
)

var ipCounter uint32 = 1

func setupNetworkForContainer(c *Container, pid int) error {
	// Ensure bridge exists
	if err := ensureBridge(); err != nil {
		return fmt.Errorf("ensure bridge: %w", err)
	}

	// Allocate IP
	ip := allocateIP()

	// Create veth pair
	vethHost := fmt.Sprintf("veth%s", ShortID(c.ID)[:8])
	vethGuest := fmt.Sprintf("eth0")

	// Create veth pair
	if err := run("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethGuest); err != nil {
		return fmt.Errorf("create veth pair: %w", err)
	}

	// Attach host end to bridge
	if err := run("ip", "link", "set", vethHost, "master", BridgeName); err != nil {
		return fmt.Errorf("attach to bridge: %w", err)
	}

	// Move guest end into container network namespace
	if err := run("ip", "link", "set", vethGuest, "netns", strconv.Itoa(pid)); err != nil {
		return fmt.Errorf("move veth to netns: %w", err)
	}

	// Bring up host end
	if err := run("ip", "link", "set", vethHost, "up"); err != nil {
		return fmt.Errorf("bring up host veth: %w", err)
	}

	// Configure guest end inside namespace
	nsRun(pid, "ip", "addr", "add", ip+"/16", "dev", "eth0")
	nsRun(pid, "ip", "link", "set", "eth0", "up")
	nsRun(pid, "ip", "link", "set", "lo", "up")
	nsRun(pid, "ip", "route", "add", "default", "via", BridgeGateway)

	c.Network = &NetworkSettings{
		IPAddress: ip,
		Gateway:   BridgeGateway,
		Bridge:    BridgeName,
		VethHost:  vethHost,
		VethGuest: vethGuest,
	}

	return nil
}

func ensureBridge() error {
	// Check if bridge already exists
	if _, err := net.InterfaceByName(BridgeName); err == nil {
		return nil
	}

	// Create bridge
	if err := run("ip", "link", "add", BridgeName, "type", "bridge"); err != nil {
		return err
	}
	if err := run("ip", "addr", "add", BridgeGateway+"/16", "dev", BridgeName); err != nil {
		return err
	}
	if err := run("ip", "link", "set", BridgeName, "up"); err != nil {
		return err
	}

	// Enable IP forwarding
	os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644)

	// NAT for outbound traffic
	run("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", BridgeSubnet, "-j", "MASQUERADE")

	return nil
}

func allocateIP() string {
	n := atomic.AddUint32(&ipCounter, 1)
	return fmt.Sprintf("10.20.%d.%d", (n>>8)&0xFF, n&0xFF)
}

// LoadIPCounter loads the IP counter from disk to avoid collisions across restarts.
func LoadIPCounter(baseDir string) {
	data, err := os.ReadFile(filepath.Join(baseDir, "network", "ip_counter"))
	if err != nil {
		return
	}
	val := strings.TrimSpace(string(data))
	if n, err := strconv.ParseUint(val, 10, 32); err == nil {
		atomic.StoreUint32(&ipCounter, uint32(n))
	}
}

func SaveIPCounter(baseDir string) {
	os.WriteFile(
		filepath.Join(baseDir, "network", "ip_counter"),
		[]byte(strconv.FormatUint(uint64(atomic.LoadUint32(&ipCounter)), 10)),
		0644,
	)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func nsRun(pid int, name string, args ...string) error {
	nsArgs := append([]string{"-t", strconv.Itoa(pid), "-n", "--", name}, args...)
	return run("nsenter", nsArgs...)
}
