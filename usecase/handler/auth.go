package handler

import (
	"context"
	"fmt"
	"log"
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
	Callback(ctx usecase.RequestContext, mn repository.SessionRepository, authService service.AuthService, tci client.TwitterClientInitializer, dc usecase.DatastoreController) (string, error)
	Stop(ctx usecase.RequestContext, mn repository.SessionRepository) (string, error)
}

type authUseCase struct {
}

type TwitterCallbackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{}
}

// Login: 次のURLとerrorを返す。
func (au *authUseCase) Login(ctx usecase.RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error) {
	// Sessionが存在する場合は削除する。

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

// Callback: 次のurlかerrorを返す。
func (au *authUseCase) Callback(ctx usecase.RequestContext, mn repository.SessionRepository, as service.AuthService, tci client.TwitterClientInitializer, dc usecase.DatastoreController) (string, error) {
	q := new(TwitterCallbackQuery)
	if err := ctx.Bind(q); err != nil {
		return "", err
	}

	reqt, ok := mn.Get("token")
	if !ok {
		return "", errors.New("error in getting session value (request_token)")
	}

	if reqt != q.OAuthToken {
		return "", fmt.Errorf("error at request_token != oauth_token, reqt: %+v, oautht: %+v ", reqt, q.OAuthToken)
	}

	reqts, ok := mn.Get("secret")
	if !ok {
		return "", errors.New("error in getting session value (request_token_secret)")
	}

	at, err := as.GetAccessToken(&oauth.Credentials{
		Token:  reqt.(string),
		Secret: reqts.(string),
	}, q.OAuthVerifier)

	if err != nil {
		return "", err
	}

	tc, err := tci.NewTwitterClient(at)
	if err != nil {
		return "", err
	}

	userID := tc.FetchMe().ID

	// ユーザ一覧情報に保存
	dc.AppendToUsers(ctx.Request().Context(), userID)

	// 認証情報を保存する
	dc.StoreAuthorizeData(ctx.Request().Context(), userID, at)

	// ユーザを初期化中に変更
	dc.SetUserStatus(ctx.Request().Context(), userID, utils.Initializing)

	// twitterIDをセッションに保存
	mn.Set("twitter_id", userID)
	if err := mn.Save(ctx.Request(), &ctx.Response().Writer); err != nil {
		log.Printf("failed to save session... err: %+v", err)
	}

	// ツイートの全ロードを行い、各種datastoreに格納を行う
	go func(ctx context.Context) {

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

	}(context.Background())

	return "https://portal.reud.net/", nil
}

func (au *authUseCase) Stop(ctx usecase.RequestContext, mn repository.SessionRepository) (string, error) {
	id, ok := mn.Get("twitter_id")
	if !ok {
		return "", errors.New("failed to fetch twitter_id")
	}

	ds, err := datastore.NewRedisDatastore()
	if err != nil {
		return "", err
	}

	dc := service.NewDatastoreController(ds)

	nowStatus := dc.GetUserStatus(ctx.Request().Context(), id.(string))

	if nowStatus == utils.Deleted {
		return "user already deleted", errors.New("user already deleted")
	}

	dc.SetUserStatus(ctx.Request().Context(), id.(string), utils.Deleted)
	dc.DeleteUserFromUsersTable(ctx.Request().Context(), id.(string))

	return "ok, change status success id :" + id.(string), nil
}
