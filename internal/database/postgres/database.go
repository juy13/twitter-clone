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

func (p *PostgresDB) NewTweet(ctx context.Context, tweet twitter.Tweet) error {
	query := `
        INSERT INTO tweets (user_id, content)
        VALUES (:user_id, :content)
    `
	_, err := p.db.NamedExecContext(ctx, query, tweet)
	if err != nil {
		return fmt.Errorf("failed to insert tweet: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetTweet(ctx context.Context, id int64) (twitter.Tweet, error) {
	var tweet twitter.Tweet
	query := `
        SELECT id, user_id, content, created_at
        FROM tweets
        WHERE id = $1
    `
	err := p.db.GetContext(ctx, &tweet, query, id)
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

func (p *PostgresDB) FollowUser(ctx context.Context, follow twitter.Follow) error {
	query := `
        INSERT INTO follows (follower_id, followed_id)
        VALUES (:follower_id, :followed_id)
        ON CONFLICT DO NOTHING
    `
	_, err := p.db.NamedExecContext(ctx, query, follow)
	if err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}
	return nil
}
