package repository

import (
	"net/http"
)

type SessionRepository interface {
	Get(key string) (interface{}, bool)
	Set(key string, value string)
	Save(req *http.Request, res *http.ResponseWriter) error
	Clear(req *http.Request, res *http.ResponseWriter) error
}
