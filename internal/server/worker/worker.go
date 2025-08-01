package worker

import (
	"context"
	"fmt"
	"strconv"
	"twitter-clone/internal/domain/cache"
	"twitter-clone/internal/domain/database"
	"twitter-clone/internal/domain/twitter"
)

type Worker struct {
	cache cache.Cache
	db    database.DatabaseI
}

func NewWorker(db database.DatabaseI, cache cache.Cache) *Worker {
	return &Worker{
		db:    db,
		cache: cache,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	var err error

	// it's very specific for the worker
	// maybe for better abstraction it should be as a handler for Subscribe channel
	var pubsub <-chan string

	if pubsub, err = w.cache.SubscribeToTweetsChannel(ctx, "tweets:channel"); err != nil {
		return fmt.Errorf("failed to subscribe to tweets channel: %w", err)
	}
	// hmm a little bit confused about the context

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-pubsub:
			var tweetID int64
			var tweet twitter.Tweet
			if tweetID, err = strconv.ParseInt(msg, 10, 64); err != nil {
				return fmt.Errorf("failed to parse tweet ID: %w", err)
			}

			if tweet, err = w.cache.GetTweet(ctx, tweetID); err != nil {
				return fmt.Errorf("failed to get tweet %d: %w", tweetID, err)
			}

			go w.ProcessTweet(ctx, tweet)
		}
	}
}

func (w *Worker) ProcessTweet(ctx context.Context, tweet twitter.Tweet) error {
	followers, err := w.cache.GetFollowers(ctx, tweet.UserID)
	if err != nil {
		return fmt.Errorf("failed to get followers for user %v: %v", tweet.UserID, err)
	}

	for _, followerID := range followers {
		// do it as a batch to reduce a time!!!
		if err := w.cache.PushToUserFeed(ctx, followerID.ID, tweet.ID); err != nil {
			return fmt.Errorf("failed to push tweet %d to user feed for follower %d: %v", tweet.ID, followerID.ID, err)
		}
		w.sendToWebSocket(ctx, followerID, tweet)
	}

	return nil
}

func (w *Worker) sendToWebSocket(ctx context.Context, user twitter.User, tweet twitter.Tweet) {
	panic("not implemented")
}
