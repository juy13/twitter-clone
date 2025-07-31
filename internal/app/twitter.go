package app

import (
	"context"
	"fmt"
	"twitter-clone/internal/domain/cache"
	"twitter-clone/internal/domain/database"
	"twitter-clone/internal/domain/twitter"
)

type TwitterService struct {
	db    database.DatabaseI
	cache cache.Cache
}

func NewTweeterService(db database.DatabaseI, cache cache.Cache) *TwitterService {
	return &TwitterService{
		db:    db,
		cache: cache,
	}
}

func (tw *TwitterService) NewTweet(ctx context.Context, tweetData twitter.Tweet) error {
	var err error
	if tweetData.ID, err = tw.db.NewTweet(ctx, tweetData); err != nil {
		return fmt.Errorf("failed to save tweet: %w", err)
	}
	if err = tw.cache.PushTweet(ctx, tweetData); err != nil {
		return fmt.Errorf("failed to push tweet to cache: %w", err)
	}
	// 1. Insert in database
	// 2. Insert in redis
	// 3. return error
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

func (tw *TwitterService) GetUser(ctx context.Context, id int64) (twitter.User, error) {
	var (
		user twitter.User
		err  error
	)
	if user, err = tw.db.GetUser(ctx, id); err != nil {
		return twitter.User{}, fmt.Errorf("failed to get user from database: %w", err)
	}
	return user, nil
}
