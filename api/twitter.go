package api

import (
	"errors"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strconv"
	"unfire/client"
	"unfire/model"
	"unfire/session"
	"unfire/tunnel"
)

type TwitterCallBackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

const (
	callbackURL        = "https://unfire.reud.app/api/v1/auth/callback"
	minimumDeleteCount = 1
	maximumDeleteCount = 1000
	minimumBorderCount = 15
	maximumBorderCount = 10000000
)

func LoginByTwitter() echo.HandlerFunc {
	return func(c echo.Context) error {
		oc := client.NewTWClient()
		rt, err := oc.RequestTemporaryCredentials(nil, callbackURL, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		mn, err := session.NewManager("request", &c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting sessio　in planting req,reqt", err))
		}

		if c.QueryParam("delete_like") == "true" {
			deleteLikeCntStr := c.QueryParam("delete_like_count")
			deleteLikeCnt, err := strconv.Atoi(deleteLikeCntStr)
			if err != nil {
				return c.JSON(http.StatusBadGateway, model.NewResponse(http.StatusBadRequest, "failed to read delete_like_count parameter, missing?", err))
			}
			if !(minimumDeleteCount <= deleteLikeCnt && deleteLikeCnt <= maximumDeleteCount) {
				return c.JSON(http.StatusBadGateway, model.NewResponse(http.StatusBadRequest, "delete_like_cnt out of range", err))
			}
			mn.Set("delete_like_count", deleteLikeCntStr)
			mn.Set("delete_like", "true")
		}

		if c.QueryParam("keep_legendary_tweet_v1_enable") == "true" {
			keepLegendaryTweetV1BorderStr := c.QueryParam("keep_legendary_tweet_v1_border")
			keepLegendaryTweetV1Border, err := strconv.Atoi(keepLegendaryTweetV1BorderStr)
			if err != nil {
				return c.JSON(http.StatusBadGateway, model.NewResponse(http.StatusBadRequest, "failed to read delete_like_count parameter, missing?", err))
			}
			if minimumBorderCount <= keepLegendaryTweetV1Border && keepLegendaryTweetV1Border <= maximumBorderCount {
				return c.JSON(http.StatusBadGateway, model.NewResponse(http.StatusBadRequest, "delete_like_cnt out of range", err))
			}
			mn.Set("keep_legendary_tweet_v1_border", keepLegendaryTweetV1BorderStr)
			mn.Set("keep_legendary_tweet_v1_enable", "true")
		}

		if c.QueryParam("callback_url") != "" {
			cb, err := url.Parse(c.QueryParam("callback_url"))
			if err != nil {
				return c.JSON(http.StatusBadGateway, model.NewResponse(http.StatusBadRequest, "invalid callback url", err))
			}
			mn.Set("callback_url", cb.String())
		}

		mn.Set("token", rt.Token)
		mn.Set("secret", rt.Secret)

		err = mn.Save(c.Request(), c.Response())
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to write session", err))
		}

		u := oc.AuthorizationURL(rt, nil)

		return c.Redirect(http.StatusMovedPermanently, u)
	}
}

func TwitterCallback() echo.HandlerFunc {
	return func(c echo.Context) error {
		q := new(TwitterCallBackQuery)
		if err := c.Bind(q); err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to read callback", err))
		}

		mn, err := session.NewManager("request", &c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting session", err))
		}

		reqt, ok := mn.Get("token")
		if !ok {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting session value (request_token)", reqt))
		}
		if reqt != q.OAuthToken {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error at request_token != oauth_token", reqt))
		}
		reqts, ok := mn.Get("secret")
		if !ok {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting session value (request_token_secret)", reqts))
		}

		code, at, err := client.GetAccessToken(&oauth.Credentials{
			Token:  reqt.(string),
			Secret: reqts.(string),
		}, q.OAuthVerifier)
		if err != nil {
			return c.JSON(code, model.NewResponse(code, "error in getting access tokrn", err))
		}

		account, err := client.GetMe(at)
		if err != nil {
			return c.JSON(code, nil)
		}

		op, err := getOptions(mn)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to get options", err))
		}

		err = mn.Clear(c.Request(), c.Response())
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to delete session", err))
		}

		if err := tunnel.AddUserByCredentials(at.Token, at.Secret, *op); err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to add user", err))
		}

		if op.CallbackURL != "" {
			u, err := url.Parse(op.CallbackURL)
			if err != nil {
				return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "callback url corrupted (set: ok get: failed)", err))
			}
			q := u.Query()
			q.Set("status", "ok")
			u.RawQuery = q.Encode()
			return c.Redirect(http.StatusMovedPermanently, u.String())
		}

		return c.JSON(http.StatusOK, model.NewResponse(http.StatusOK, "ok", account))
	}

}

func getOptions(mn *session.Manager) (*model.Options, error) {
	op := &model.Options{
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
