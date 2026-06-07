package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"sheep/internal/cli"
	"sheep/internal/container"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Handle commands that don't need the container manager
	switch os.Args[1] {
	case "init":
		handleInit()
		return
	case "version":
		fmt.Println("sheep v0.1.0")
		return
	case "help", "--help", "-h":
		printUsage()
		return
	}

	dataDir := os.Getenv("SHEEP_DATA_DIR")
	mgr := container.NewManager(dataDir)
	if err := mgr.Init(); err != nil {
		fatal("init: %v", err)
	}

	switch os.Args[1] {
	case "run":
		cmdRun(mgr)
	case "create":
		cmdCreate(mgr)
	case "start":
		cmdStart(mgr)
	case "stop":
		cmdStop(mgr)
	case "rm":
		cmdRemove(mgr)
	case "ps":
		cmdPs(mgr)
	case "inspect":
		cmdInspect(mgr)
	case "images":
		cmdImages(mgr)
	case "pull":
		cmdPull(mgr)
	case "push":
		cmdPush(mgr)
	case "tag":
		cmdTag(mgr)
	case "import":
		cmdImport(mgr)
	case "bootstrap":
		cmdBootstrap(mgr)
	case "rmi":
		cmdRemoveImage(mgr)
	case "logs":
		cmdLogs(mgr)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleInit() {
	// Parse init arguments: init --rootfs <path> --hostname <name> -- <command...>
	var rootfs, hostname string
	var command []string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--rootfs":
			i++
			if i < len(args) {
				rootfs = args[i]
			}
		case "--hostname":
			i++
			if i < len(args) {
				hostname = args[i]
			}
		case "--":
			command = args[i+1:]
			i = len(args)
		}
	}

	if rootfs == "" {
		fatal("init: --rootfs is required")
	}

	if err := container.ContainerInit(rootfs, hostname, command); err != nil {
		fatal("init: %v", err)
	}
}

func cmdRun(mgr *container.Manager) {
	opts := parseRunFlags(os.Args[2:])

	c, err := mgr.Create(opts)
	if err != nil {
		fatal("create: %v", err)
	}

	if err := mgr.Start(c.ID); err != nil {
		fatal("start: %v", err)
	}

	if opts.Detach {
		fmt.Println(c.ID)
	} else {
		fmt.Printf("container %s started (pid %d)\n", container.ShortID(c.ID), c.Pid)
	}
}

func cmdCreate(mgr *container.Manager) {
	opts := parseRunFlags(os.Args[2:])

	c, err := mgr.Create(opts)
	if err != nil {
		fatal("create: %v", err)
	}

	fmt.Println(c.ID)
}

func cmdStart(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep start <container>")
	}

	c, err := mgr.Get(os.Args[2])
	if err != nil {
		fatal("%v", err)
	}

	if err := mgr.Start(c.ID); err != nil {
		fatal("start: %v", err)
	}

	fmt.Println(container.ShortID(c.ID))
}

func cmdStop(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep stop <container>")
	}

	c, err := mgr.Get(os.Args[2])
	if err != nil {
		fatal("%v", err)
	}

	if err := mgr.Stop(c.ID); err != nil {
		fatal("stop: %v", err)
	}

	fmt.Println(container.ShortID(c.ID))
}

func cmdRemove(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep rm <container>")
	}

	for _, arg := range os.Args[2:] {
		c, err := mgr.Get(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		if err := mgr.Remove(c.ID); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Println(container.ShortID(c.ID))
	}
}

func cmdPs(mgr *container.Manager) {
	all := false
	for _, arg := range os.Args[2:] {
		if arg == "-a" || arg == "--all" {
			all = true
		}
	}

	containers := mgr.List(all)
	tbl := cli.NewTable("CONTAINER ID", "IMAGE", "COMMAND", "CREATED", "STATUS", "NAME")

	for _, c := range containers {
		cmd := strings.Join(c.Command, " ")
		created := timeAgo(c.CreatedAt)
		status := formatStatus(c)
		tbl.AddRow(
			container.ShortID(c.ID),
			c.Image,
			cli.Truncate(cmd, 30),
			created,
			status,
			c.Name,
		)
	}

	tbl.Render(os.Stdout)
}

