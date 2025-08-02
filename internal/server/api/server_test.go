package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"twitter-clone/internal/domain/twitter"

	app "twitter-clone/internal/app/twitter"

	"github.com/stretchr/testify/assert"
)

func TestNewTweet(t *testing.T) {
	var mockNewTweetFuncNil = func(ctx context.Context, tweetData twitter.Tweet) error { return nil }
	tests := []struct {
		name             string
		queryParams      string
		mockNewTweetFunc app.NewTweetFunc
		mockGetUserFunc  app.GetUserFunc
		expectedStatus   int
		expectedBody     map[string]string
		expectedFollow   twitter.Follow
		method           string
	}{
		{
			name:             "Unknown user",
			queryParams:      "user=1000&followee=2",
			mockNewTweetFunc: mockNewTweetFuncNil,
			mockGetUserFunc: func(ctx context.Context, id int64) (twitter.User, error) {
				return twitter.User{}, errors.New("user not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "user 1000 does not exist"},
			method:         "POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := app.NewMockTweeterService(tt.mockNewTweetFunc, nil, tt.mockGetUserFunc)
			server := &ServerV1{tweeterService: mockService}

			req := httptest.NewRequest(http.MethodGet, "/follow?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			server.newTweet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var result map[string]string
				err := json.NewDecoder(w.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, result)
			} else {
				var result twitter.Follow
				err := json.NewDecoder(w.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFollow.FollowerID, result.FollowerID)
				assert.Equal(t, tt.expectedFollow.FolloweeID, result.FolloweeID)
				assert.False(t, result.CreatedAt.IsZero())
			}

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestFollowUser(t *testing.T) {
	var mockGetUserFuncOK = func(ctx context.Context, id int64) (twitter.User, error) {
		return twitter.User{}, nil
	}
	var mockFollowFuncNil = func(ctx context.Context, follow twitter.Follow) error { return nil }
	tests := []struct {
		name            string
		queryParams     string
		mockFollowFunc  func(ctx context.Context, follow twitter.Follow) error
		mockGetUserFunc func(ctx context.Context, id int64) (twitter.User, error)
		expectedStatus  int
		expectedBody    map[string]string
		expectedFollow  twitter.Follow
	}{
		{
			name:            "Valid input",
			queryParams:     "user=1&followee=2",
			mockFollowFunc:  mockFollowFuncNil,
			mockGetUserFunc: mockGetUserFuncOK,
			expectedStatus:  http.StatusOK,
			expectedFollow: twitter.Follow{
				FollowerID: 1,
				FolloweeID: 2,
			},
		},
		{
			name:            "Missing followee ID",
			queryParams:     "user=1",
			mockFollowFunc:  mockFollowFuncNil,
			mockGetUserFunc: mockGetUserFuncOK,
			expectedStatus:  http.StatusNotFound,
			expectedBody:    map[string]string{"error": "user ID is required"},
		},
		{
			name:            "Invalid user ID",
			queryParams:     "user=invalid&followee=2",
			mockFollowFunc:  mockFollowFuncNil,
			mockGetUserFunc: mockGetUserFuncOK,
			expectedStatus:  http.StatusNotFound,
			expectedBody:    map[string]string{"error": "invalid user ID"},
		},
		{
			name:        "Service failure",
			queryParams: "user=1&followee=2",
			mockFollowFunc: func(ctx context.Context, follow twitter.Follow) error {
				return errors.New("database error")
			},
			mockGetUserFunc: mockGetUserFuncOK,
			expectedStatus:  http.StatusInternalServerError,
			expectedBody:    map[string]string{"error": "Failed to follow user: database error"},
		},
		{
			name:           "Unknown user",
			queryParams:    "user=1000&followee=2",
			mockFollowFunc: mockFollowFuncNil,
			mockGetUserFunc: func(ctx context.Context, id int64) (twitter.User, error) {
				return twitter.User{}, errors.New("user not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "user 1000 does not exist"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := app.NewMockTweeterService(nil, tt.mockFollowFunc, tt.mockGetUserFunc)
			server := &ServerV1{tweeterService: mockService}

			req := httptest.NewRequest(http.MethodGet, "/follow?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			server.followUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var result map[string]string
				err := json.NewDecoder(w.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, result)
			} else {
				var result twitter.Follow
				err := json.NewDecoder(w.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFollow.FollowerID, result.FollowerID)
				assert.Equal(t, tt.expectedFollow.FolloweeID, result.FolloweeID)
				assert.False(t, result.CreatedAt.IsZero())
			}

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}
