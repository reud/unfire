package usecase

import (
	"context"
	"time"
	"unfire/domain/model"
	"unfire/utils"

	"github.com/garyburd/go-oauth/oauth"
)

// errorを返さない様なメソッドは基本panicやFatalfを投げる事
type DatastoreController interface {
	StoreAllTweet(ctx context.Context, twitterID string, tweets []model.Tweet)
	InsertTweetToTimeLine(ctx context.Context, twitterID string, tweet model.Tweet)
	StoreAuthorizeData(ctx context.Context, twitterID string, cred *oauth.Credentials)
	PickAuthorizeData(ctx context.Context, twitterID string) *oauth.Credentials
	AppendToUsers(ctx context.Context, twitterID string)
	SetUserStatus(ctx context.Context, twitterID string, status utils.UserStatus)
	// GetOldestTweetInfoFromTimeLine: returns <createdAt, userID> error(empty時などに利用)
	GetOldestTweetInfoFromTimeLine(ctx context.Context) (time.Time, string, error)
	PopOldestTweetInfoFromTimeLine(ctx context.Context)
	// GetUserLastTweet: ユーザの一番古いツイートIDを取得する。 returns <userID, status(if false, maybe user tweets empty)>
	GetUserLastTweet(ctx context.Context, twitterID string) (string, bool)
}
