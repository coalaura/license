package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coalaura/plain"
	"github.com/urfave/cli/v3"
)

var Version = "dev"

var log = plain.New()

func main() {
	command := &cli.Command{
		Name:    "license",
		Usage:   "add a license to a project",
		Version: Version,
		Action: func(_ context.Context, command *cli.Command) error {
			args := command.Args().Slice()
			if len(args) != 0 {
				return fmt.Errorf("unexpected arguments: %q", args)
			}

			return run()
		},
	}

	err := command.Run(context.Background(), os.Args)
	if err == nil {
		return
	}

	log.Errorln(err)

	os.Exit(1)
}
