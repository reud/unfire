package scenario

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"unfire/domain/service"
	"unfire/infrastructure/datastore"
	"unfire/infrastructure/repository"
	"unfire/tests/e2e/scenario/mock"

	"github.com/gorilla/sessions"

	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo/v4"
)

func generateUser() *Scenario {
	return &Scenario{Work: func(t *testing.T, cases UseCases) {
		e := echo.New()

		var cookies []*http.Cookie
		cookieStore := sessions.NewCookieStore([]byte("secret"))

		// login
		{
			req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
			rq := req.URL.Query()
			rq.Set("callback_url", "https://example.com/callback")
			req.URL.RawQuery = rq.Encode()

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.Set("_session_store", cookieStore)

			// spew.Dump(ctx)

			sein := repository.NewSessionInitializer()
			sess, err := sein.NewSessionRepository("request", &ctx)
			if err != nil {
				log.Fatalf("failed to new session repository err: %+v", err)
			}

			mas := mock.NewMockAuthService()

			redirect, err := cases.Au.Login(ctx, sess, mas)
			assert.Equal(t, "http://example.com/authorize", redirect)
			assert.Nil(t, err)

			// ここからCookieを取り出す。
			cookies = rec.Result().Cookies()

			assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		}

		// callback
		{
			req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
			for _, v := range cookies {
				req.AddCookie(v)
			}
			params := req.URL.Query()
			params.Add("oauth_token", "mock token")
			req.URL.RawQuery = params.Encode()

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.Set("_session_store", cookieStore)

			sein := repository.NewSessionInitializer()
			sess, err := sein.NewSessionRepository("request", &ctx)
			if err != nil {
				log.Fatalf("failed to new session repository err: %+v", err)
			}

			mas := mock.NewMockAuthService()
			tc := mock.NewTwitterClientInitializer()
			ds, err := datastore.NewRedisDatastore()
			if err != nil {
				log.Fatal(err)
			}
			dc := service.NewDatastoreController(ds)

			callback, err := cases.Au.Callback(ctx, sess, mas, tc, dc)
			assert.Equal(t, "https://example.com/callback", callback)
			assert.Nil(t, err)
		}

	}}
}
