package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"twitter-clone/internal/domain/config"
	"twitter-clone/internal/domain/twitter"

	"github.com/gorilla/mux"
)

type ServerV1 struct {
	tweeterService twitter.TwitterServiceI
	server         *http.Server

	info string
}

func NewServerV1(service twitter.TwitterServiceI, config config.APIConfig) *ServerV1 {
	muxServer := NewMuxServer(config)
	server := &ServerV1{
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
	router.HandleFunc("/api/v1/follow_user", s.followUser).Methods("GET")
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

func (s *ServerV1) followUser(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		user        int64
		followee    int64
		userStr     string
		followeeStr string
	)
	ctx := r.Context()
	if userStr = r.URL.Query().Get("user"); userStr == "" {
		result := map[string]string{
			"error": "User ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	if followeeStr = r.URL.Query().Get("followee"); followeeStr == "" {
		result := map[string]string{
			"error": "Followee ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	if user, err = strconv.ParseInt(r.URL.Query().Get("user"), 10, 64); err != nil {
		result := map[string]string{
			"error": "Invalid user ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	if followee, err = strconv.ParseInt(r.URL.Query().Get("followee"), 10, 64); err != nil {
		result := map[string]string{
			"error": "Invalid user ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if _, err := s.tweeterService.GetUser(ctx, user); err != nil {
		result := map[string]string{
			"error": fmt.Sprintf("User %v does not exist", user),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if _, err := s.tweeterService.GetUser(ctx, followee); err != nil {
		result := map[string]string{
			"error": fmt.Sprintf("User %v does not exist", user),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	followResult := twitter.Follow{
		FollowerID: user,
		FolloweeID: followee,
		CreatedAt:  time.Now().UTC(),
	}
	// maybe it's better to return the follow struct also
	if err = s.tweeterService.FollowUser(ctx, followResult); err != nil {
		result := map[string]string{
			"error": fmt.Sprintf("Failed to follow user: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(followResult)
}
