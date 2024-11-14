package app

import "github.com/urfave/cli/v2"

var CommandUp = &cli.Command{
	Name:   "up",
	Usage:  "Quickly start a pod and attach to it",
	Action: up,
}

func up(ctx *cli.Context) error {
	return nil
}
