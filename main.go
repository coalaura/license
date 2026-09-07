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
	var options Options

	command := &cli.Command{
		Name:    "license",
		Usage:   "add a license to a project",
		Version: Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "interactive",
				Aliases:     []string{"i"},
				Usage:       "recommend a license by asking about the project",
				Destination: &options.Interactive,
			},
			&cli.StringFlag{
				Name:        "license",
				Aliases:     []string{"l"},
				Usage:       "set the license",
				Destination: &options.License,
			},
			&cli.StringFlag{
				Name:        "author",
				Aliases:     []string{"a"},
				Usage:       "set the author",
				Destination: &options.Author,
			},
			&cli.StringFlag{
				Name:        "year",
				Aliases:     []string{"y"},
				Usage:       "set the year",
				Destination: &options.Year,
			},
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Usage:       "set the project name",
				Destination: &options.Name,
			},
			&cli.StringFlag{
				Name:        "description",
				Aliases:     []string{"d"},
				Usage:       "set the project description",
				Destination: &options.Description,
			},
		},
		Action: func(_ context.Context, command *cli.Command) error {
			args := command.Args().Slice()
			if len(args) != 0 {
				return fmt.Errorf("unexpected arguments: %q", args)
			}

			if options.Interactive && options.License != "" {
				return fmt.Errorf("--interactive cannot be used with --license")
			}

			return run(options)
		},
	}

	err := command.Run(context.Background(), os.Args)
	if err == nil {
		return
	}

	log.Errorln(err)

	os.Exit(1)
}
