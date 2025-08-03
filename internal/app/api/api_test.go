package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"twitter-clone/internal/app/api"
	"twitter-clone/internal/domain/twitter"

	"github.com/stretchr/testify/require"
)

//	type Tweet struct {
//	    ID        int64     `json:"id" db:"id"`
//	    UserID    int64     `json:"user_id" db:"user_id"`
//	    Content   string    `json:"content" db:"content"`
//	    CreatedAt time.Time `json:"created_at" db:"created_at"`
//	}
func TestGetTwee(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/get_tweet", r.URL.Path)
		require.Equal(t, "123", r.URL.Query().Get("tweet"))
		response := twitter.Tweet{
			ID:        1,
			UserID:    123,
			Content:   "Hello World",
			CreatedAt: time.Now(),
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	api := api.NewAPIService(server.URL)
	tweet, err := api.GetTweet(context.Background(), 123)
	require.NoError(t, err)
	require.Equal(t, int64(1), tweet.ID)
}

func TestGetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/get_user", r.URL.Path)
		require.Equal(t, "42", r.URL.Query().Get("user"))

		response := twitter.User{
			ID:       42,
			Username: "testuser",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	api := api.NewAPIService(server.URL)
	user, err := api.GetUser(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, int64(42), user.ID)
	require.Equal(t, "testuser", user.Username)
}

func TestGetFollowers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/followers", r.URL.Path)
		require.Equal(t, "42", r.URL.Query().Get("user"))

		response := []twitter.User{
			{ID: 1, Username: "follower1"},
			{ID: 2, Username: "follower2"},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	api := api.NewAPIService(server.URL)
	followers, err := api.GetFollowers(context.Background(), 42)
	require.NoError(t, err)
	require.Len(t, followers, 2)
	require.Equal(t, "follower1", followers[0].Username)
	require.Equal(t, int64(2), followers[1].ID)
}

func TestGetTimeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/tweets", r.URL.Path)
		require.Equal(t, "42", r.URL.Query().Get("user"))

		response := []twitter.Tweet{
			{ID: 100, UserID: 42, Content: "First tweet", CreatedAt: time.Now()},
			{ID: 101, UserID: 42, Content: "Second tweet", CreatedAt: time.Now()},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	api := api.NewAPIService(server.URL)
	tweets, err := api.GetTimeline(context.Background(), 42)
	require.NoError(t, err)
	require.Len(t, tweets, 2)
	require.Equal(t, int64(100), tweets[0].ID)
	require.Equal(t, "Second tweet", tweets[1].Content)
}
