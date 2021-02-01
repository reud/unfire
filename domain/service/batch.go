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
			// 保存されているツイートの中で最も古いものを取得する。
			data, err := ds.GetMinElement(ctx, utils.TimeLine)
			if err != nil {
				log.Printf("GetMinElement Error: %+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			sp := strings.Split(data, "_")
			if len(sp) != 2 {
				err := errors.New(fmt.Sprintf("bad data got: %+v", sp))
				log.Printf("strings.Split Error: %+v\n", err)
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

			// 取得したuserIDのaccess tokenを取り出す。
			atStr, err := ds.GetHash(ctx, utils.TokenSuffix+userID, "at")
			if err != nil {
				log.Printf("pick at error: : %+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// 取得したuserIDのsecret tokenを取り出す
			secStr, err := ds.GetHash(ctx, utils.TokenSuffix+userID, "sec")
			if err != nil {
				log.Printf("pick secret token error:  %+v\n", err)
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			// 認可情報を作成
			cred := &oauth.Credentials{
				Token:  atStr,
				Secret: secStr,
			}

			log.Printf("created token:%+v secret:%+v\n", cred.Token, cred.Secret)
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
				select {
				case <-ctx.Done():
					break
				default:
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
					log.Printf("tweet fetch success  %+v: %+v\n", tweet.ID, tweet.Text)

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

}

func (bs *batchService) runTask() {

}
