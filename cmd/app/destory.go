package app

import "github.com/urfave/cli/v2"

var CommandDestroy = &cli.Command{
	Name:   "destroy",
	Usage:  "Destroy a pod",
	Action: destroy,
}

func destroy(ctx *cli.Context) error {
	return nil
}
