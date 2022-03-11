package commands

import (
	"fmt"
	"log"

	"github.com/Gympass/aws-vault-scg/pkg/aws"
	"github.com/urfave/cli/v2"
)

var Generate = &cli.Command{
	Name:  "generate",
	Usage: "Generate AWS config",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "ssoURL",
			Aliases:  []string{"s"},
			Usage:    "AWS Single Sign On URL",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "region",
			Aliases: []string{"r"},
			Usage:   "AWS Region used by SSO",
		},
	},
	Action: func(c *cli.Context) error {
		ssoURL := c.String("ssoURL")
		region := c.String("region")

		if c.String("region") == "" {
			region = "us-east-1"
		}

		var o string
		fmt.Print("Overwrite current config(~/.aws/config)[y/N]? ")
		_, err := fmt.Scanf("%s", &o)
		if err != nil {
			fmt.Println("Default option")
		}
		switch {
		case o == "y" || o == "Y":
			err := aws.GenerateConfig(ssoURL, region, true)
			if err != nil {
				log.Fatalf("Error to generate config file: %v", err)
			}
		case o == "n" || o == "N":
			err := aws.GenerateConfig(ssoURL, region, false)
			fmt.Println("You can use this profile values to update your config file(~/.aws/config)")
			if err != nil {
				log.Fatalf("Error to print config: %v", err)
			}
		default:
			err := aws.GenerateConfig(ssoURL, region, false)
			fmt.Println("You can use this profile values to update your config file(~/.aws/config)")
			if err != nil {
				log.Fatalf("Error to print config: %v", err)
			}
		}
		return nil
	},
}
