package client

import (
	"unfire/domain/model"
)

type FetchTweetOption struct {
	GetAll bool // すべてのツイートを取得する。
}

// Functional Options. ref: https://qiita.com/yoshinori_hisakawa/items/f0c326c99fec116070d4
type FetchTweetOptionFunc func(option *FetchTweetOption)

func GetAll() FetchTweetOptionFunc {
	return func(option *FetchTweetOption) {
		option.GetAll = true
	}
}

type TwitterClient interface {
	FetchMe() *model.WorkerData
	FetchTweets(options ...FetchTweetOptionFunc) ([]model.Tweet, error)
	FetchFavorites() ([]model.Tweet, error)
	DestroyTweet(tweetID string) error
	DestroyFavorite(tweetID string) error
	FetchTweetFromIDStr(tweetID string) (*model.Tweet, error)
}
