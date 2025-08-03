package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	maxTweetsTimelineItems  int
	tweetTimelineExpireTime time.Duration
}

func NewRedisCache(config config.CacheConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.CacheAddress(),
		Password: config.CachePassword(),
		DB:       config.CacheDB(),
	})
	return &RedisCache{
		client:                  client,
		maxTweets2Keep:          config.MaxTweets2Keep(),
		tweetExpireTime:         time.Duration(config.TweetExpireTimeMinutes()),
		userFeedExpireTime:      time.Duration(config.TweetExpireTimeMinutes()),
		maxTweetsTimelineItems:  config.MaxTweetsTimelineItems(),
		tweetTimelineExpireTime: time.Duration(config.TweetTimelineExpireTimeMinutes()),
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
	var err error
	pipe := c.client.TxPipeline()
	feedKey := fmt.Sprintf("timeline:%d", userID)
	pipe.LPush(ctx, feedKey, tweetID)
	pipe.LTrim(ctx, feedKey, 0, int64(c.maxTweetsTimelineItems))
	pipe.Expire(ctx, feedKey, c.tweetTimelineExpireTime*time.Minute)

	///// did I lost it? yep I did
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to push tweet %v to global list: %v", tweetID, err)
	}
	return nil
}

/////////////////////////////////////
//	Timeline / Feed
////////////////////////////////////

func (c *RedisCache) GetUserTimeline(ctx context.Context, userID int64, limit int) ([]int64, error) {
	key := fmt.Sprintf("timeline:%d", userID)
	values, err := c.client.LRange(ctx, key, 0, int64(limit)-1).Result() // TODO here limit should be from config & related to maxTweetsTimelineItems
	if err != nil {
		return nil, err
	}

	result := make([]int64, len(values))
	for i, v := range values {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse int64 from '%s': %w", v, err)
		}
		result[len(values)-i-1] = id
	}
	return result, nil
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

func (c *RedisCache) StoreTimeline(ctx context.Context, userID int64, tweets []twitter.Tweet) error {
	pipe := c.client.TxPipeline()
	feedKey := fmt.Sprintf("timeline:%d", userID)
	for _, tweet := range tweets {
		pipe.LPush(ctx, feedKey, tweet.ID)
		pipe.LTrim(ctx, feedKey, 0, int64(c.maxTweetsTimelineItems))
		pipe.Expire(ctx, feedKey, c.tweetTimelineExpireTime*time.Minute)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to push tweets to users %d timeline: %v", userID, err)
	}
	return nil
}

func (c *RedisCache) PushToTweetChannel(ctx context.Context, channelTweet twitter.ChannelTweet) error {
	var err error
	data, err := json.Marshal(channelTweet)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet %v: %v", channelTweet.Tweet.ID, err)
	}
	pipe := c.client.TxPipeline()

	pipe.Publish(ctx, "workers:channel", data)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to public tweet: %v, %w", channelTweet.Tweet.ID, err)
	}
	return nil
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
				_ = pubsub.Close() // lint issue
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
// TODO fro better optimization keep just followers ID
func (c *RedisCache) GetFollowers(ctx context.Context, userID int64) ([]int64, error) {
	key := fmt.Sprintf("followers:%d", userID)
	values, err := c.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	result := make([]int64, len(values))
	for i, v := range values {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse int64 from '%s': %w", v, err)
		}
		result[len(values)-i-1] = id
	}
	return result, nil
}

// It overwrites the followers list!!!
func (c *RedisCache) SetFollowers(ctx context.Context, userID int64, followers []twitter.User) error {
	// or maybe we just have to store only ids?
	// yes so, I'll put indexes in list
	var err error
	pipe := c.client.TxPipeline()
	followerKey := fmt.Sprintf("followers:%d", userID)
	pipe.Del(ctx, followerKey)
	for _, follower := range followers {
		pipe.LPush(ctx, followerKey, follower.ID)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

func (c *RedisCache) FollowUser(ctx context.Context, follow twitter.Follow) error {
	var err error
	pipe := c.client.TxPipeline()
	followerKey := fmt.Sprintf("followers:%d", follow.FolloweeID)
	pipe.LPush(ctx, followerKey, follow.FollowerID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}
