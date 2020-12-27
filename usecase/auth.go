package usecase

import (
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"unfire/client"
	"unfire/domain/repository"
	"unfire/domain/service"
	"unfire/model"
	"unfire/session"
)

type AuthUseCase interface {
	Login(ctx *echo.Context) (string, error)
	Callback(ctx *echo.Context) error
}

type authUseCase struct {
	sessionRepository repository.SessionRepository
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

func NewAuthUseCase(sr repository.SessionRepository) AuthUseCase {
	return &authUseCase{sessionRepository: sr}
}

func isnil(x interface{}) bool {
	return (x == nil) || reflect.ValueOf(x).IsNil()
}

// Login: 次のURLとerrorを返す。
func (au *authUseCase) Login(ctx *echo.Context) (string, error) {
	mn, err := session.NewManager("request", ctx)
	if err != nil {
		return "", nil
	}

	// パラメータのバインド
	ps := newGetLoginParameter()
	if err := (*ctx).Bind(&ps); err != nil {
		return "", err
	}

	if err := (*ctx).Validate(&ps); err != nil {
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

	rt, u, err := service.NewAuthService().RequestTemporaryCredentialsAuthorizationURL()
	if err != nil {
		return "", err
	}

	mn.Set("token", rt.Token)
	mn.Set("secret", rt.Secret)

	return u, nil
}

func (au *authUseCase) Callback(ctx *echo.Context) error {
	q := new(TwitterCallbackQuery)
	if err := (*ctx).Bind(q); err != nil {
		return err
	}

	mn, err := session.NewManager("request", ctx)
	if err != nil {
		return err
	}

	reqt, ok := mn.Get("token")
	if !ok {
		return errors.New("error in getting session value (request_token)")
	}

	if reqt != q.OAuthToken {
		return errors.New("error at request_token != oauth_token")
	}

	reqts, ok := mn.Get("secret")
	if !ok {
		return errors.New("error in getting session value (request_token_secret)")
	}

	_, at, err := client.GetAccessToken(&oauth.Credentials{
		Token:  reqt.(string),
		Secret: reqts.(string),
	}, q.OAuthVerifier)

	if err != nil {
		return err
	}
	account, err := client.GetMe(at)

	return nil
}

func getOptions(mn *session.Manager) (*Option, error) {
	op := &Option{
		DeleteLike:                 false,
		DeleteLikeCount:            0,
		KeepLegendaryTweetV1Enable: false,
		KeepLegendaryTweetV1Border: 0,
		CallbackURL:                "",
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
		op.CallbackURL = callbackURL.(string)
	}
	return op, nil
}
