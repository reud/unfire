package batch

import (
	"context"
	"fmt"
	"log"
	"time"
	"unfire/infrastructure/client"
	"unfire/usecase"
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
			log.Printf("picking oldest tweets")
			// 保存されているツイートの中で最も古いものを取得する。
			t, userID, err := dc.GetOldestTweetInfoFromTimeLine(ctx)
			if err != nil {
				log.Printf("error in deleteTask: %+v\n ", err)
				return err
			}

			cred := dc.PickAuthorizeData(ctx, userID)
			log.Printf("created token:%+v secret:%+v\n", cred.Token, cred.Secret)

			tc, err := client.NewTwitterClient(cred)
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// その最小値が24時間以上経過しているかどうか。 経過していない場合は終了
			if !time.Now().After(t.AddDate(0, 0, 1)) {
				log.Printf("its new tweet \n")
				cancel()
				continue
			}

			// 24時間以上経過しているならばその最小値を消す。
			dc.PopOldestTweetInfoFromTimeLine(ctx)

			// そのuserIDのtweetを後ろから取る。
			for {

				tweetID, ok := dc.GetUserLastTweet(ctx, userID)
				if !ok {
					break
				}

				tweet, err := tc.FetchTweetFromIDStr(tweetID)

				if err != nil {
					log.Printf("%+v\n", err)
					break
				}

				log.Printf("tweet fetch success  %+v", tweet.ID)

				t, err := time.Parse("2006-01-02T15:04:05.000Z", tweet.CreatedAt)
				if err != nil {
					log.Printf("%+v\n", err)
					ctx = context.WithValue(ctx, "error", err)
					cancel()
					continue
				}

				// そのツイートの投稿時間が24時間以上経過しているかどうか
				// 1日経っていないならbreak
				if !time.Now().After(t.AddDate(0, 0, 1)) {
					log.Printf("最古のツイートが一日経っていないので終了します。\n")
					break
				}

				log.Printf("deleting... %+v\n", tweet.ID)
				// ツイートの削除
				if err := tc.DestroyTweet(tweet.ID); err != nil {
					log.Printf("%+v\n", err)
					ctx = context.WithValue(ctx, "error", err)
					cancel()
					continue
				}

			}

		}

	}

}
