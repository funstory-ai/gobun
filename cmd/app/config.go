package app

import "github.com/urfave/cli/v2"

var CommandConfig = &cli.Command{
	Name:   "config",
	Usage:  "Configure the CLI",
	Action: config,
}

func config(ctx *cli.Context) error {
	return nil
}
