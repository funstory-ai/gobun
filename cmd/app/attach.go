package app

import (
	"fmt"
	"os"
	"strconv"

	"github.com/funstory-ai/gobun/adaptors/xiangongyun"
	"github.com/funstory-ai/gobun/internal/ssh"
	"github.com/urfave/cli/v2"
)

var CommandAttach = &cli.Command{
	Name:      "attach",
	Usage:     "Attach to a running pod",
	ArgsUsage: "POD_ID",
	Action:    attach,
}

func attach(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return cli.Exit("Pod ID is required", 1)
	}
	id := ctx.Args().First()
	pool := xiangongyun.NewPool("Bearer " + os.Getenv("XGY_TOKEN"))
	pod, err := pool.GetPod(id)
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Create SSH client options
	port, err := strconv.Atoi(pod.SSHPort)
	if err != nil {
		return fmt.Errorf("failed to parse SSH port: %w", err)
	}
	opt := ssh.Options{
		Server:   pod.SSHDomain,
		Port:     port,
		User:     pod.SSHUser,
		Password: pod.Password,
		Auth:     true,
	}

	// Create new SSH client
	client, err := ssh.NewClient(opt)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// Attach to the pod
	if err := client.Attach(); err != nil {
		return fmt.Errorf("failed to attach to pod: %w", err)
	}

	return nil
}
