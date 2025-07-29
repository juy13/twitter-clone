package config

import (
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	API API `yaml:"port"`

	// database
	Database Database `yaml:"database"`
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

// /////////////////////////////////
//
//	API
//
// /////////////////////////////////
func (c YamlConfig) Port() int {
	return c.API.Port
}

func (c *YamlConfig) Host() string {
	return c.API.Host
}

// /////////////////////////////////
//
//	Database
//
// /////////////////////////////////
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
