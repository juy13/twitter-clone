package postgres

import (
	"context"
	"fmt"
	"twitter-clone/internal/domain/config"
	"twitter-clone/internal/domain/twitter"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sqlx.DB
}

func NewPostgresDB(config config.DatabaseConfig) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DatabaseHost(),
		config.DatabasePort(),
		config.DatabaseUser(),
		config.DatabasePassword(),
		config.DatabaseName(),
	)

	// Connect to the database using sqlx
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) NewTweet(ctx context.Context, tweet twitter.Tweet) (int64, error) {
	query := `
        INSERT INTO tweets (user_id, content)
        VALUES (:user_id, :content)
        RETURNING id
    ` // https://stackoverflow.com/questions/19167349/postgresql-insert-from-select-returning-id
	var tweetID int64
	err := p.db.QueryRowxContext(ctx, query, tweet).Scan(&tweetID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert tweet: %w", err)
	}
	return tweetID, nil
}

func (p *PostgresDB) GetTweet(ctx context.Context, tweetID int64) (twitter.Tweet, error) {
	var tweet twitter.Tweet
	query := `
        SELECT id, user_id, content, created_at
        FROM tweets
        WHERE id = $1
    `
	err := p.db.GetContext(ctx, &tweet, query, tweetID)
	if err != nil {
		return twitter.Tweet{}, fmt.Errorf("failed to get tweet: %w", err)
	}
	return tweet, nil
}

func (p *PostgresDB) GetUsersTweets(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	var tweets []twitter.Tweet
	query := `
        SELECT id, user_id, content, created_at
        FROM tweets
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	err := p.db.SelectContext(ctx, &tweets, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's tweets: %w", err)
	}
	return tweets, nil
}

func (p *PostgresDB) GetTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	var tweets []twitter.Tweet
	query := `
        SELECT t.id, t.user_id, t.content, t.created_at
        FROM tweets t
        JOIN follows f ON t.user_id = f.followed_id
        WHERE f.follower_id = $1
        ORDER BY t.created_at DESC
    `
	err := p.db.SelectContext(ctx, &tweets, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}
	return tweets, nil
}

// type User struct {
// 	ID          int64     `json:"id"`
// 	Username    string    `json:"username"`
// 	CreatedAt   time.Time `json:"created_at"`
// }

func (p *PostgresDB) GetUser(ctx context.Context, id int64) (twitter.User, error) {
	var user twitter.User
	query := `
        SELECT id, username, created_at
        FROM users
        WHERE id = $1
		`
	err := p.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return twitter.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (p *PostgresDB) CreateUser(ctx context.Context, userData twitter.User) (int64, error) {
	var userID int64
	query := `
		INSERT INTO users (username) VALUES ($1)  RETURNING id;
	`
	err := p.db.QueryRowxContext(ctx, query, userData.Username).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}
	return userID, nil
}

// /////////////////////////////////////////
//
//	Follow part
//
// /////////////////////////////////////////
func (p *PostgresDB) FollowUser(ctx context.Context, follow twitter.Follow) error {
	query := `
        INSERT INTO follows (follower_id, followed_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `
	_, err := p.db.ExecContext(ctx, query, follow.FollowerID, follow.FolloweeID)
	if err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}
	return nil
}

func (p *PostgresDB) Followers(ctx context.Context, userId int64) ([]twitter.User, error) {
	var users []twitter.User
	query := `
        SELECT id, username, follows.created_at
        FROM users
        JOIN follows ON follows.follower_id = users.id
		WHERE follows.followed_id = $1
		`
	err := p.db.SelectContext(ctx, &users, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	return users, nil
}

func (p *PostgresDB) Following(ctx context.Context, userId int64) ([]twitter.User, error) {
	var users []twitter.User
	query := `
        SELECT id, username, follows.created_at as created_at
        FROM users
        JOIN follows ON follows.followed_id = users.id
		WHERE follows.follower_id = $1
		`
	err := p.db.SelectContext(ctx, &users, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	return users, nil
}
