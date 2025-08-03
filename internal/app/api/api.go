package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"twitter-clone/internal/domain/twitter"
)

type APIService struct {
	path string
}

func NewAPIService(path string) *APIService {
	return &APIService{
		path: path,
	}
}

func (a *APIService) GetUser(ctx context.Context, userID int64) (twitter.User, error) {
	var err error
	var user twitter.User
	userPath := fmt.Sprintf("%s/api/v1/get_user?user=%d", a.path, userID)
	if err = a.request(ctx, userPath, "GET", &user); err != nil {
		return user, fmt.Errorf("error making request: %v", err)
	}

	if user.ID == 0 {
		return user, fmt.Errorf("invalid user ID")
	}
	return user, nil
}

func (a *APIService) GetFollowers(ctx context.Context, userID int64) ([]twitter.User, error) {
	var (
		err       error
		followers []twitter.User
	)
	followersPath := fmt.Sprintf("%s/api/v1/followers?user=%d", a.path, userID)
	if err = a.request(ctx, followersPath, "GET", &followers); err != nil {
		return followers, fmt.Errorf("error making request: %v", err)
	}
	// this should be just a rc or smth
	// if len(followers) == 0 {
	// 	return nil, fmt.Errorf("no followers found for user ID %d", userID)
	// }
	return followers, nil
}

func (a *APIService) GetTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	var (
		err    error
		tweets []twitter.Tweet
	)
	tweetsPath := fmt.Sprintf("%s/api/v1/tweets?user=%d", a.path, userID)
	if err = a.request(ctx, tweetsPath, "GET", &tweets); err != nil {
		return tweets, fmt.Errorf("error making request: %v", err)
	}
	return tweets, nil
}

func (a *APIService) GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error) {
	var (
		err   error
		tweet twitter.Tweet
	)
	tweetPath := fmt.Sprintf("%s/api/v1/get_tweet?tweet=%d", a.path, tweetID)
	if err = a.request(ctx, tweetPath, "GET", &tweet); err != nil {
		return tweet, fmt.Errorf("error making request: %v", err)
	}
	return tweet, nil
}

func (a *APIService) request(ctx context.Context, path string, method string, response any) error {
	var err error
	var req *http.Request
	var resp *http.Response
	if req, err = http.NewRequestWithContext(ctx, method, path, nil); err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(response)
	return nil
}
