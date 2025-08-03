package server

import (
	"context"
	"encoding/json"
	"errors"
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
	router         *mux.Router

	info string
}

func NewServerV1(service twitter.TwitterServiceI, config config.APIConfig) *ServerV1 {
	muxServer, router := NewMuxServer(config)
	server := &ServerV1{
		tweeterService: service,
		server:         muxServer,
		info:           fmt.Sprintf("Running server on %v", config.Host()+":"+strconv.Itoa(config.Port())),
		router:         router,
	}
	server.registerRoutes()
	return server
}

func (s *ServerV1) Info() string {
	return s.info
}

func (s *ServerV1) registerRoutes() {
	router := s.router

	// Tweets
	router.HandleFunc("/api/v1/tweet", s.newTweet).Methods("POST")
	router.HandleFunc("/api/v1/tweets", s.returnTweets).Methods("GET")
	router.HandleFunc("/api/v1/get_tweet", s.getTweet).Methods("GET")
	router.HandleFunc("/api/v1/tweet_by_user", s.getTweetByUser).Methods("GET") // not implemented yet

	// Follow
	// Maybe it's better to unite them,
	// until we are using same code and use params for logic????
	router.HandleFunc("/api/v1/follow_user", s.followUser).Methods("GET")
	router.HandleFunc("/api/v1/followings", s.getFollowings).Methods("GET")
	router.HandleFunc("/api/v1/followers", s.getFollowers).Methods("GET")
	// Add more routes
	router.HandleFunc("/api/v1/get_user", s.getUser).Methods("GET")

	// Add user
	router.HandleFunc("/api/v1/new_user", s.newUser).Methods("POST")
}

func (s *ServerV1) Start() error {
	return s.server.ListenAndServe()
}

func (s *ServerV1) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *ServerV1) extractAndCheckUser(ctx context.Context, r *http.Request, userField string) (int64, error) {
	var (
		userStr string
		user    int64
		err     error
	)
	if userStr = r.URL.Query().Get(userField); userStr == "" {
		return 0, errors.New("user ID is required")
	}
	if user, err = strconv.ParseInt(userStr, 10, 64); err != nil {
		return 0, errors.New("invalid user ID")
	}
	if _, err := s.tweeterService.GetUser(ctx, user); err != nil {
		return 0, fmt.Errorf("user %v does not exist", user)
	}
	return user, nil
}

func (s *ServerV1) newUser(w http.ResponseWriter, r *http.Request) {
	var err error
	var userID int64
	var user twitter.User
	ctx := r.Context()

	if r.Method != http.MethodPost {
		result := map[string]string{
			"error": "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		result := map[string]string{
			"error": "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	// TODO check if user exists

	if userID, err = s.tweeterService.CreateUser(ctx, user); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(twitter.User{
		ID:        userID,
		Username:  user.Username,
		CreatedAt: time.Now(),
	})
}

func (s *ServerV1) getUser(w http.ResponseWriter, r *http.Request) {
	var err error
	var userID int64
	var user twitter.User
	ctx := r.Context()
	if userID, err = s.extractAndCheckUser(ctx, r, "user"); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if user, err = s.tweeterService.GetUser(ctx, userID); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

/////////////////////////////////////////////////////
// 				TWEET PART
/////////////////////////////////////////////////////

func (s *ServerV1) newTweet(w http.ResponseWriter, r *http.Request) {
	var err error
	var user int64
	ctx := r.Context()

	if user, err = s.extractAndCheckUser(ctx, r, "user"); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // even if incorrect user return not found
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if r.Method != http.MethodPost {
		result := map[string]string{
			"error": "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	type tweetRequest struct {
		Content string `json:"content"`
	}
	var tweet tweetRequest
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		result := map[string]string{
			"error": "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	// have to return the creation time
	if err := s.tweeterService.NewTweet(ctx, twitter.Tweet{
		UserID:    user,
		Content:   tweet.Content,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		result := map[string]string{
			"error": "Failed to create tweet",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]string{
		"message": "Tweet created successfully",
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (s *ServerV1) returnTweets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	var user int64
	var tweets []twitter.Tweet
	if user, err = s.extractAndCheckUser(ctx, r, "user"); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // even if incorrect user return not found
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if tweets, err = s.tweeterService.GetTimeline(ctx, user); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // even if incorrect user return not found
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tweets)
}

func (s *ServerV1) getTweetByUser(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (s *ServerV1) getTweet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var tweetStr string
	var tweetID int64
	var err error
	var tweet twitter.Tweet
	if tweetStr = r.URL.Query().Get("tweet"); tweetStr == "" {
		return // TODO return error
	}
	if tweetID, err = strconv.ParseInt(tweetStr, 10, 64); err != nil {
		return // TODO return error
	}

	if tweet, err = s.tweeterService.GetTweet(ctx, tweetID); err != nil {
		return // TODO return error
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tweet)

}

/////////////////////////////////////////////////////
// 				FOLLOWING PART
/////////////////////////////////////////////////////

func (s *ServerV1) getFollowings(w http.ResponseWriter, r *http.Request) {
	var err error
	var user int64
	var users []twitter.User
	ctx := r.Context()

	if user, err = s.extractAndCheckUser(ctx, r, "user"); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // even if incorrect user return not found
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if users, err = s.tweeterService.Following(ctx, user); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		// for better distinguish between not found and internal error -- better errors
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)

}

func (s *ServerV1) getFollowers(w http.ResponseWriter, r *http.Request) {
	var err error
	var user int64
	var users []twitter.User
	ctx := r.Context()

	if user, err = s.extractAndCheckUser(ctx, r, "user"); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // even if incorrect user return not found
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if users, err = s.tweeterService.Followers(ctx, user); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		// for better distinguish between not found and internal error -- better errors
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}

func (s *ServerV1) followUser(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		user     int64
		followee int64
	)
	ctx := r.Context()

	if user, err = s.extractAndCheckUser(ctx, r, "user"); err != nil {
		result := map[string]string{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if followee, err = s.extractAndCheckUser(ctx, r, "followee"); err != nil {
		result := map[string]string{
			"error": err.Error(),
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
