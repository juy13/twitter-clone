package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
	"twitter-clone/internal/domain/twitter"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

type MockCacheConfig struct{}

func (mc *MockCacheConfig) CacheAddress() string                { return "test" }
func (mc *MockCacheConfig) CachePassword() string               { return "test" }
func (mc *MockCacheConfig) CacheDB() int                        { return 1 }
func (mc *MockCacheConfig) MaxTweets2Keep() int                 { return 1 }
func (mc *MockCacheConfig) TweetExpireTimeMinutes() int         { return 1 }
func (mc *MockCacheConfig) UserFeedExpireTimeMinutes() int      { return 1 }
func (mc *MockCacheConfig) TweetTimelineExpireTimeMinutes() int { return 1 }
func (mc *MockCacheConfig) MaxTweetsTimelineItems() int         { return 1 }

var mockConfig MockCacheConfig

func TestPushTweet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	c := NewRedisCache(&mockConfig)
	c.client = db

	tweet := twitter.Tweet{
		ID:      123,
		UserID:  1,
		Content: "Hello World!",
	}

	data, err := json.Marshal(tweet)
	require.NoError(t, err)

	ctx := context.Background()

	mock.ExpectTxPipeline()
	mock.ExpectSet(fmt.Sprintf("tweet:%v", tweet.ID), data, c.tweetExpireTime*time.Minute).SetVal("OK")
	mock.ExpectPublish("tweets:channel", tweet.ID).SetVal(1)
	mock.ExpectLPush("tweets:global", tweet.ID).SetVal(1)
	mock.ExpectLTrim("tweets:global", 0, int64(c.maxTweets2Keep)-1).SetVal("OK")
	mock.ExpectTxPipelineExec()

	err = c.PushTweet(ctx, tweet)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

// PushToUserFeed(ctx context.Context, userID, tweetID int64) error
func TestPushToUserFeed(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	c := NewRedisCache(&mockConfig)
	c.client = db

	ctx := context.Background()

	userID := int64(1)
	tweetID := int64(1)
	mock.ExpectTxPipeline()
	mock.ExpectLPush(fmt.Sprintf("timeline:%v", 1), tweetID).SetVal(1)
	mock.ExpectLTrim(fmt.Sprintf("timeline:%v", 1), 0, int64(c.maxTweetsTimelineItems)-1).SetVal("OK")
	mock.ExpectExpire(fmt.Sprintf("timeline:%v", 1), c.tweetTimelineExpireTime*time.Minute).SetVal(true)
	mock.ExpectTxPipelineExec()

	err := c.PushToUserFeed(ctx, userID, tweetID)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

// GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error)
func TestGetTweet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db

	ctx := context.Background()
	tweetID := int64(123)

	expectedTweet := twitter.Tweet{
		ID:      tweetID,
		UserID:  1,
		Content: "Hello from Redis!",
	}

	data, err := json.Marshal(expectedTweet)
	require.NoError(t, err)

	mock.ExpectGet(fmt.Sprintf("tweet:%v", tweetID)).SetVal(string(data))

	tweet, err := cache.GetTweet(ctx, tweetID)

	require.NoError(t, err)
	require.Equal(t, expectedTweet, tweet)
	require.NoError(t, mock.ExpectationsWereMet())
}

// GetFollowers(ctx context.Context, userID int64) ([]int64, error)
func TestGetFollowers(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db
	ctx := context.Background()
	userID := int64(1)
	key := fmt.Sprintf("followers:%d", userID)

	redisFollowers := []string{"4", "3", "2"}
	mock.ExpectLRange(key, 0, -1).SetVal(redisFollowers)

	expected := []int64{2, 3, 4}

	followers, err := cache.GetFollowers(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, expected, followers)
	require.NoError(t, mock.ExpectationsWereMet())
}

// SetFollowers(ctx context.Context, userID int64, followers []twitter.User) error
func TestSetFollowers(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db
	ctx := context.Background()
	userID := int64(1)
	key := fmt.Sprintf("followers:%d", userID)

	followers := []twitter.User{
		{ID: 2},
		{ID: 3},
		{ID: 4},
	}

	mock.ExpectTxPipeline()
	mock.ExpectDel(key).SetVal(1)

	for _, f := range followers {
		mock.ExpectLPush(key, f.ID).SetVal(1)
	}

	mock.ExpectTxPipelineExec()
	err := cache.SetFollowers(ctx, userID, followers)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// FollowUser(ctx context.Context, follow twitter.Follow) error
func TestFollowUser(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db
	ctx := context.Background()

	follow := twitter.Follow{
		FollowerID: 101,
		FolloweeID: 202,
	}

	key := fmt.Sprintf("followers:%d", follow.FolloweeID)

	mock.ExpectTxPipeline()
	mock.ExpectLPush(key, follow.FollowerID).SetVal(1)
	mock.ExpectTxPipelineExec()

	err := cache.FollowUser(ctx, follow)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// // Timeline / Feed
// GetUserTimeline(ctx context.Context, userID int64, limit int) ([]int64, error)
func TestGetUserTimeline(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db
	ctx := context.Background()
	userID := int64(1)
	limit := 3
	key := fmt.Sprintf("timeline:%d", userID)

	redisValues := []string{"3", "2", "1"}
	mock.ExpectLRange(key, int64(0), int64(limit-1)).SetVal(redisValues)

	expected := []int64{1, 2, 3}

	timeline, err := cache.GetUserTimeline(ctx, userID, limit)
	require.NoError(t, err)
	require.Equal(t, expected, timeline)
	require.NoError(t, mock.ExpectationsWereMet())
}

// CheckUserTimelineExists(ctx context.Context, userID int64) (bool, error)
func TestCheckUserTimelineExists(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db
	ctx := context.Background()
	userID := int64(1)
	key := fmt.Sprintf("timeline:%d", userID)

	mock.ExpectExists(key).SetVal(1)

	exists, err := cache.CheckUserTimelineExists(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, true, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

// StoreTimeline(ctx context.Context, userID int64, timeline []twitter.Tweet) error
func TestStoreTimeline(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer func() {
		_ = db.Close() // lint
	}()

	cache := NewRedisCache(&mockConfig)
	cache.client = db
	ctx := context.Background()
	userID := int64(1)

	tweets := []twitter.Tweet{
		{
			ID:      1,
			UserID:  2,
			Content: "Hello from Redis!",
		},
		{
			ID:      2,
			UserID:  3,
			Content: "Hello next one!",
		},
	}

	feedKey := fmt.Sprintf("timeline:%d", userID)
	mock.ExpectTxPipeline()
	for _, tweet := range tweets {
		mock.ExpectLPush(feedKey, tweet.ID).SetVal(tweet.ID)
		mock.ExpectLTrim(feedKey, 0, int64(cache.maxTweetsTimelineItems)).SetVal("OK")
		mock.ExpectExpire(feedKey, cache.tweetTimelineExpireTime*time.Minute).SetVal(true)
	}
	mock.ExpectTxPipelineExec()

	err := cache.StoreTimeline(ctx, userID, tweets)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// // this part should be moved to pub/sub service but no time rn
// // subscribe to tweets channel
// PushToTweetChannel(ctx context.Context, channelTweet twitter.ChannelTweet) error
// SubscribeToTweetsChannel(ctx context.Context, channel string) (<-chan string, error)
