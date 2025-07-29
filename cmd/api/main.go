package twitter_clone_api

import (
	"context"
	"fmt"
	"twitter-clone/internal/config"
	"twitter-clone/internal/domain/database"
	"twitter-clone/internal/domain/twitter"
	"twitter-clone/internal/server"

	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

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
		Action: runServer,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func runServer(cCtx *cli.Context) error {
	var (
		err            error
		configYaml     *config.YamlConfig
		database       database.DatabaseI
		tweeterService twitter.TwitterService
	)

	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configYaml, err = config.NewYamlConfig(cCtx.String("config"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	server := server.NewServerV1(tweeterService, database, configYaml) // MOCKED

	go func() {
		log.Info().Msgf("Starting data server: %s \n", server.Info())
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Common server failed: %v", err)
		}

	}()

	<-signalCtx.Done()
	log.Info().Msg("Shut down data server")
	if err = server.Stop(context.TODO()); err != nil {
		log.Fatal().Msg("Can't terminate data server")
	}

	// select {
	// case
	// }
	return nil
}
