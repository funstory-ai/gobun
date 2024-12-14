package app

import (
	"fmt"
	"strconv"

	gobun_config "github.com/funstory-ai/gobun/internal/config"
	"github.com/funstory-ai/gobun/internal/resource"
	"github.com/funstory-ai/gobun/internal/ssh"
	"github.com/funstory-ai/gobun/vendors"
	"github.com/urfave/cli/v2"
)

var CommandAttach = &cli.Command{
	Name:      "attach",
	Usage:     "Attach to a running pod",
	ArgsUsage: "POD_ID",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "vendor",
			Usage:    "Specify the vendor",
			Required: true,
		},
	},
	Action: attach,
}

func attach(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return cli.Exit("Pod ID is required", 1)
	}

	// Get vendor configuration
	vendorId := ctx.String("vendor")
	vendorConfigs, err := gobun_config.NewVendorConfigs()
	if err != nil {
		return fmt.Errorf("failed to create vendor configs: %w", err)
	}
	vendorConfig, err := vendorConfigs.GetVendorConfig(vendorId)
	if err != nil {
		return fmt.Errorf("failed to get vendor config: %w", err)
	}

	// Create vendor instance
	options := vendors.VendorOptions{
		APISecret: vendorConfig.APISecret,
	}
	vendor, err := vendors.NewVendor(ctx.Context, vendors.CloudProviderTypeXianGongYun, options)
	if err != nil {
		return fmt.Errorf("failed to create vendor: %w", err)
	}

	// Get pod information
	podID := ctx.Args().First()
	pod, err := vendor.GetPod(ctx.Context, resource.Pod{ID: podID})
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
