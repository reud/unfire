package main

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"strconv"
	"unfire/route"
	"unfire/utils"
)

type Config struct {
	Port                  int    `envconfig:"PORT" default:"8080"`
	TwitterConsumerKey    string `envconfig:"TWITTER_CONSUMER_KEY"`
	TwitterConsumerSecret string `envconfig:"TWITTER_CONSUMER_SECRET"`
}

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

func main() {
	cfg := readConfig()
	e := route.Init()
	if err := e.Start(":" + strconv.Itoa(cfg.Port)); err != nil {
		panic(err)
	}
}
