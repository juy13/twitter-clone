package config

import (
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	API API `yaml:"api,omitempty"`

	// database
	Database Database `yaml:"database,omitempty"`

	// cache
	Cache CacheConfig `yaml:"cache,omitempty"`

	// wss
	WSS WSSConfig `yaml:"wss,omitempty"`

	// metrics
	MetricsServer MetricsConfig `yaml:"metrics,omitempty"`
}

type API struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type Database struct {
	Path     string `yaml:"path,omitempty"`
	User     string `yaml:"user,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Host     string `yaml:"host,omitempty"`
	Name     string `yaml:"name,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type CacheConfig struct {
	Address                        string `yaml:"address"`
	Password                       string `yaml:"password"`
	DB                             int    `yaml:"db"`
	MaxTweets2Keep                 int    `yaml:"max_tweets_to_keep"`
	TweetExpireTimeMinutes         int    `yaml:"tweet_expire_time_minutes"`
	UserFeedExpireTimeMinutes      int    `yaml:"user_feed_expire_time_minutes"`
	TweetTimelineExpireTimeMinutes int    `yaml:"tweet_timeline_expire_time_minutes"`
	MaxTweetsTimelineItems         int    `yaml:"max_tweets_timeline_items"`
}

type WSSConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	API  string `yaml:"api"`
}

type MetricsConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func NewYamlConfig(configFilePath string) (*YamlConfig, error) {
	var (
		err  error
		file *os.File
		data []byte
	)
	dc := &YamlConfig{}

	if file, err = os.Open(configFilePath); err != nil {
		log.Fatal().Msg("can't open config file")
	}
	defer func() {
		_ = file.Close()
	}() // TODO check closings

	if data, err = io.ReadAll(file); err != nil {
		log.Fatal().Msg("can't read config file")
	}
	if err = yaml.Unmarshal(data, dc); err != nil {
		log.Fatal().Msg("can't unmarshal config file")
	}
	return dc, nil
}

///////////////////////////////////
//	API
///////////////////////////////////

func (c YamlConfig) Port() int {
	return c.API.Port
}

func (c *YamlConfig) Host() string {
	return c.API.Host
}

///////////////////////////////////
//	Database
///////////////////////////////////

func (c *YamlConfig) DatabasePath() string {
	return c.Database.Path
}

func (c *YamlConfig) DatabasePort() string {
	return c.Database.Port
}
func (c *YamlConfig) DatabaseHost() string {
	return c.Database.Host
}
func (c *YamlConfig) DatabaseName() string {
	return c.Database.Name
}

func (c *YamlConfig) DatabasePassword() string {
	return c.Database.Password
}

func (c *YamlConfig) DatabaseUser() string {
	return c.Database.User
}

///////////////////////////////////
//	Cache Config
///////////////////////////////////

func (c *YamlConfig) CacheAddress() string {
	return c.Cache.Address
}
func (c *YamlConfig) CachePassword() string {
	return c.Cache.Password
}
func (c *YamlConfig) CacheDB() int {
	return c.Cache.DB
}
func (c *YamlConfig) MaxTweets2Keep() int {
	return c.Cache.MaxTweets2Keep
}
func (c *YamlConfig) TweetExpireTimeMinutes() int {
	return c.Cache.TweetExpireTimeMinutes
}
func (c *YamlConfig) UserFeedExpireTimeMinutes() int {
	return c.Cache.UserFeedExpireTimeMinutes
}
func (c *YamlConfig) TweetTimelineExpireTimeMinutes() int {
	return c.Cache.TweetTimelineExpireTimeMinutes
}
func (c *YamlConfig) MaxTweetsTimelineItems() int {
	return c.Cache.MaxTweetsTimelineItems
}

///////////////////////////////////
//	WSS Config
///////////////////////////////////

func (c *YamlConfig) WSServerHost() string {
	return c.WSS.Host
}
func (c *YamlConfig) WSServerPort() int {
	return c.WSS.Port
}
func (c *YamlConfig) WSServerAPIPath() string {
	return c.WSS.API
}

///////////////////////////////////
//	Metrics Config
///////////////////////////////////

func (c *YamlConfig) MetricsServerHost() string {
	return c.MetricsServer.Host
}
func (c *YamlConfig) MetricsServerPort() int {
	return c.MetricsServer.Port
}
