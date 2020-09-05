package api

import (
	"github.com/labstack/echo"
	"net/http"
	"unfire/model"
)

func Health() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK,model.NewResponse(http.StatusOK,"alive","ok"))
	}
}
