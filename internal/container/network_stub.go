//go:build !linux

package container

func setupNetworkForContainer(c *Container, pid int) error {
	return nil // Networking not available on non-Linux
}

func LoadIPCounter(baseDir string) {}
func SaveIPCounter(baseDir string) {}
