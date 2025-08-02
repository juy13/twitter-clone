package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"twitter-clone/internal/app/api"
	"twitter-clone/internal/config"
	wsserver "twitter-clone/internal/server/ws_server"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	redis_cache "twitter-clone/internal/cache"
)

// Idea is that there is a web socket server that does:
// 1. accepts connects from users
// 2. stores active users in redis
// 3. if any following pushes tweet -- send to this user

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
		Name:            "Twitter WebSocket Server",
		Version:         GitTag,
		HideHelpCommand: true,
		HideVersion:     false,
		Description:     "Simulates(actually works) a web socket connection for the real time updates",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
			},
		},
		Action: runWebSocketServer,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func runWebSocketServer(cCtx *cli.Context) error {
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
	apiService := api.NewAPIService(configYaml.WSServerAPIPath())
	websocketServer := wsserver.NewWebSocketServer(cache, configYaml, apiService)

	go func() {
		log.Info().Msgf("Starting web socket server: %s \n", websocketServer.Info())
		if err := websocketServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Common server failed: %v", err)
		}

	}()

	<-signalCtx.Done()
	log.Info().Msg("Shut down web socket server")
	if err = websocketServer.Stop(context.TODO()); err != nil {
		log.Fatal().Msg("Can't terminate web socket server")
	}

	return nil
}
