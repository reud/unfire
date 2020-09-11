package api

import (
	"github.com/garyburd/go-oauth/oauth"
	"github.com/labstack/echo/v4"
	"net/http"
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
	callbackURL = "https://unfire.herokuapp.com/api/v1/auth/callback"
)

func pickAccessToken(c *echo.Context) (string, string, bool, error) {
	atmn, err := session.NewManager("at", c)
	if err != nil {
		return "", "", false, err
	}
	t, ok := atmn.Get("token")
	if !ok {
		return "", "", false, nil
	}
	s, ok := atmn.Get("secret")
	if !ok {
		return "", "", false, nil
	}
	return t.(string), s.(string), true, nil
}

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

		mn.Set("token", rt.Token)
		mn.Set("secret", rt.Secret)

		err = mn.Save(c.Request(), c.Response())
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

		err = mn.Clear(c.Request(), c.Response())
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to delete session", err))
		}

		if err := tunnel.AddUserByCredentials(at.Token, at.Secret); err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to add user", err))
		}
		return c.JSON(http.StatusOK, model.NewResponse(http.StatusOK, "ok", account))
	}

}
