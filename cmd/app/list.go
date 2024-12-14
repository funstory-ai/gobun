package app

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	gobun_config "github.com/funstory-ai/gobun/internal/config"
	"github.com/funstory-ai/gobun/vendors"
	"github.com/urfave/cli/v2"
)

var CommandList = &cli.Command{
	Name:  "list",
	Usage: "List all pods",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "watch",
			Aliases: []string{"w"},
			Usage:   "Watch pods status, refresh every 5 seconds",
		},
		&cli.StringFlag{
			Name:     "vendor",
			Usage:    "Specify the vendor",
			Required: true,
		},
	},
	Action: list,
}

func list(ctx *cli.Context) error {
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

	// Function to display pods
	displayPods := func() error {
		pods, err := vendor.ListPods(ctx.Context)
		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}

		if ctx.Bool("watch") {
			fmt.Print("\033[H\033[2J")
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "ID\tPOOL ID\tNAME\tSTATUS\tGPU\tGPU MODEL\tMEMORY")

		for _, pod := range pods {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				pod.ID,
				pod.PoolID,
				pod.Name,
				pod.Status,
				pod.GPUCount,
				pod.GPUModel,
				humanReadableMemory(pod.MemorySize),
			)
		}

		return w.Flush()
	}

	if ctx.Bool("watch") {
		for {
			if err := displayPods(); err != nil {
				return err
			}
			time.Sleep(5 * time.Second)
		}
	}

	return displayPods()
}

// humanReadableMemory converts bytes to a human-readable format
func humanReadableMemory(bytes int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
