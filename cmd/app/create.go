package app

import (
	"fmt"
	"os"
	"text/tabwriter"

	gobun_config "github.com/funstory-ai/gobun/internal/config"
	"github.com/funstory-ai/gobun/internal/resource"
	"github.com/funstory-ai/gobun/vendors"
	"github.com/urfave/cli/v2"
)

var CommandCreate = &cli.Command{
	Name:  "create",
	Usage: "Create a new pod",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "vendor",
			Usage:    "Specify the vendor",
			Required: true,
		},
	},
	Action: create,
}

func create(ctx *cli.Context) error {
	vendorId := ctx.String("vendor")
	vendorConfigs, err := gobun_config.NewVendorConfigs()
	if err != nil {
		return fmt.Errorf("failed to create vendor configs: %w", err)
	}
	vendorConfig, err := vendorConfigs.GetVendorConfig(vendorId)
	if err != nil {
		return fmt.Errorf("failed to get vendor config: %w", err)
	}
	options := vendors.VendorOptions{
		APISecret: vendorConfig.APISecret,
	}
	vendor, err := vendors.NewVendor(ctx.Context, vendors.CloudProviderTypeXianGongYun, options)
	if err != nil {
		return fmt.Errorf("failed to create vendor: %w", err)
	}
	// Create pod with default options
	podOptions := resource.PodOptions{
		GPUModel: resource.GPUModelRTX4090,
		GPUCount: 1,
	}
	pod, err := vendor.CreatePod(ctx.Context, resource.PodOptions{
		GPUModel: podOptions.GPUModel,
		GPUCount: podOptions.GPUCount,
	})
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "ID\tPOOL ID\tNAME\tSTATUS\tGPU\tGPU MODEL\tMEMORY")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
		pod.ID,
		pod.PoolID,
		pod.Name,
		pod.Status,
		pod.GPUCount,
		pod.GPUModel,
		humanReadableMemory(pod.MemorySize),
	)
	w.Flush()

	return nil
}
