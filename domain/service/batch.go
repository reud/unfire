package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
	"unfire/infrastructure/persistence"
	"unfire/utils"
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
			fmt.Printf("batch started: %+v", t)

		}
	}()
}

func fetchOldestTweet(ds persistence.Datastore) error {
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
			// 最小値を持ってくる
			data, err := ds.GetMinElement(ctx, utils.TimeLine)
			if err != nil {
				ctx = context.WithValue(ctx, "error", err)
				cancel()
				continue
			}

			sp := strings.Split(data, "_")
			if len(data) != 2 {
				ctx = context.WithValue(ctx, "error", errors.New(fmt.Sprintf("bad data got: %+v", sp)))
				cancel()
				continue
			}

			tweetTime, err := strconv.ParseInt(sp[0], 10, 64)
			if err != nil {
				ctx = context.WithValue(ctx, "error", errors.Wrap(err, "tweetTime parse failed"))
				cancel()
				continue
			}

			userID := sp[1]
			if err != nil {
				ctx = context.WithValue(ctx, "error", errors.Wrap(err, "userID parse failed"))
				cancel()
				continue
			}

			// その最小値が24時間以上経過しているかどうか。
			t := time.Unix(tweetTime, 0)
			// 経過していない場合は終了
			if !time.Now().After(t.AddDate(0, 0, 1)) {
				cancel()
				continue
			}

			// そのuserIDのtweetを後ろから取る。
			for {
				lastTweetID, err := ds.LastPop(ctx, userID+utils.TweetsSuffix)
				if err != nil {
					ctx = context.WithValue(ctx, "error", errors.Wrap(err, "fetch last tweet failed"))
					cancel()
					continue
				}
				// TODO: lastTweetIDからtweetを取得する。

			}

		}

	}

}

func (bs *batchService) runTask() {

}
