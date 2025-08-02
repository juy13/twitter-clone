package config

type Config interface {
	APIConfig
	DatabaseConfig
	CacheConfig
}

type APIConfig interface {
	Port() int
	Host() string
}

type DatabaseConfig interface {
	DatabasePath() string
	DatabasePort() string
	DatabasePassword() string
	DatabaseHost() string
	DatabaseName() string
	DatabaseUser() string
}

type CacheConfig interface {
	CacheAddress() string
	CachePassword() string
	CacheDB() int
	MaxTweets2Keep() int
	TweetExpireTimeMinutes() int
	UserFeedExpireTimeMinutes() int
	TweetTimelineExpireTimeMinutes() int
	MaxTweetsTimelineItems() int
}

type WSServerConfig interface {
	WSServerHost() string
	WSServerPort() int
	WSServerAPIPath() string
}