func cmdInspect(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep inspect <container>")
	}

	c, err := mgr.Get(os.Args[2])
	if err != nil {
		fatal("%v", err)
	}

	fmt.Printf("ID:        %s\n", c.ID)
	fmt.Printf("Name:      %s\n", c.Name)
	fmt.Printf("Image:     %s\n", c.Image)
	fmt.Printf("Command:   %s\n", strings.Join(c.Command, " "))
	fmt.Printf("State:     %s\n", c.State)
	fmt.Printf("PID:       %d\n", c.Pid)
	fmt.Printf("Created:   %s\n", c.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Hostname:  %s\n", c.Config.Hostname)
	fmt.Printf("RootFS:    %s\n", c.RootFS)

	if c.Config.Memory > 0 {
		fmt.Printf("Memory:    %s\n", cli.FormatBytes(c.Config.Memory))
	}
	if c.Config.CPUQuota > 0 {
		fmt.Printf("CPU Quota: %d\n", c.Config.CPUQuota)
	}
	if c.Config.PidsLimit > 0 {
		fmt.Printf("PID Limit: %d\n", c.Config.PidsLimit)
	}

	if c.Network != nil {
		fmt.Printf("IP:        %s\n", c.Network.IPAddress)
		fmt.Printf("Gateway:   %s\n", c.Network.Gateway)
		fmt.Printf("Bridge:    %s\n", c.Network.Bridge)
	}
}

func cmdPull(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep pull <image>[:<tag>]")
	}

	ref := container.ParseImageRef(os.Args[2])
	client := container.NewRegistryClient()

	img, err := client.Pull(ref, mgr.Images(), func(msg string) {
		fmt.Println(msg)
	})
	if err != nil {
		fatal("pull: %v", err)
	}

	fmt.Printf("pulled %s:%s (%s)\n", img.Name, img.Tag, container.ShortID(img.ID))
}

func cmdPush(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep push <registry/repo:tag>")
	}

	ref := container.ParseImageRef(os.Args[2])

	// Find the local image by repo name
	img, err := mgr.Images().Get(ref.Repo, ref.Tag)
	if err != nil {
		// Try without library/ prefix
		name := ref.Repo
		if strings.HasPrefix(name, "library/") {
			name = strings.TrimPrefix(name, "library/")
		}
		img, err = mgr.Images().Get(name, ref.Tag)
		if err != nil {
			fatal("image not found locally: %s:%s", ref.Repo, ref.Tag)
		}
	}

	if err := container.PushImage(img, ref, func(msg string) {
		fmt.Println(msg)
	}); err != nil {
		fatal("push: %v", err)
	}
}

func cmdTag(mgr *container.Manager) {
	if len(os.Args) < 4 {
		fatal("usage: sheep tag <source> <target>")
	}

	srcRef := container.ParseImageRef(os.Args[2])
	dstRef := container.ParseImageRef(os.Args[3])

	srcName := srcRef.Repo
	if strings.HasPrefix(srcName, "library/") {
		srcName = strings.TrimPrefix(srcName, "library/")
	}

	img, err := mgr.Images().Get(srcName, srcRef.Tag)
	if err != nil {
		fatal("source image not found: %v", err)
	}

	dstName := dstRef.Repo
	if strings.HasPrefix(dstName, "library/") {
		dstName = strings.TrimPrefix(dstName, "library/")
	}

	tagged, err := mgr.Images().Tag(img.ID, dstName, dstRef.Tag)
	if err != nil {
		fatal("tag: %v", err)
	}

	fmt.Printf("tagged %s:%s as %s:%s\n", srcName, srcRef.Tag, tagged.Name, tagged.Tag)
}

func cmdImages(mgr *container.Manager) {
	images, err := mgr.Images().List()
	if err != nil {
		fatal("images: %v", err)
	}

	tbl := cli.NewTable("IMAGE ID", "NAME", "TAG", "SIZE", "CREATED")
	for _, img := range images {
		tbl.AddRow(
			container.ShortID(img.ID),
			img.Name,
			img.Tag,
			cli.FormatBytes(img.Size),
			timeAgo(img.CreatedAt),
		)
	}
	tbl.Render(os.Stdout)
}

func cmdImport(mgr *container.Manager) {
	if len(os.Args) < 4 {
		fatal("usage: sheep import <name> <tarball>")
	}

	name := os.Args[2]
	tarPath := os.Args[3]

	img, err := mgr.Images().Import(name, "latest", tarPath)
	if err != nil {
		fatal("import: %v", err)
	}

	fmt.Printf("imported %s:%s (%s)\n", img.Name, img.Tag, container.ShortID(img.ID))
}

func cmdBootstrap(mgr *container.Manager) {
	name := "minimal"
	if len(os.Args) > 2 {
		name = os.Args[2]
	}

	img, err := mgr.Images().Bootstrap(name)
	if err != nil {
		fatal("bootstrap: %v", err)
	}

	fmt.Printf("bootstrapped %s:%s (%s)\n", img.Name, img.Tag, container.ShortID(img.ID))
}

func cmdRemoveImage(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep rmi <image>")
	}

	img, err := mgr.Images().Get(os.Args[2], "latest")
	if err != nil {
		fatal("%v", err)
	}

	if err := mgr.Images().Remove(img.ID); err != nil {
		fatal("remove image: %v", err)
	}

	fmt.Println(container.ShortID(img.ID))
}

