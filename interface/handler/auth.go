package handler

import (
	"fmt"
	"net/http"
	"unfire/domain/service"
	repository2 "unfire/infrastructure/repository"
	"unfire/usecase/handler"

	"github.com/labstack/echo/v4"
)

type TwitterCallBackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

type AuthHandler interface {
	GetLogin(usecase handler.AuthUseCase, au service.AuthService) echo.HandlerFunc
	GetCallback(usecase handler.AuthUseCase, as service.AuthService) echo.HandlerFunc
}

type authHandler struct{}

func NewAuthHandler() AuthHandler {
	return &authHandler{}
}

func (ah *authHandler) GetLogin(usecase handler.AuthUseCase, as service.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := repository2.NewSessionRepository("request", &c)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		redirect, err := usecase.Login(c, sr, as)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.Redirect(http.StatusMovedPermanently, redirect)
	}
}

func (ah *authHandler) GetCallback(usecase handler.AuthUseCase, as service.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := repository2.NewSessionRepository("request", &c)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		_, err = usecase.Callback(c, sr, as)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		fmt.Printf("callback finished\n")
		return c.Redirect(http.StatusMovedPermanently, "https://portal.reud.net/")
	}
}
