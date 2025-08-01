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
	pipe.Set(ctx, tweetKey, data, c.tweetExpireTime*time.Minute)
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
	pipe := c.client.TxPipeline()
	feedKey := fmt.Sprintf("timeline:%d", userID)
	pipe.LPush(ctx, feedKey, tweetID)
	pipe.LTrim(ctx, feedKey, 0, 99) // TODO add to config as a pram for redis
	pipe.Expire(ctx, feedKey, 24*time.Hour)
	return nil
}

/////////////////////////////////////
//	Timeline / Feed
////////////////////////////////////

func (c *RedisCache) GetUserTimeline(ctx context.Context, userID int64, limit int) ([]int, error) {
	return nil, nil
}

func (c *RedisCache) CheckUserTimelineExists(ctx context.Context, userID int64) (bool, error) {
	var err error
	var exists int64
	timelineKey := fmt.Sprintf("timeline::%v", userID)
	if exists, err = c.client.Exists(ctx, timelineKey).Result(); err != nil {
		return false, err
	}
	// check it i'm not sure
	return exists != 0, nil
}

/////////////////////////////////////////////////////////////////////////////////////////////////

func (c *RedisCache) GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error) {
	var tweet twitter.Tweet
	tweetData, err := c.client.Get(ctx, fmt.Sprintf("tweet:%v", tweetID)).Result()
	if err != nil {
		return tweet, fmt.Errorf("failed to get tweet %v: %v", tweetID, err)
	}
	if err := json.Unmarshal([]byte(tweetData), &tweet); err != nil {
		return tweet, fmt.Errorf("failed to unmarshal tweet %v: %v", tweetID, err)
	}
	return tweet, nil
}

func (c *RedisCache) SetActiveUser(ctx context.Context, userID int64, ttl time.Duration) error {
	return nil
}

func (c *RedisCache) GetActiveUsers(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (c *RedisCache) SubscribeToTweetsChannel(ctx context.Context, channel string) (<-chan string, error) {
	pubsub := c.client.Subscribe(ctx, channel)
	ch := pubsub.Channel()
	chanRet := make(chan string)
	go func() {
		for {
			select {
			case <-ctx.Done():
				pubsub.Close()
				close(chanRet)
				return
			default:
				msg, ok := <-ch
				if !ok {
					return // send error and die
				}
				chanRet <- msg.Payload
			}
		}
	}()
	return chanRet, nil
}

// Follow part

func (c *RedisCache) GetFollowers(ctx context.Context, userID int64) ([]twitter.User, error) {
	var followers []twitter.User
	err := c.client.Get(ctx, fmt.Sprintf("followers:%d", userID)).Scan(&followers)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get followers from cache: %w", err)
	}
	return followers, nil
}

func (c *RedisCache) SetFollowers(ctx context.Context, userID int64, followers []twitter.User) error {
	// or maybe we just have to store only ids?
	followersJSON, err := json.Marshal(followers)
	if err != nil {
		return fmt.Errorf("failed to marshal followers: %w", err)
	}
	err = c.client.Set(ctx, fmt.Sprintf("followers:%d", userID), followersJSON, 24*time.Hour).Err() // TODO set to the config expire time for the user
	if err != nil {
		return fmt.Errorf("failed to set followers in cache: %w", err)
	}
	return nil
}
