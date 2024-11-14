package app

import "github.com/urfave/cli/v2"

var CommandCreate = &cli.Command{
	Name:   "create",
	Usage:  "Create a new pod",
	Action: create,
}

func create(ctx *cli.Context) error {
	return nil
}
