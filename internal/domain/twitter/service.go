package twitter

import "context"

type TwitterServiceI interface {
	NewTweet(ctx context.Context, tweetData Tweet) error
	GetTweet(ctx context.Context, id int64) (Tweet, error)
	GetUsersTweets(ctx context.Context, userId int64) ([]Tweet, error) // returns tweets made by user
	GetTimeline(ctx context.Context, userId int64) ([]Tweet, error)    // returns tweets from users the user is following
	FollowUser(ctx context.Context, follow Follow) error
}
