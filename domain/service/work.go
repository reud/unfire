package service

import (
	"context"
	"time"
	"unfire/domain/client"
	"unfire/domain/model"
	"unfire/infrastructure/persistence"
)

type WorkService interface {
	AddUser() error
	LoadAllTweet() ([]model.Tweet, error)
}

type workService struct {
	client    client.TwitterClient
	datastore persistence.Datastore
	ctx       context.Context
	user      *model.WorkerData
}

const (
	tweetSuffix = "_tweets"
)

// TODO: 定期的にRedisに入ったunixtimeの最小値を見て、それを見て削除を行う。
// unixtime -> twitterIDのハッシュをもっておき、twitterID: []tweetsを更新して、それの最小値を再度格納する。
// TODO(tweetはどこでアップデートする？24hでタイマー掛ける？)
func (ws *workService) AddUser() error {
	tweets, err := ws.LoadAllTweet()
	if err != nil {
		return err
	}

	for _, v := range tweets {
		if err := ws.datastore.AppendString(ws.ctx, ws.user.ID+tweetSuffix, v.IDStr); err != nil {
			return err
		}
	}

	return nil
}

// ここで全ツイートの取得を行う。
func (ws *workService) LoadAllTweet() ([]model.Tweet, error) {
	latestTweets, err := ws.client.FetchTweets()
	if err != nil {
		return []model.Tweet{}, err
	}

	var tweets []model.Tweet
	for _, v := range latestTweets {
		tweets = append(tweets, v)
	}

	for len(latestTweets) != 0 {
		lastID := latestTweets[len(latestTweets)-1].IDStr
		// TODO: API Limit回避の方法について考える。
		time.Sleep(time.Second * 30)
		latestTweets, err = ws.client.FetchTweets(client.SinceId(lastID))
		if err != nil {
			return tweets, err
		}
		for _, v := range latestTweets {
			tweets = append(tweets, v)
		}
	}
	return tweets, nil
}

func NewWorkService(ctx context.Context, user *model.WorkerData, client client.TwitterClient, datastore persistence.Datastore) WorkService {
	return &workService{ctx: ctx, user: user, client: client, datastore: datastore}
}
