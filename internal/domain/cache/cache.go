package cache

import (
	"context"
	"time"
	"twitter-clone/internal/domain/twitter"
)

type Cache interface {
	PushTweet(ctx context.Context, tweet twitter.Tweet) error
	PushToUserFeed(ctx context.Context, userID, tweetID int64) error
	GetUserFeed(ctx context.Context, userID int64, limit int) ([]int, error)
	GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error)
	SetActiveUser(ctx context.Context, userID int64, ttl time.Duration) error
	GetActiveUsers(ctx context.Context) ([]string, error)
}
