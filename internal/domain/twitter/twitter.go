package twitter

import "time"

type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CREATE TABLE tweets (
//
//	id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
//	user_id BIGINT NOT NULL,
//	content TEXT NOT NULL,
//	created_at TIMESTAMP CURRENT_TIMESTAMP,
//	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
//
// );
type Tweet struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Follow struct {
	FollowerID int64     `json:"follower_id" db:"follower_id"`
	FolloweeID int64     `json:"followee_id" db:"followee_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type ChannelTweet struct {
	UserID int64 `json:"user_id"`
	Tweet  Tweet `json:"tweet"`
}
