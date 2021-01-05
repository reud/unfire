package client

import (
	"net/url"
	"unfire/domain/model"
)

type FetchTweetOption struct {
	SinceId *string // ページングに利用する。ツイートのIDを指定すると、これを含まず、これより未来のツイートを取得できる。
}

// Functional Options. ref: https://qiita.com/yoshinori_hisakawa/items/f0c326c99fec116070d4
type FetchTweetOptionFunc func(q *url.Values)

func SinceId(sinceId string) FetchTweetOptionFunc {
	return func(q *url.Values) {
		q.Set("since_id", sinceId)
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
