package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unfire/domain/model"
	"unfire/infrastructure/datastore"
	"unfire/usecase"
	"unfire/utils"

	"github.com/pkg/errors"

	"github.com/garyburd/go-oauth/oauth"
)

type DatastoreController struct {
	ds datastore.Datastore
}

// TODO: 失敗したらどうする？
func (dc *DatastoreController) StoreAllTweet(ctx context.Context, twitterID string, tweets []model.Tweet) {
	for _, v := range tweets {
		if err := dc.ds.AppendString(ctx, twitterID+utils.TweetsSuffix, v.ID); err != nil {
			log.Fatalf("redis error. tweet append failed: %+v", err)
		}
	}
}

func (dc *DatastoreController) InsertTweetToTimeLine(ctx context.Context, twitterID string, tweet model.Tweet) {
	createdAt, err := time.Parse("2006-01-02T15:04:05.000Z", tweet.CreatedAt)

	if err != nil {
		fmt.Printf("time parse failed(originalCreatedAT -> time.Date) from: %+v", tweet.CreatedAt)
		// API変更などしか考えられないためpanicさせる(正常動作を継続させない方が良い)
		log.Fatal("fatal error failed to parsing time...")
	}

	idi64 := createdAt.Unix()

	// 一番古いツイートの作成時間(unixtime)とそのツイートの保持者を格納する。
	if err := dc.ds.Insert(ctx, utils.TimeLine, float64(idi64-utils.TimeLinePrefix), strconv.FormatInt(idi64, 10)+"_"+twitterID); err != nil {
		fmt.Printf("failed to insert timeline: %+v", err)
		// これに失敗する場合は考えられないのでpanicを行う。redis回りでコケているならば、そもそも停止させるべきである。
		log.Fatal("failed to insert to timeline...")
	}
}

// セキュアなデータを保存する。
func (dc *DatastoreController) StoreAuthorizeData(ctx context.Context, twitterID string, cred *oauth.Credentials) {
	// token周りの情報を保存(at)
	if err := dc.ds.SetHash(ctx, utils.TokenSuffix+twitterID, "at", cred.Token); err != nil {
		log.Fatalf("failed to save at.Token: %+v", err)
	}

	// token周りの情報を保存(sec)
	if err := dc.ds.SetHash(ctx, utils.TokenSuffix+twitterID, "sec", cred.Secret); err != nil {
		log.Fatalf("failed to save at.Sec: %+v", err)
	}
}

// セキュアなデータを保存する。
func (dc *DatastoreController) PickAuthorizeData(ctx context.Context, twitterID string) *oauth.Credentials {
	// 取得したuserIDのaccess tokenを取り出す。
	atStr, err := dc.ds.GetHash(ctx, utils.TokenSuffix+twitterID, "at")
	if err != nil {
		log.Fatalf("pick at error: : %+v\n", err)
	}

	// 取得したuserIDのsecret tokenを取り出す
	secStr, err := dc.ds.GetHash(ctx, utils.TokenSuffix+twitterID, "sec")
	if err != nil {
		log.Printf("pick secret token error:  %+v\n", err)
	}

	// 認可情報を作成
	return &oauth.Credentials{
		Token:  atStr,
		Secret: secStr,
	}
}

func (dc *DatastoreController) AppendToUsers(ctx context.Context, twitterID string) {
	// user一覧情報に保存
	if err := dc.ds.AppendString(ctx, utils.Users, twitterID); err != nil {
		log.Fatalf("failed to set add user: %+v", err)
	}
}

func (dc *DatastoreController) SetUserStatus(ctx context.Context, twitterID string, status utils.UserStatus) {
	if err := dc.ds.SetString(ctx, twitterID+utils.StatusSuffix, status.String()); err != nil {
		log.Fatalf("failed to set user status: %+v", err)
	}
}
func (dc *DatastoreController) GetOldestTweetInfoFromTimeLine(ctx context.Context) (time.Time, string, error) {
	// 保存されているツイートの中で最も古いものを取得する。
	data, err := dc.ds.GetMinElement(ctx, utils.TimeLine)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("failed to pick element(GetOldestTweetInfoFromTimeLine): %+v", err)
	}

	sp := strings.Split(data, "_")
	if len(sp) != 2 {
		err := errors.New(fmt.Sprintf("bad data got: %+v", sp))
		log.Fatalf("strings.Split Error: %+v\n", err)
	}

	tweetTimei64, err := strconv.ParseInt(sp[0], 10, 64)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	t := time.Unix(tweetTimei64, 0)
	userID := sp[1]

	return t, userID, nil
}

func (dc *DatastoreController) PopOldestTweetInfoFromTimeLine(ctx context.Context) {

}

func (dc *DatastoreController) GetUserLastTweet(ctx context.Context, twitterID string) (string, bool) {
	lastTweetID, err := dc.ds.LastPop(ctx, twitterID+utils.TweetsSuffix)
	if err != nil {
		log.Printf("GetUserLastTweetErr: %+v\n", err)
		return "", false
	}

	return lastTweetID, true
}

func NewDatastoreController(ds datastore.Datastore) usecase.DatastoreController {
	return &DatastoreController{ds: ds}
}
