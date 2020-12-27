package repository

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type SessionRepository interface {
	Get(key string) (interface{}, bool)
	Set(key string, value string)
	Save(req *http.Request, res *echo.Response) error
	Clear(req *http.Request, res *echo.Response) error
}
