package batch

import (
	"context"
	"fmt"
	"log"
	"time"
	client2 "unfire/domain/client"
	"unfire/infrastructure/client"
	"unfire/usecase"
	"unfire/utils"
)

type reloadBatchService struct {
	interval time.Duration
	dc       usecase.DatastoreController
}

func NewReloadBatchService(interval time.Duration, dc usecase.DatastoreController) BatchService {
	return &reloadBatchService{
		interval: interval,
		dc:       dc,
	}
}

func (bs *reloadBatchService) Start() {
	ticker := time.NewTicker(bs.interval)
	go func() {
		for t := range ticker.C {
			fmt.Printf("batch started: %+v\n", t)
			reloadTask(bs.dc)
			fmt.Printf("batch finished: %+v\n", t)
		}
	}()
}

func (bs *reloadBatchService) StartOnce() {
	fmt.Println("[force]reload batch started")
	reloadTask(bs.dc)
	fmt.Println("[force]reload batch finished")
}

// ユーザ一覧をO(n)で取得して、Waiting状態のユーザに対してツイートのロードを行う。
func reloadTask(dc usecase.DatastoreController) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	// 予防のために追加
	defer cancel()

	// TODO: N+1だけど解消のしようがない気もする。
	// FIXME: ここgoroutineにすれば速度は間違いなく上がる。別に今は速度要らないけど、時期を見て確認する。
	users := dc.GetAllUsers(ctx)
	for _, twitterID := range users {
		if status := dc.GetUserStatus(ctx, twitterID); status == utils.Waiting {
			dc.SetUserStatus(ctx, twitterID, utils.Initializing)

			at := dc.PickAuthorizeData(ctx, twitterID)

			tcii := client.NewTwitterClientInitializer()
			cred, err := tcii.NewTwitterClient(at)
			// tokenの有効期限切れの場合なども考えてハンドリングする。
			if err != nil {
				log.Printf("[reload task] generate client failed (reloadTask): user status change to deleted. err: %+v", err)
				dc.SetUserStatus(ctx, twitterID, utils.Deleted)
				continue
			}

			tweets, err := cred.FetchTweets(client2.GetAll())
			if err != nil {
				log.Printf("[reload task] reload tweets failed (reloadTask): user status change to deleted. err: %+v", err)
				dc.SetUserStatus(ctx, twitterID, utils.Deleted)
				continue
			}

			// ツイートが入っていない場合はWaitingステータスに戻す
			if len(tweets) == 0 {
				dc.SetUserStatus(ctx, twitterID, utils.Waiting)
				continue
			}

			// ツイートが入っている場合はツイートの中で一番古い時間をタイムラインに格納する
			dc.InsertTweetToTimeLine(ctx, twitterID, tweets[len(tweets)-1])

			// 全ツイートの保管
			dc.StoreAllTweet(ctx, twitterID, tweets)

			// タスク終了後はユーザのステータスをワーキングにする。
			dc.SetUserStatus(ctx, twitterID, utils.Working)
		}
	}
}
