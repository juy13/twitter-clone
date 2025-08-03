package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"twitter-clone/internal/config"
	"twitter-clone/internal/server/metrics"
	"twitter-clone/internal/server/worker"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	redis_cache "twitter-clone/internal/cache"
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
	)

	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configYaml, err = config.NewYamlConfig(cCtx.String("config"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cache := redis_cache.NewRedisCache(configYaml)

	worker := worker.NewWorker(cache)
	debugServer := metrics.NewMetricsServer(configYaml)

	go func() {
		_ = worker.Start(signalCtx) // lint issue
	}()

	go func() {
		log.Info().Msgf("Starting data server: %s \n", debugServer.Info())
		if err := debugServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Common server failed: %v", err)
		}

	}()

	<-signalCtx.Done()
	log.Info().Msg("Shut down the worker")

	return nil
}
