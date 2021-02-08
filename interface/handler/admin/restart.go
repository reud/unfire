package admin

import (
	"github.com/labstack/echo/v4"
	"net/http"
	config2 "unfire/config"
	"unfire/usecase/handler/admin"
)

type RestartReloadQuery struct {
	KeyPhrase string `query:"key_phrase"`
}
type RestartDeleteQuery struct {
	KeyPhrase string `query:"key_phrase"`
}

type RestartHandler interface {
	GetReload(usecase admin.RestartUseCase) echo.HandlerFunc
	GetDelete(usecase admin.RestartUseCase) echo.HandlerFunc
}

type restartHandlerImpl struct{}

func NewRestartHandler() RestartHandler {
	return &restartHandlerImpl{}
}

func (rh *restartHandlerImpl) GetReload(usecase admin.RestartUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := &RestartReloadQuery{}
		if err := c.Bind(q); err != nil {
			return c.JSON(http.StatusBadRequest, "failed to parse key_phrase")
		}
		cfg := config2.GetInstance()

		if q.KeyPhrase != cfg.AdminAPIPassword {
			return c.JSON(http.StatusBadRequest, "failed to match key_phrase")
		}

		usecase.Reload(c)
		return c.JSON(http.StatusOK, "ok, reload task stacked")
	}
}

func (rh *restartHandlerImpl) GetDelete(usecase admin.RestartUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := &RestartDeleteQuery{}
		if err := c.Bind(q); err != nil {
			return c.JSON(http.StatusBadRequest, "failed to parse key_phrase")
		}
		cfg := config2.GetInstance()

		if q.KeyPhrase != cfg.AdminAPIPassword {
			return c.JSON(http.StatusBadRequest, "failed to match key_phrase")
		}

		usecase.Delete(c)
		return c.JSON(http.StatusOK, "ok, delete task stacked")
	}
}
