package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"unfire/domain/service"
	repository2 "unfire/infrastructure/repository"
	"unfire/usecase"
)

type TwitterCallBackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

type AuthHandler interface {
	GetLogin(usecase usecase.AuthUseCase, au service.AuthService) echo.HandlerFunc
	GetCallback(usecase usecase.AuthUseCase, as service.AuthService) echo.HandlerFunc
}

type authHandler struct{}

func NewAuthHandler() AuthHandler {
	return &authHandler{}
}

func (ah *authHandler) GetLogin(usecase usecase.AuthUseCase, as service.AuthService) echo.HandlerFunc {
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

func (ah *authHandler) GetCallback(usecase usecase.AuthUseCase, as service.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := repository2.NewSessionRepository("request", &c)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		redirect, err := usecase.Callback(c, sr, as)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.Redirect(http.StatusMovedPermanently, redirect)
	}
}
