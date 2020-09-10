package session

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Manager struct {
	session *sessions.Session
	key     string
}

func NewManager(key string, c *echo.Context) (*Manager, error) {
	sess, err := session.Get(key, *c)
	if err != nil {
		return nil, err
	}

	return &Manager{session: sess, key: key}, nil
}

func (sm *Manager) Get(key string) (interface{}, bool) {
	val, ok := sm.session.Values[key]
	return val, ok
}

func (sm *Manager) Set(key string, value string) {
	sm.session.Values[key] = value
}

func (sm *Manager) Save(req *http.Request, res *echo.Response) error {
	return sm.session.Save(req, res)
}

func (sm *Manager) Clear(req *http.Request, res *echo.Response) error {
	sm.session.Options = &sessions.Options{MaxAge: -1, Path: "/"}
	return sm.Save(req, res)
}
