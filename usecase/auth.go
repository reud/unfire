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

const (
	tweetsSuffix = "_tweets"
	tweetPrefix  = "tweet_"
)

// Login: 次のURLとerrorを返す。
func (au *authUseCase) Login(ctx RequestContext, mn repository.SessionRepository, authService service.AuthService) (string, error) {

	// パラメータのバインド
	ps := newGetLoginParameter()
	if err := ctx.Bind(&ps); err != nil {
		return "", err
	}

	if err := ctx.Validate(&ps); err != nil {
		return "", err
	}

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

	return u, nil
}

// TODO: (これもしやecho.Contextじゃなくていい感じの引数にすればライブラリ非依存でテスト出来てめっちゃハッピーになるのでは？)
// Callback: 次のurlかerrorを返す。
func (au *authUseCase) Callback(ctx RequestContext, mn repository.SessionRepository, as service.AuthService) (string, error) {
	q := new(TwitterCallbackQuery)
	if err := ctx.Bind(q); err != nil {
		return "", err
	}

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

	at, err := as.GetAccessToken(&oauth.Credentials{
		Token:  reqt.(string),
		Secret: reqts.(string),
	}, q.OAuthVerifier)

	if err != nil {
		return "", err
	}

	tc, err := client.NewTwitterClient(at)
	if err != nil {
		return "", err
	}

	op, err := getOptions(mn)

	err = mn.Clear(ctx.Request(), ctx.Response())
	if err != nil {
		return "", errors.Wrap(err, "failed to clear session")
	}

	ds, err := persistence.NewRedisDatastore()
	if err != nil {
		return "", errors.Wrap(err, "redis init failed")
	}

	if err := ds.SetString(ctx.Request().Context(), tc.FetchMe().ID+utils.StatusSuffix, utils.Initializing.String()); err != nil {
		return "", errors.Wrap(err, "redis set status failed")
	}

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
			if err := ds.AppendString(ctx, tc.FetchMe().ID+tweetsSuffix, v.IDStr); err != nil {
				fmt.Printf("redis error. tweet append failed: %+v", err)
				return
			}
		}

		lnth, err := ds.ListLen(ctx, tc.FetchMe().ID+tweetsSuffix)
		if err != nil {
			fmt.Printf("redis error. tweet read(ListLen): %+v", err)
			return
		}
		// ツイートが入っていれば、一番古いツイートを格納する。
		if lnth != 0 {
			oldest, err := ds.GetStringByIndex(ctx, tc.FetchMe().ID+tweetsSuffix, lnth-1)
			if err != nil {
				fmt.Printf("redis error (GetStringByIndex) %+v", err)

			}
			if err := ds.SetString(ctx, tweetPrefix+oldest, tc.FetchMe().ID); err != nil {
				fmt.Printf("redis error (SetString) %+v", err)
				return
			}
			idi64, err := strconv.ParseInt(oldest, 10, 64)
			if err != nil {
				fmt.Printf("tweet idstr parse failed: %+v  original: %+v", err, oldest)
				return
			}
			if err := ds.InsertInt64(ctx, utils.TimeLine, idi64); err != nil {
				fmt.Printf("failed to insert timeline: %+v", err)
				return
			}
		}

		if err := ds.SetString(ctx, tc.FetchMe().ID+utils.StatusSuffix, utils.Working.String()); err != nil {
			fmt.Printf("failed to set status timeline: %+v", err)
			return
		}

	}(ctx.Request().Context())

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
