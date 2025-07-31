package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"twitter-clone/internal/domain/config"
	"twitter-clone/internal/domain/twitter"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client

	maxTweets2Keep     int
	tweetExpireTime    time.Duration
	userFeedExpireTime time.Duration
}

func NewRedisCache(config config.CacheConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.CacheAddress(),
		Password: config.CachePassword(),
		DB:       config.CacheDB(),
	})
	return &RedisCache{
		client:             client,
		maxTweets2Keep:     config.MaxTweets2Keep(),
		tweetExpireTime:    time.Duration(config.TweetExpireTimeMinutes()),
		userFeedExpireTime: time.Duration(config.TweetExpireTimeMinutes()),
	}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) PushTweet(ctx context.Context, tweet twitter.Tweet) error {
	data, err := json.Marshal(tweet)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet %v: %v", tweet.ID, err)
	}

	pipe := c.client.TxPipeline()

	tweetKey := fmt.Sprintf("tweet:%v", tweet.ID)
	pipe.Set(ctx, tweetKey, data, c.tweetExpireTime)
	pipe.Publish(ctx, "tweets:channel", tweet.ID) // here is coming the pub/sub model so the worker have to know when the tweet is published

	pipe.LPush(ctx, "tweets:global", tweet.ID)
	pipe.LTrim(ctx, "tweets:global", 0, int64(c.maxTweets2Keep)-1)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to push tweet %v to global list: %v", tweet.ID, err)
	}

	return nil
}

func (c *RedisCache) PushToUserFeed(ctx context.Context, userID, tweetID int64) error {
	return nil
}

func (c *RedisCache) GetUserFeed(ctx context.Context, userID int64, limit int) ([]int, error) {
	return nil, nil
}

func (c *RedisCache) GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error) {
	var tweet twitter.Tweet
	return tweet, nil
}

func (c *RedisCache) SetActiveUser(ctx context.Context, userID int64, ttl time.Duration) error {
	return nil
}

func (c *RedisCache) GetActiveUsers(ctx context.Context) ([]string, error) {
	return nil, nil
}
