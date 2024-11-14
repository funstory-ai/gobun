package app

import (
	"fmt"
	"os"

	"github.com/funstory-ai/gobun/adaptors/xiangongyun"
	"github.com/urfave/cli/v2"
)

var CommandDestroy = &cli.Command{
	Name:      "destroy",
	Usage:     "Destroy a pod",
	ArgsUsage: "<pod-id>",
	Action:    destroy,
}

func destroy(ctx *cli.Context) error {
	pool := xiangongyun.NewPool("Bearer " + os.Getenv("XGY_TOKEN"))
	if ctx.NArg() < 1 {
		return fmt.Errorf("pod ID is required")
	}
	podID := ctx.Args().First()

	return pool.DestroyPod(podID)
}
