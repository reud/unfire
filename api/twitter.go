package api

import (
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
	"unfire/client"
	"unfire/model"
)

type TwitterCallBackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

const (
	callbackURL = "http://127.0.0.1:8080/api/v1/auth/callback"
)

func LoginByTwitter() echo.HandlerFunc {
	return func(c echo.Context) error {
		oc := client.NewTWClient()
		rt, err := oc.RequestTemporaryCredentials(nil, callbackURL, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		sess, err := session.Get("session", c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting sessio　in planting req,reqt", err))
		}
		sess.Values["request_token"] = rt.Token
		sess.Values["request_token_secret"] = rt.Secret

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to write session", err))
		}

		url := oc.AuthorizationURL(rt, nil)

		return c.Redirect(http.StatusMovedPermanently, url)
	}
}

func TwitterCallback() echo.HandlerFunc {
	return func(c echo.Context) error {
		q := new(TwitterCallBackQuery)
		if err := c.Bind(q); err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to read callback", err))
		}
		sess, err := session.Get("session", c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting session", err))
		}
		reqt, ok := sess.Values["request_token"].(string)
		if !ok {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting session value (request_token)", reqt))
		}
		if reqt != q.OAuthToken {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error at request_token != oauth_token", reqt))
		}
		reqts, ok := sess.Values["request_token_secret"].(string)
		if !ok {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "error in getting session value (request_token_secret)", reqts))
		}
		code, at, err := client.GetAccessToken(&oauth.Credentials{
			Token:  reqt,
			Secret: reqts,
		}, q.OAuthVerifier)
		if err != nil {
			return c.JSON(code, model.NewResponse(code, "error in getting access tokrn", err))
		}
		account := struct {
			ID         string `json:"id_str"`
			ScreenName string `json:"screen_name"`
		}{}
		code, err = client.GetMe(at, &account)
		if err != nil {
			return c.JSON(code, nil)
		}

		fmt.Println(account)

		return c.JSON(http.StatusOK, model.NewResponse(http.StatusOK, "ok", account))
	}

}
