package handler

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	client2 "unfire/domain/client"
	"unfire/domain/repository"
	"unfire/domain/service"
	"unfire/infrastructure/client"
	"unfire/infrastructure/datastore"
	"unfire/usecase"
	"unfire/utils"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/pkg/errors"
)

type AuthUseCase interface {
	Login(ctx usecase.RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error)
	Callback(ctx usecase.RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error)
}

type authUseCase struct {
}

type GetLoginParameter struct {
	DeleteLike                 bool   `query:"delete_like"`
	DeleteLikeCount            int    `query:"delete_like_count" validate:"min=1,max=1000"`
	KeepLegendaryTweetV1Enable bool   `query:"keep_legendary_tweet_v1_enable"`
	KeepLegendaryTweetV1Border int    `query:"keep_legendary_tweet_v1_border" validate:"min=15,max=10000000"`
	CallbackUrl                string `query:"callback_url" validate:"omitempty,url_encoded"`
}

type Option struct {
	DeleteLike                 bool   `query:"delete_like"`
	DeleteLikeCount            int    `query:"delete_like_count" validate:"min=1,max=1000"`
	KeepLegendaryTweetV1Enable bool   `query:"keep_legendary_tweet_v1_enable"`
	KeepLegendaryTweetV1Border int    `query:"keep_legendary_tweet_v1_border" validate:"min=15,max=10000000"`
	CallbackUrl                string `query:"callback_url" validate:"omitempty,url_encoded"`
}

type TwitterCallbackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

func newGetLoginParameter() GetLoginParameter {
	// defaults
	return GetLoginParameter{
		DeleteLike:                 false,
		DeleteLikeCount:            30,
		KeepLegendaryTweetV1Enable: false,
		KeepLegendaryTweetV1Border: 20000,
		CallbackUrl:                "",
	}
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{}
}

func isnil(x interface{}) bool {
	return (x == nil) || reflect.ValueOf(x).IsNil()
}

// Login: 次のURLとerrorを返す。
func (au *authUseCase) Login(ctx usecase.RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error) {

	// パラメータのバインド
	ps := newGetLoginParameter()
	if err := ctx.Bind(&ps); err != nil {
		return "", err
	}

	// TODO: Bind機能を実装する。以下のコードは{}のerrorが変えるので調査が必要
	/*
		if err := ctx.Validate(&ps); err != nil {
			return "", err
		}
	*/

	if ps.DeleteLike {
		mn.Set("delete_like_count", strconv.Itoa(ps.DeleteLikeCount))
		mn.Set("delete_like", "true")
	}

	if ps.KeepLegendaryTweetV1Enable {
		mn.Set("keep_legendary_tweet_v1_border", strconv.Itoa(ps.KeepLegendaryTweetV1Border))
		mn.Set("keep_legendary_tweet_v1_enable", "true")
	}

	if ps.CallbackUrl != "" {
		mn.Set("callback_url", ps.CallbackUrl)
	}

	rt, u, err := authService.RequestTemporaryCredentialsAuthorizationURL()
	if err != nil {
		return "", err
	}

	mn.Set("token", rt.Token)
	mn.Set("secret", rt.Secret)

	if err := mn.Save(ctx.Request(), &ctx.Response().Writer); err != nil {
		return "", errors.New("failed to save session")
	}
	return u, nil
}

