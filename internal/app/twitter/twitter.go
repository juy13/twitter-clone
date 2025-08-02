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

// TODO return tweet id
func (tw *TwitterService) NewTweet(ctx context.Context, tweetData twitter.Tweet) error {
	var err error
	if tweetData.ID, err = tw.db.NewTweet(ctx, tweetData); err != nil {
		return fmt.Errorf("failed to save tweet: %w", err)
	}
	if err = tw.cache.PushTweet(ctx, tweetData); err != nil {
		return fmt.Errorf("failed to push tweet to cache: %w", err)
	}
	// but in between we can push it to any ML service to analyze the data
	// just for future work
	return nil // actually that's all I think, nothing more
}

func (tw *TwitterService) GetTweet(ctx context.Context, id int64) (twitter.Tweet, error) {
	var tweet twitter.Tweet
	var err error
	if tweet, err = tw.db.GetTweet(ctx, id); err != nil {
		return tweet, fmt.Errorf("failed to get tweet from db: %w", err)
	}
	return tweet, nil
}

func (tw *TwitterService) GetUsersTweets(ctx context.Context, userId int64) ([]twitter.Tweet, error) {
	panic("not implemented")
}

func (tw *TwitterService) GetTimeline(ctx context.Context, userId int64) ([]twitter.Tweet, error) {
	var err error
	var tweets []twitter.Tweet

	if tweets, err = tw.db.GetTimeline(ctx, userId); err != nil {
		return nil, fmt.Errorf("failed to get users tweets from db: %w", err)
	}
	return tweets, nil
}

// Follow part

func (tw *TwitterService) FollowUser(ctx context.Context, follow twitter.Follow) error {
	var err error
	// Here we have to update the cache registry for user
	if err = tw.db.FollowUser(ctx, follow); err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}
	if err = tw.cache.FollowUser(ctx, follow); err != nil {
		return fmt.Errorf("failed to follow user in cache: %w", err)
	}
	return nil
}

func (tw *TwitterService) Followers(ctx context.Context, userId int64) ([]twitter.User, error) {
	var (
		followers []twitter.User
		err       error
	)
	if followers, err = tw.db.Followers(ctx, userId); err != nil {
		return []twitter.User{}, fmt.Errorf("failed to get followers: %w", err)
	}
	// populate them to cache (?)
	return followers, nil
}

func (tw *TwitterService) Following(ctx context.Context, userId int64) ([]twitter.User, error) {
	var (
		following []twitter.User
		err       error
	)
	if following, err = tw.db.Following(ctx, userId); err != nil {
		return []twitter.User{}, fmt.Errorf("failed to get following: %w", err)
	}
	return following, nil
}

// User part
func (tw *TwitterService) CreateUser(ctx context.Context, user twitter.User) (int64, error) {
	var (
		id  int64
		err error
	)
	if id, err = tw.db.CreateUser(ctx, user); err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
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