func cmdLogs(mgr *container.Manager) {
	if len(os.Args) < 3 {
		fatal("usage: sheep logs <container>")
	}

	c, err := mgr.Get(os.Args[2])
	if err != nil {
		fatal("%v", err)
	}

	// Read log file if it exists
	base := os.Getenv("SHEEP_DATA_DIR")
	if base == "" {
		base = "/var/lib/sheep"
	}
	logPath := fmt.Sprintf("%s/containers/%s/output.log", base, c.ID)
	data, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no logs available for container %s\n", container.ShortID(c.ID))
		return
	}
	os.Stdout.Write(data)
}

func parseRunFlags(args []string) container.RunOpts {
	opts := container.RunOpts{}
	var i int

	for i = 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			i++
			if i < len(args) {
				opts.Name = args[i]
			}
		case "-d", "--detach":
			opts.Detach = true
		case "-m", "--memory":
			i++
			if i < len(args) {
				opts.Config.Memory = parseMemory(args[i])
			}
		case "--cpu-shares":
			i++
			if i < len(args) {
				v, _ := strconv.ParseInt(args[i], 10, 64)
				opts.Config.CPUShares = v
			}
		case "--cpu-quota":
			i++
			if i < len(args) {
				v, _ := strconv.ParseInt(args[i], 10, 64)
				opts.Config.CPUQuota = v
			}
		case "--pids-limit":
			i++
			if i < len(args) {
				v, _ := strconv.ParseInt(args[i], 10, 64)
				opts.Config.PidsLimit = v
			}
		case "--hostname", "-h":
			i++
			if i < len(args) {
				opts.Config.Hostname = args[i]
			}
		case "-e", "--env":
			i++
			if i < len(args) {
				opts.Config.Env = append(opts.Config.Env, args[i])
			}
		case "-v", "--volume":
			i++
			if i < len(args) {
				parts := strings.SplitN(args[i], ":", 3)
				if len(parts) >= 2 {
					m := container.Mount{Source: parts[0], Target: parts[1]}
					if len(parts) == 3 && parts[2] == "ro" {
						m.ReadOnly = true
					}
					opts.Mounts = append(opts.Mounts, m)
				}
			}
		case "-w", "--workdir":
			i++
			if i < len(args) {
				opts.Config.WorkDir = args[i]
			}
		default:
			// First non-flag is the image, rest is command
			if opts.Image == "" {
				opts.Image = args[i]
			} else {
				opts.Command = append(opts.Command, args[i])
			}
		}
	}

	if opts.Image == "" {
		fatal("image is required")
	}
	if len(opts.Command) == 0 {
		opts.Command = []string{"/bin/sh"}
	}

	return opts
}

func parseMemory(s string) int64 {
	s = strings.TrimSpace(s)
	multiplier := int64(1)

	if strings.HasSuffix(s, "g") || strings.HasSuffix(s, "G") {
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "m") || strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "k") || strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = s[:len(s)-1]
	}

	v, _ := strconv.ParseInt(s, 10, 64)
	return v * multiplier
}

func formatStatus(c *container.Container) string {
	switch c.State {
	case container.StateRunning:
		return fmt.Sprintf("Up %s", timeAgo(c.StartedAt))
	case container.StateStopped:
		return fmt.Sprintf("Exited (%d) %s", c.ExitCode, timeAgo(c.StoppedAt))
	default:
		return string(c.State)
	}
}

func timeAgo(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "sheep: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`sheep - container runtime

Usage: sheep <command> [options]

Container Commands:
  run         Create and start a container
  create      Create a container
  start       Start a stopped container
  stop        Stop a running container
  rm          Remove a container
  ps          List containers (-a for all)
  inspect     Show container details
  logs        Show container logs

Image Commands:
  pull        Pull an image from a registry (Docker Hub, Meadow, etc.)
  push        Push an image to a registry
  tag         Tag an image with a new name
  images      List images
  import      Import a rootfs tarball as an image
  bootstrap   Create a minimal image from host
  rmi         Remove an image

Run Options:
  --name       Container name
  -d           Run in background (detach)
  -m           Memory limit (e.g., 256m, 1g)
  --cpu-shares CPU shares (relative weight)
  --cpu-quota  CPU quota (microseconds)
  --pids-limit Maximum number of PIDs
  --hostname   Container hostname
  -e           Environment variable (KEY=VALUE)
  -v           Volume mount (host:container[:ro])
  -w           Working directory

Examples:
  sheep pull nginx
  sheep pull alpine:3.19
  sheep run --name web -m 256m nginx nginx
  sheep bootstrap minimal
  sheep run --name box minimal /bin/sh
  sheep ps -a
  sheep stop web
  sheep rm web`)
}
