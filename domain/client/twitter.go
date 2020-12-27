package client

import (
	"unfire/domain/model"
)

type FetchTweetOption struct {
	SinceId *string // ページングに利用する。ツイートのIDを指定すると、これを含まず、これより未来のツイートを取得できる。
}

func NewFetchTweetOption() *FetchTweetOption {
	return &FetchTweetOption{SinceId: nil}
}

// Functional Options. ref: https://qiita.com/yoshinori_hisakawa/items/f0c326c99fec116070d4
type FetchTweetOptionFunc func(option *FetchTweetOption)

func SinceId(sinceId string) FetchTweetOptionFunc {
	return func(option *FetchTweetOption) {
		option.SinceId = &sinceId
	}
}

type TwitterClient interface {
	FetchTweets(options ...FetchTweetOptionFunc) ([]model.Tweet, error)
	FetchFavorites() error
	DestroyTweet(tweetID string) error
	DestroyFavorite(tweetID string) error
}
