package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"twitter-clone/internal/config"
	"twitter-clone/internal/worker"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	redis_cache "twitter-clone/internal/cache"
	postgres_db "twitter-clone/internal/database/postgres"
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
		Name:            "Twitter Worker",
		Version:         GitTag,
		HideHelpCommand: true,
		HideVersion:     false,
		Description:     "Simulates the Twitter worker (not Elon)",
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
	var (
		err        error
		configYaml *config.YamlConfig
		database   *postgres_db.PostgresDB
	)

	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configYaml, err = config.NewYamlConfig(cCtx.String("config"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if database, err = postgres_db.NewPostgresDB(configYaml); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	cache := redis_cache.NewRedisCache(configYaml)

	worker := worker.NewWorker(database, cache)

	go worker.Start(signalCtx)

	<-signalCtx.Done()
	log.Info().Msg("Shut down the worker")

	// select {
	// case
	// }
	return nil
}
