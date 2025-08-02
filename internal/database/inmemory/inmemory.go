package inmemory

import (
	"context"
	"fmt"
	"sync"
	"twitter-clone/internal/domain/twitter"
)

type InMemoryDB struct {
	tweets     map[int64]twitter.Tweet
	userTweets map[int64][]twitter.Tweet
	follows    map[int64]map[int64]struct{}
	users      map[int64]twitter.User
	nextID     int64
	mu         sync.RWMutex
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		tweets:     make(map[int64]twitter.Tweet),
		userTweets: make(map[int64][]twitter.Tweet),
		follows:    make(map[int64]map[int64]struct{}),
		nextID:     1,
	}
}

func (db *InMemoryDB) NewTweet(ctx context.Context, tweet twitter.Tweet) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	tweet.ID = db.nextID
	db.nextID++

	db.tweets[tweet.ID] = tweet
	db.userTweets[tweet.UserID] = append(db.userTweets[tweet.UserID], tweet)

	return tweet.ID, nil
}

func (db *InMemoryDB) GetTweet(ctx context.Context, id int64) (twitter.Tweet, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	tweet, exists := db.tweets[id]
	if !exists {
		return twitter.Tweet{}, fmt.Errorf("tweet with ID %d not found", id)
	}
	return tweet, nil
}

func (db *InMemoryDB) GetUsersTweets(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	tweets, exists := db.userTweets[userID]
	if !exists {
		return nil, fmt.Errorf("no tweets found for user with ID %d", userID)
	}

	result := make([]twitter.Tweet, len(tweets))
	copy(result, tweets)
	return result, nil
}

func (db *InMemoryDB) GetTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var timeline []twitter.Tweet

	if tweets, exists := db.userTweets[userID]; exists {
		timeline = append(timeline, tweets...)
	}

	if followed, exists := db.follows[userID]; exists {
		for followedID := range followed {
			if tweets, exists := db.userTweets[followedID]; exists {
				timeline = append(timeline, tweets...)
			}
		}
	}

	for i := 0; i < len(timeline)-1; i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[i].ID < timeline[j].ID {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}

	return timeline, nil
}

func (db *InMemoryDB) FollowUser(ctx context.Context, follow twitter.Follow) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.follows[follow.FollowerID]; !exists {
		db.follows[follow.FollowerID] = make(map[int64]struct{})
	}

	db.follows[follow.FollowerID][follow.FolloweeID] = struct{}{}
	return nil
}

func (db *InMemoryDB) GetUser(ctx context.Context, id int64) (twitter.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	user, exists := db.users[id]
	if !exists {
		return twitter.User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

func (db *InMemoryDB) CreateUser(ctx context.Context, user twitter.User) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	id := int64(len(db.users)) + 1
	db.users[id] = user
	return id, nil
}
