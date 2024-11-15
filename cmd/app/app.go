package app

import "github.com/urfave/cli/v2"

type BunApp struct {
	cli.App
}

func New() BunApp {
	internalApp := cli.NewApp()
	internalApp.EnableBashCompletion = true
	internalApp.Name = "GoBun"
	internalApp.Usage = "Managing GPU resources across multiple clouds"
	internalApp.HideHelpCommand = true
	internalApp.HideVersion = true
	internalApp.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in logs",
		},
	}
	internalApp.Commands = []*cli.Command{
		CommandList,
		CommandCreate,
		CommandAttach,
		CommandDestroy,
		CommandUp,
	}
	return BunApp{
		App: *internalApp,
	}
}
