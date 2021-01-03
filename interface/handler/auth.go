package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"unfire/domain/service"
	repository2 "unfire/infrastructure/repository"
	"unfire/model"
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
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to login(session error)", err))
		}
		redirect, err := usecase.Login(c, sr, as)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to login(usecase error)", err))
		}
		return c.Redirect(http.StatusMovedPermanently, redirect)
	}
}

func (ah *authHandler) GetCallback(usecase usecase.AuthUseCase, as service.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := repository2.NewSessionRepository("request", &c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to login", err))
		}
		redirect, err := usecase.Callback(c, sr, as)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.NewResponse(http.StatusBadRequest, "failed to callback", err))
		}
		return c.Redirect(http.StatusMovedPermanently, redirect)
	}
}