// TODO: (これもしやecho.Contextじゃなくていい感じの引数にすればライブラリ非依存でテスト出来てめっちゃハッピーになるのでは？)
// Callback: 次のurlかerrorを返す。
func (au *authUseCase) Callback(ctx usecase.RequestContext, mn repository.SessionRepository, as service.AuthService) (string, error) {
	q := new(TwitterCallbackQuery)
	if err := ctx.Bind(q); err != nil {
		return "", err
	}

	fmt.Printf("%+v\n", q)

	reqt, ok := mn.Get("token")
	if !ok {
		return "", errors.New("error in getting session value (request_token)")
	}

	if reqt != q.OAuthToken {
		return "", errors.New("error at request_token != oauth_token")
	}

	reqts, ok := mn.Get("secret")
	if !ok {
		return "", errors.New("error in getting session value (request_token_secret)")
	}

	fmt.Printf("reqt: %+v reqts: %+v\n", reqt, reqts)
	at, err := as.GetAccessToken(&oauth.Credentials{
		Token:  reqt.(string),
		Secret: reqts.(string),
	}, q.OAuthVerifier)

	fmt.Printf("got accesstoken\n")
	if err != nil {
		return "", err
	}

	fmt.Printf("got new client\n")
	tc, err := client.NewTwitterClient(at)
	if err != nil {
		return "", err
	}

	op, err := getOptions(mn)

	err = mn.Clear(ctx.Request(), &ctx.Response().Writer)
	if err != nil {
		return "", err
	}
	fmt.Printf("session cleared\n")

	ds, err := datastore.NewRedisDatastore()
	if err != nil {
		return "", err
	}

	dc := service.NewDatastoreController(ds)
	userID := tc.FetchMe().ID

	// ユーザ一覧情報に保存
	dc.AppendToUsers(ctx.Request().Context(), userID)

	// 認証情報を保存する
	dc.StoreAuthorizeData(ctx.Request().Context(), userID, at)

	// ユーザを初期化中に変更
	dc.SetUserStatus(ctx.Request().Context(), userID, utils.Initializing)

	// ツイートの全ロードを行い、各種datastoreに格納を行う
	go func(ctx context.Context) {
		log.Printf("goroutine start \n")

		tweets, err := tc.FetchTweets(client2.GetAll())

		if err != nil {
			fmt.Printf("tweet fetch error.  failed: %+v", err)
			return
		}

		// ツイートの保存
		dc.StoreAllTweet(ctx, userID, tweets)

		// ツイートが入っていない場合はWaitingステータスに変更する。
		if len(tweets) == 0 {
			dc.SetUserStatus(ctx, userID, utils.Waiting)
			return
		}

		// ツイートが入っている場合はツイートの中で一番古い時間をタイムラインに格納する
		dc.InsertTweetToTimeLine(ctx, userID, tweets[len(tweets)-1])

		// タスク終了後はユーザのステータスをワーキングにする。
		dc.SetUserStatus(ctx, userID, utils.Working)

		fmt.Printf("goroutine finished\n")
	}(context.Background())

	fmt.Printf("callback request finished\n")
	return op.CallbackUrl, nil
}

func getOptions(mn repository.SessionRepository) (*Option, error) {
	op := &Option{
		DeleteLike:                 false,
		DeleteLikeCount:            0,
		KeepLegendaryTweetV1Enable: false,
		KeepLegendaryTweetV1Border: 0,
		CallbackUrl:                "",
	}

	deleteLike, ok := mn.Get("delete_like")
	if ok && deleteLike == "true" {
		op.DeleteLike = true
		cntStr, ok := mn.Get("delete_like_count")
		if !ok {
			return nil, errors.New("error in getting session value (delete_like_count)")
		}
		cnt, err := strconv.Atoi(cntStr.(string))
		if err != nil {
			return nil, err
		}
		op.DeleteLikeCount = cnt
	}

	keepLegendaryTweetV1, ok := mn.Get("keep_legendary_tweet_v1_enable")
	if ok && keepLegendaryTweetV1 == "true" {
		op.KeepLegendaryTweetV1Enable = true
		cntStr, ok := mn.Get("keep_legendary_tweet_v1_border")
		if !ok {
			return nil, errors.New("error in getting session value (keep_legendary_tweet_v1_border)")
		}
		cnt, err := strconv.Atoi(cntStr.(string))
		if err != nil {
			return nil, err
		}
		op.KeepLegendaryTweetV1Border = cnt
	}

	callbackURL, ok := mn.Get("callback_url")
	if ok && callbackURL != "" {
		op.CallbackUrl = callbackURL.(string)
	}
	return op, nil
}
