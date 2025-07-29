package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"twitter-clone/internal/domain/config"
	"twitter-clone/internal/domain/database"
	"twitter-clone/internal/domain/twitter"

	"github.com/gorilla/mux"
)

type ServerV1 struct {
	db             database.DatabaseI
	tweeterService twitter.TwitterServiceI
	server         *http.Server

	info string
}

func NewServerV1(service twitter.TwitterServiceI, database database.DatabaseI, config config.APIConfig) *ServerV1 {
	muxServer := NewMuxServer(config)
	server := &ServerV1{
		db:             database,
		tweeterService: service,
		server:         muxServer,
		info:           fmt.Sprintf("Running server on %v", config.Host()+":"+strconv.Itoa(config.Port())),
	}
	server.registerRoutes()
	return server
}

func (s *ServerV1) Info() string {
	return s.info
}

func (s *ServerV1) registerRoutes() {
	router := s.server.Handler.(*mux.Router)
	router.HandleFunc("/api/v1/tweets", s.returnTweets).Methods("GET")
	router.HandleFunc("/api/v1/tweet", s.newTweet).Methods("POST")
	router.HandleFunc("/api/v1/tweet_by_user", s.getTweet).Methods("GET")
	// Add more routes
}

func (s *ServerV1) Start() error {
	return s.server.ListenAndServe()
}

func (s *ServerV1) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *ServerV1) returnTweets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.tweeterService.GetTweet(ctx, 0)
}

func (s *ServerV1) newTweet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.tweeterService.NewTweet(ctx, twitter.Tweet{})
}

func (s *ServerV1) getTweet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.tweeterService.GetUsersTweets(ctx, 0)
}
