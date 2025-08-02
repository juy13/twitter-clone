package cache

import (
	"context"
	"time"
	"twitter-clone/internal/domain/twitter"
)

type Cache interface {
	PushTweet(ctx context.Context, tweet twitter.Tweet) error
	PushToUserFeed(ctx context.Context, userID, tweetID int64) error

	GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error)
	SetActiveUser(ctx context.Context, userID int64, ttl time.Duration) error
	GetActiveUsers(ctx context.Context) ([]string, error)

	// Timeline / Feed
	GetUserTimeline(ctx context.Context, userID int64, limit int) ([]int64, error)
	CheckUserTimelineExists(ctx context.Context, userID int64) (bool, error)
	StoreTimeline(ctx context.Context, userID int64, timeline []twitter.Tweet) error
	PushToTweetChannel(ctx context.Context, channelTweet twitter.ChannelTweet) error

	// this part should be moved to pub/sub service but no time rn
	// subscribe to tweets channel
	SubscribeToTweetsChannel(ctx context.Context, channel string) (<-chan string, error)

	// Follower
	GetFollowers(ctx context.Context, userID int64) ([]int64, error)
	SetFollowers(ctx context.Context, userID int64, followers []twitter.User) error
	FollowUser(ctx context.Context, follow twitter.Follow) error
}
