package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// The idea is that worker will check the Redis global queue and on new item
// it will send to the user's local queue new tweet

var (
	GitCommit string
	GitTag    string
	BuildTime string
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("Git Tag: %s\n", GitTag)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
	}
	app := &cli.App{
		Name:            "Twitter Clone API",
		Version:         GitTag,
		HideHelpCommand: true,
		HideVersion:     false,
		Description:     "Simulates Twitter",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
			},
		},
		Action: runWorker,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func runWorker(cCtx *cli.Context) error {
	panic("not implemented yet")
}
