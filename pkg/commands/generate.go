package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

		configDir := configDirName()
		configFile := filepath.Join(configDir, "config")

		if !fileExists(configFile) {
			// no config file, make sure path exists
			createPath(configDir)
			generateConfig(ssoURL, region)
		} else {
			// config already exists, check if the user wants to overwrite
			var selectedOption string
			fmt.Print("Overwrite current config(~/.aws/config)[y/N]? ")
			_, err := fmt.Scanf("%s", &selectedOption)
			if err != nil {
				fmt.Println("Default option")
			}

			if selectedOption == "y" || selectedOption == "Y" {
				generateConfig(ssoURL, region)
			} else {
				printConfig(ssoURL, region)
			}
		}

		return nil
	},
}

func configDirName() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error identifying the home dir: %v", err)
	}
	return filepath.Join(homeDir, ".aws")
}

func fileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true
	} else {
		return false
	}
}

func createPath(path string) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatalf("Error creating %v: %v", path, err)
	}
}

func generateConfig(ssoURL string, region string) {
	err := aws.GenerateConfig(ssoURL, region, true)
	if err != nil {
		log.Fatalf("Error to generate config file: %v", err)
	}
}

func printConfig(ssoURL string, region string) {
	err := aws.GenerateConfig(ssoURL, region, false)
	fmt.Println("You can use this profile values to update your config file(~/.aws/config)")
	if err != nil {
		log.Fatalf("Error to print config: %v", err)
	}
}
