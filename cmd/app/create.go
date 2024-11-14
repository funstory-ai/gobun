package app

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/funstory-ai/gobun/adaptors/xiangongyun"
	"github.com/funstory-ai/gobun/internal"
	"github.com/urfave/cli/v2"
)

var CommandCreate = &cli.Command{
	Name:   "create",
	Usage:  "Create a new pod",
	Action: create,
}

func create(ctx *cli.Context) error {
	pool := xiangongyun.NewPool("Bearer " + os.Getenv("XGY_TOKEN"))

	// Create pod with default options
	options := internal.PodOptions{
		GPUModel: internal.GPUModelRTX4090,
		GPUCount: 1,
	}
	pod, err := pool.CreatePod(options)
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
