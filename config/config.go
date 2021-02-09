package config

import (
	"unfire/utils"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port                  int    `envconfig:"APP_PORT" default:"5000"`
	TwitterConsumerKey    string `envconfig:"TWITTER_CONSUMER_KEY" default:"mock"`
	TwitterConsumerSecret string `envconfig:"TWITTER_CONSUMER_SECRET" default:"mock"`
	Domain                string `envconfig:"DOMAIN" default:"unfire.reud.app"`
	AdminAPIPassword      string `envconfig:"ADMIN_API_PASSWORD" default:"test"`
}

var sharedConfig = readConfig()

func readConfig() *Config {
	// .envが存在するか、存在するなら読み込み
	if utils.FileExists(".env") {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
	}
	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		panic(err)
	}
	if config.TwitterConsumerKey == "" {
		panic("TwitterConsumerKeyが空です。")
	}
	if config.TwitterConsumerSecret == "" {
		panic("TwitterConsumerSecretが空です。")
	}
	return config
}

func GetInstance() *Config {
	return sharedConfig
}
