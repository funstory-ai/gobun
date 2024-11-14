package app

import "github.com/urfave/cli/v2"

var CommandAttach = &cli.Command{
	Name:   "attach",
	Usage:  "Attach to a running pod",
	Action: attach,
}

func attach(ctx *cli.Context) error {
	return nil
}
