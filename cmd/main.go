package main

import (
	"os"

	"github.com/funstory-ai/gobun/cmd/app"
)

func run(args []string) error {
	app := app.New()
	return app.Run(args)
}

func handleErr(err error) {
	if err == nil {
		return
	}
	os.Exit(1)
}

func main() {
	err := run(os.Args)
	handleErr(err)
}
