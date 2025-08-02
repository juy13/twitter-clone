package api

import (
	"context"
	"twitter-clone/internal/domain/twitter"
)

type API interface {
	GetUser(ctx context.Context, userID int64) (twitter.User, error)
	GetFollowers(ctx context.Context, userID int64) ([]twitter.User, error)
	GetTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error)
	GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error)
}
