package app

import (
	"context"
	"twitter-clone/internal/domain/twitter"
)

type TwitterService struct {
	// fields
}

func (tw *TwitterService) NewTweet(ctx context.Context, tweetData twitter.Tweet) error {
	panic("not implemented")
}

func (tw *TwitterService) GetTweet(ctx context.Context, id int64) (twitter.Tweet, error) {
	panic("not implemented")
}

func (tw *TwitterService) GetUsersTweets(ctx context.Context, userId int64) ([]twitter.Tweet, error) {
	panic("not implemented")
}

func (tw *TwitterService) GetTimeline(ctx context.Context, userId int64) ([]twitter.Tweet, error) {
	panic("not implemented")
}
func (tw *TwitterService) FollowUser(ctx context.Context, follow twitter.Follow) error {
	panic("not implemented")
}

func NewTweeterService() *TwitterService {
	return &TwitterService{}
}
