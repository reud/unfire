package batch

import (
	"context"
	"fmt"
	"log"
	"time"
	"unfire/infrastructure/client"
	"unfire/usecase"
	"unfire/utils"
)

type deleteBatchService struct {
	interval time.Duration
	dc       usecase.DatastoreController
}

func NewDeleteBatchService(interval time.Duration, dc usecase.DatastoreController) BatchService {
	return &deleteBatchService{
		interval: interval,
		dc:       dc,
	}
}

func (bs *deleteBatchService) Start() {
	ticker := time.NewTicker(bs.interval)
	go func() {
		for t := range ticker.C {
			fmt.Printf("batch started: %+v\n", t)
			if err := deleteTask(bs.dc); err != nil {
				fmt.Printf("batch error occured: %+v\n", err)
			}
			fmt.Printf("batch finished: %+v\n", t)
		}
	}()
}

func (bs *deleteBatchService) StartOnce() {
	fmt.Println("[force] delete batch started")
	if err := deleteTask(bs.dc); err != nil {
		panic(err)
	}
	fmt.Println("[force] delete batch finished")
}

// TODO: 同時に二個以上走らせると死ぬ。channelで状態を通知するとかやる？
func deleteTask(dc usecase.DatastoreController) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	// 予防のために追加
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			err, ok := ctx.Value("error").(error)
			if !ok {
				return err
			}
			return nil
		default:
			log.Printf("[delete task] picking oldest tweets")
			// 保存されているツイートの中で最も古いものを取得する。
			t, userID, err := dc.GetOldestTweetInfoFromTimeLine(ctx)
			if err != nil {
				// ここでツイートが存在しない可能性もあるのでエラーを返したりはしない。
				log.Printf("[delete task] failed to fetch OldestTweet: %+v\n ", err)
				return nil
			}

			// 削除済みユーザの場合はそれをpopしてスキップ
			if status := dc.GetUserStatus(ctx, userID); status == utils.Deleted {
				dc.PopOldestTweetInfoFromTimeLine(ctx)
				continue
			}

			cred := dc.PickAuthorizeData(ctx, userID)

			tcii := client.NewTwitterClientInitializer()
			tc, err := tcii.NewTwitterClient(cred)
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// その最小値が24時間以上経過しているかどうか。 経過していない場合は終了
			if !time.Now().After(t.AddDate(0, 0, 1)) {
				log.Printf("[delete task] its new tweet: fetch time: %+v \n", t.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
				cancel()
				continue
			}

			// 24時間以上経過しているならばその最小値を消す。
			dc.PopOldestTweetInfoFromTimeLine(ctx)

			// そのuserIDのtweetを後ろから取る。
			for {

				tweetID, ok := dc.GetUserLastTweet(ctx, userID)
				if !ok {
					// ツイートが存在していない可能性もあるのでWaitingにする。
					dc.SetUserStatus(ctx, tweetID, utils.Waiting)
					break
				}

				tweet, err := tc.FetchTweetFromIDStr(tweetID)
				// tweetIDからtweet情報を取ってくることに失敗した場合。
				if err != nil {
					// ユーザが手動で削除した場合もある。その時は次のツイートを取り直す。
					log.Printf("[delete task] maybe, user deleted this tweet yourself: %+v\n", err)
					continue
				}

				t, err := time.Parse("2006-01-02T15:04:05.000Z", tweet.CreatedAt)
				if err != nil {
					// パースがコケるのは存在し得ないのでpanic
					log.Fatalf("[delete task] unexpected error in parsing time: %+v", err)
				}

				// そのツイートの投稿時間が24時間以上経過しているかどうか
				// 1日経っていないなら、それをtweetsに戻してからbreak
				if !time.Now().After(t.AddDate(0, 0, 1)) {
					log.Printf("[delete task] 最古のツイートが一日経っていないので終了します。\n")
					dc.InsertTweetToTimeLine(ctx, userID, *tweet)
					dc.PutUserLastTweet(ctx, userID, tweetID)
					cancel()
					break
				}

				log.Printf("[delete task] deleting... %+v\n", tweet.ID)
				// ツイートの削除
				if err := tc.DestroyTweet(tweet.ID); err != nil {
					log.Printf("%+v\n", err)
					ctx = context.WithValue(ctx, "error", err)
					cancel()
					break
				}

			}

		}

	}

}
