package usecase

import (
	"context"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"time"
	client2 "unfire/domain/client"
	"unfire/domain/model"
	"unfire/domain/repository"
	"unfire/domain/service"
	"unfire/infrastructure/client"
	"unfire/infrastructure/persistence"
	"unfire/utils"
)

type AuthUseCase interface {
	Login(ctx RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error)
	Callback(ctx RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error)
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
func (au *authUseCase) Login(ctx RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error) {

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
func (au *authUseCase) Callback(ctx RequestContext, mn repository.SessionRepository, as service.AuthService) (string, error) {
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

	ds, err := persistence.NewRedisDatastore()
	if err != nil {
		return "", err
	}

	if err := ds.SetString(ctx.Request().Context(), tc.FetchMe().ID+utils.StatusSuffix, utils.Initializing.String()); err != nil {
		return "", err
	}

	fmt.Printf("goroutine start \n")
	// ツイートの全ロードを行い、各種datastoreに格納を行う
	go func(ctx context.Context) {
		latestTweets, err := tc.FetchTweets()
		// loggerを使う
		if err != nil {
			fmt.Printf("err in goroutine (tweet fetching): %+v", err)
			return
		}

		var tweets []model.Tweet
		for _, v := range latestTweets {
			tweets = append(tweets, v)
		}

		for len(latestTweets) != 0 {
			fmt.Printf("tweet leading...\n")
			lastID := latestTweets[len(latestTweets)-1].IDStr
			// TODO: API Limit回避の方法について考える。
			time.Sleep(time.Second * 30)
			latestTweets, err = tc.FetchTweets(client2.SinceId(lastID))
			if err != nil {
				fmt.Printf("err in goroutine (tweet fetching): %+v", err)
				return
			}
			for _, v := range latestTweets {
				tweets = append(tweets, v)
			}
		}

		for _, v := range tweets {
			if err := ds.AppendString(ctx, tc.FetchMe().ID+utils.TweetsSuffix, v.IDStr); err != nil {
				fmt.Printf("redis error. tweet append failed: %+v", err)
				return
			}
		}

		lnth, err := ds.ListLen(ctx, tc.FetchMe().ID+utils.TweetsSuffix)
		if err != nil {
			fmt.Printf("redis error. tweet read(ListLen): %+v", err)
			return
		}
		// ツイートが入っていれば、一番古い時間を格納する。(多分一番最後)
		if lnth != 0 {

			idi64, err := strconv.ParseInt(tweets[len(tweets)-1].CreatedAt, 10, 64)
			if err != nil {
				fmt.Printf("tweet idstr parse failed: %+v  original: %+v", err, tweets[len(tweets)-1].CreatedAt)
				return
			}

			if err := ds.Insert(ctx, utils.TimeLine, float64(idi64-utils.TimeLinePrefix), tweets[len(tweets)-1].CreatedAt+"_"+tweets[len(tweets)-1].IDStr); err != nil {
				fmt.Printf("failed to insert timeline: %+v", err)
				return
			}
		}

		// token周りの情報を保存(at)
		if err := ds.SetHash(ctx, utils.TokenSuffix+tc.FetchMe().ID, "at", at.Token); err != nil {
			fmt.Printf("failed to save at.Token: %+v", err)
			return
		}

		// token周りの情報を保存(sec)
		if err := ds.SetHash(ctx, utils.TokenSuffix+tc.FetchMe().ID, "sec", at.Secret); err != nil {
			fmt.Printf("failed to save at.Sec: %+v", err)
			return
		}

		// user一覧情報に保存
		if err := ds.AppendString(ctx, utils.Users, tc.FetchMe().ID); err != nil {
			fmt.Printf("failed to set add user: %+v", err)
			return
		}

		if err := ds.SetString(ctx, tc.FetchMe().ID+utils.StatusSuffix, utils.Working.String()); err != nil {
			fmt.Printf("failed to set status timeline: %+v", err)
			return
		}

	}(ctx.Request().Context())
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
