package main

import (
	"github.com/Gympass/aws-vault-scg/pkg/commands"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "acg",
		Usage: "CLI to generate the aws config file with your permissions",
		Commands: []*cli.Command{
			commands.Generate,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
