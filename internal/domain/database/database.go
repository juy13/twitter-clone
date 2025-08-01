package database

import (
	"context"
	"twitter-clone/internal/domain/twitter"
)

type DatabaseI interface {
	NewTweet(ctx context.Context, tweet twitter.Tweet) (int64, error)
	GetTweet(ctx context.Context, id int64) (twitter.Tweet, error)
	GetUsersTweets(ctx context.Context, userID int64) ([]twitter.Tweet, error)
	GetTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error)
	GetUser(ctx context.Context, id int64) (twitter.User, error)

	// Follow
	FollowUser(ctx context.Context, follow twitter.Follow) error
	Followers(ctx context.Context, userId int64) ([]twitter.User, error)
	Following(ctx context.Context, userId int64) ([]twitter.User, error)

	// User part
	CreateUser(ctx context.Context, user twitter.User) (int64, error)
}
