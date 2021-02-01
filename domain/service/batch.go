package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unfire/infrastructure/client"
	"unfire/infrastructure/persistence"
	"unfire/utils"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/pkg/errors"
)

type BatchService interface {
	Start()
}

type batchService struct {
	interval time.Duration
	ds       persistence.Datastore
}

func NewBatchService(interval time.Duration, ds persistence.Datastore) BatchService {
	return &batchService{
		interval: interval,
		ds:       ds,
	}
}

func (bs *batchService) Start() {
	ticker := time.NewTicker(bs.interval)
	go func() {
		for t := range ticker.C {
			fmt.Printf("batch started: %+v\n", t)
			if err := task(bs.ds); err != nil {
				fmt.Printf("batch error occured: %+v\n", err)
			}
			fmt.Printf("batch finished: %+v\n", t)
		}
	}()
}

func task(ds persistence.Datastore) error {
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
			// 最小値を持ってくる
			data, err := ds.GetMinElement(ctx, utils.TimeLine)
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			sp := strings.Split(data, "_")
			if len(data) != 2 {
				err := errors.New(fmt.Sprintf("bad data got: %+v", sp))
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			tweetTime, err := strconv.ParseInt(sp[0], 10, 64)
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			userID := sp[1]
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// 取得したuserIDのaccess tokenを取り出す。
			atStr, err := ds.GetHash(ctx, utils.TokenSuffix+userID, "at")
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// 取得したuserIDのsecret tokenを取り出す
			secStr, err := ds.GetHash(ctx, utils.TokenSuffix+userID, "sec")
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// 認可情報を作成
			cred := &oauth.Credentials{
				Token:  atStr,
				Secret: secStr,
			}

			tc, err := client.NewTwitterClient(cred)
			if err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// その最小値が24時間以上経過しているかどうか。
			t := time.Unix(tweetTime, 0)
			// 経過していない場合は終了
			if !time.Now().After(t.AddDate(0, 0, 1)) {
				log.Printf("its new tweet \n")
				cancel()
				continue
			}

			// 24時間以上経過しているならばその最小値を消す。
			if err := ds.PopMin(ctx, utils.TimeLine); err != nil {
				log.Printf("%+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// そのuserIDのtweetを後ろから取る。
			for {
				lastTweetID, err := ds.LastPop(ctx, userID+utils.TweetsSuffix)
				if err != nil {
					log.Printf("%+v\n", err)
					ctx = context.WithValue(ctx, "error", err)
					cancel()
					continue
				}

				tweet, err := tc.FetchTweetFromIDStr(lastTweetID)
				if err != nil {
					log.Printf("%+v\n", err)
					ctx = context.WithValue(ctx, "error", err)
					cancel()
					continue
				}

				ct, err := strconv.ParseInt(tweet.CreatedAt, 10, 64)
				if err != nil {
					log.Printf("%+v\n", err)
					ctx = context.WithValue(ctx, "error", err)
					cancel()
					continue
				}

				// そのツイートの投稿時間が24時間以上経過しているかどうか
				t := time.Unix(ct, 0)
				// 1日経っていないならbreak
				if !time.Now().After(t.AddDate(0, 0, 1)) {
					break
				}

				log.Printf("deleting... %+v: %+v\n", tweet.ID, tweet.Text)
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

func (bs *batchService) runTask() {

}
