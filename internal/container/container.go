package container

import (
	"crypto/rand"
	"fmt"
	"time"
)

type State string

const (
	StateCreated State = "created"
	StateRunning State = "running"
	StateStopped State = "stopped"
)

type Container struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Image     string           `json:"image"`
	Command   []string         `json:"command"`
	State     State            `json:"state"`
	Pid       int              `json:"pid"`
	ExitCode  int              `json:"exit_code"`
	CreatedAt time.Time        `json:"created_at"`
	StartedAt time.Time        `json:"started_at,omitempty"`
	StoppedAt time.Time        `json:"stopped_at,omitempty"`
	Config    Config           `json:"config"`
	RootFS    string           `json:"rootfs"`
	Mounts    []Mount          `json:"mounts,omitempty"`
	Network   *NetworkSettings `json:"network,omitempty"`
}

type Config struct {
	Hostname string   `json:"hostname"`
	Env      []string `json:"env,omitempty"`
	WorkDir  string   `json:"work_dir"`

	// Resource limits (cgroups)
	Memory    int64 `json:"memory"`     // bytes, 0 = unlimited
	CPUShares int64 `json:"cpu_shares"` // relative weight
	CPUQuota  int64 `json:"cpu_quota"`  // microseconds per period
	PidsLimit int64 `json:"pids_limit"` // 0 = unlimited
}

type Mount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"readonly"`
}

type NetworkSettings struct {
	IPAddress  string `json:"ip_address"`
	Gateway    string `json:"gateway"`
	Bridge     string `json:"bridge"`
	VethHost   string `json:"veth_host"`
	VethGuest  string `json:"veth_guest"`
	MacAddress string `json:"mac_address"`
}

type RunOpts struct {
	Name    string
	Image   string
	Command []string
	Config  Config
	Mounts  []Mount
	Detach  bool
}

func GenerateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func ShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
