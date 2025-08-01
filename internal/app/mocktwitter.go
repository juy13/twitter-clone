package app

import (
	"context"
	"twitter-clone/internal/domain/twitter"
)

type FollowFunc func(ctx context.Context, follow twitter.Follow) error
type GetUserFunc func(ctx context.Context, userId int64) (twitter.User, error)
type NewTweetFunc func(ctx context.Context, tweetData twitter.Tweet) error

// Mock implementation of TweeterService
type MockTweeterService struct {
	newTweet       func(ctx context.Context, tweetData twitter.Tweet) error
	getTweet       func(ctx context.Context, id int64) (twitter.Tweet, error)
	getUsersTweets func(ctx context.Context, userId int64) ([]twitter.Tweet, error) // returns tweets made by user
	getTimeline    func(ctx context.Context, userId int64) ([]twitter.Tweet, error) // returns tweets from users the user is following
	getUser        func(ctx context.Context, userId int64) (twitter.User, error)

	// Follow
	followUser   func(ctx context.Context, follow twitter.Follow) error
	getFollowers func(ctx context.Context, userId int64) ([]twitter.User, error)
	getFollowing func(ctx context.Context, userId int64) ([]twitter.User, error)

	// User
	createUser func(ctx context.Context, user twitter.User) (int64, error)
}

// I have to redefine it
// Maybe with use of Options pattern
func NewMockTweeterService(newTweet NewTweetFunc, followUser FollowFunc, getUser GetUserFunc) *MockTweeterService {
	return &MockTweeterService{
		newTweet:   newTweet,
		followUser: followUser,
		getUser:    getUser,
	}
}

func (m *MockTweeterService) NewTweet(ctx context.Context, tweetData twitter.Tweet) error {
	return m.newTweet(ctx, tweetData)
}

func (m *MockTweeterService) GetTweet(ctx context.Context, id int64) (twitter.Tweet, error) {
	return m.getTweet(ctx, id)
}

func (m *MockTweeterService) GetUsersTweets(ctx context.Context, userId int64) ([]twitter.Tweet, error) {
	return m.getUsersTweets(ctx, userId)
}

func (m *MockTweeterService) GetTimeline(ctx context.Context, userId int64) ([]twitter.Tweet, error) {
	return m.getTimeline(ctx, userId)
}

func (m *MockTweeterService) FollowUser(ctx context.Context, follow twitter.Follow) error {
	return m.followUser(ctx, follow)
}

func (m *MockTweeterService) GetUser(ctx context.Context, id int64) (twitter.User, error) {
	return m.getUser(ctx, id)
}

func (m *MockTweeterService) Followers(ctx context.Context, userId int64) ([]twitter.User, error) {
	return m.getFollowers(ctx, userId)
}

func (m *MockTweeterService) Following(ctx context.Context, userId int64) ([]twitter.User, error) {
	return m.getFollowing(ctx, userId)
}

func (m *MockTweeterService) CreateUser(ctx context.Context, user twitter.User) (int64, error) {
	return m.createUser(ctx, user)
}
