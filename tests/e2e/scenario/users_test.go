package scenario

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"unfire/infrastructure/repository"
	"unfire/tests/e2e/scenario/mock"

	"github.com/davecgh/go-spew/spew"

	"github.com/gorilla/sessions"

	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo/v4"
)

func generateUser() *Scenario {
	return &Scenario{Work: func(t *testing.T, cases UseCases) {
		e := echo.New()
		{
			req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.Set("_session_store", sessions.NewCookieStore([]byte("secret")))
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

			// TODO: ここからCookieを取り出す。
			// TODO: 以後のリクエストではそのCookieの内容をリクエストに付加して送る
			spew.Dump(rec.Result())
			if rec.Result().StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(rec.Result().Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				log.Printf(bodyString)
			}
		}

	}}
}

func parseCookies(value string) map[string]*http.Cookie {
	m := map[string]*http.Cookie{}
	for _, c := range (&http.Request{Header: http.Header{"Cookie": {value}}}).Cookies() {
		m[c.Name] = c
	}
	return m
}

func GenerateNormalUsers(n int) []*Scenario {
	var users []*Scenario
	cnt := 0
	for cnt < n {
		users = append(users, generateUser())
		cnt++
	}
	return users
}
