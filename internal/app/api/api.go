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
	var req *http.Request
	var resp *http.Response
	var user twitter.User
	userPath := fmt.Sprintf("%s/api/v1/get_user?user=%d", a.path, userID)
	if req, err = http.NewRequestWithContext(ctx, "GET", userPath, nil); err != nil {
		return user, fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return user, fmt.Errorf("error making request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(&user)

	if user.ID == 0 {
		return user, fmt.Errorf("invalid user ID")
	}
	return user, nil
}

func (a *APIService) GetFollowers(ctx context.Context, userID int64) ([]twitter.User, error) {
	var (
		err       error
		req       *http.Request
		resp      *http.Response
		user      twitter.User
		followers []twitter.User
	)
	followersPath := fmt.Sprintf("%s/api/v1/followers?user=%d", a.path, user.ID)
	if req, err = http.NewRequestWithContext(ctx, "GET", followersPath, nil); err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(&followers)

	// this should be just a rc or smth
	// if len(followers) == 0 {
	// 	return nil, fmt.Errorf("no followers found for user ID %d", userID)
	// }
	return followers, nil
}

func (a *APIService) GetTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	var (
		err    error
		req    *http.Request
		resp   *http.Response
		tweets []twitter.Tweet
	)
	tweetsPath := fmt.Sprintf("%s/api/v1/tweets?user=%d", a.path, userID)
	if req, err = http.NewRequestWithContext(ctx, "GET", tweetsPath, nil); err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(&tweets)

	return tweets, nil
}

func (a *APIService) GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error) {
	var (
		err   error
		req   *http.Request
		resp  *http.Response
		tweet twitter.Tweet
	)
	tweetPath := fmt.Sprintf("%s/api/v1/get_tweet?tweet=%d", a.path, tweetID)
	if req, err = http.NewRequestWithContext(ctx, "GET", tweetPath, nil); err != nil {
		return tweet, fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return tweet, fmt.Errorf("error making request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(&tweet)

	return tweet, nil
}
