-- +goose Up
-- +goose StatementBegin
-- Users table to store user information
CREATE TABLE users (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tweets (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE follows (
    follower_id BIGINT NOT NULL, -- the one who follows
    followed_id BIGINT NOT NULL, -- the one being followed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followed_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followed_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT no_self_follow CHECK (follower_id != followed_id)
);

CREATE INDEX idx_tweets_user_id ON tweets(user_id);
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_followed_id ON follows(followed_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tweets CASCADE;
DROP TABLE IF EXISTS follows CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP INDEX IF EXISTS idx_tweets_user_id;
DROP INDEX IF EXISTS idx_follows_follower_id;
DROP INDEX IF EXISTS idx_follows_followed_id;
-- +goose StatementEnd
