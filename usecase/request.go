package usecase

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type RequestContext interface {
	Bind(interface{}) error
	Request() *http.Request
	Response() *echo.Response // TODO: (抽象化)
	Validate(interface{}) error
}
