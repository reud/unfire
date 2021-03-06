package handler

import (
	"fmt"
	"net/http"
	"unfire/domain/service"
	client2 "unfire/infrastructure/client"
	repository2 "unfire/infrastructure/repository"
	"unfire/usecase"
	"unfire/usecase/handler"

	"github.com/labstack/echo/v4"
)

type TwitterCallBackQuery struct {
	OAuthToken    string `query:"oauth_token"`
	OAuthVerifier string `query:"oauth_verifier"`
}

type AuthHandler interface {
	GetLogin(usecase handler.AuthUseCase, au service.AuthService, si repository2.SessionInitializer) echo.HandlerFunc
	GetCallback(usecase handler.AuthUseCase, as service.AuthService, si repository2.SessionInitializer, tc client2.TwitterClientInitializer, dc usecase.DatastoreController) echo.HandlerFunc
	GetStop(usecase handler.AuthUseCase, si repository2.SessionInitializer) echo.HandlerFunc
}

type authHandler struct{}

func NewAuthHandler() AuthHandler {
	return &authHandler{}
}

func (ah *authHandler) GetLogin(usecase handler.AuthUseCase, as service.AuthService, si repository2.SessionInitializer) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := si.NewSessionRepository("request", &c)
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

func (ah *authHandler) GetCallback(usecase handler.AuthUseCase, as service.AuthService, si repository2.SessionInitializer, tc client2.TwitterClientInitializer, dc usecase.DatastoreController) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := si.NewSessionRepository("request", &c)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		_, err = usecase.Callback(c, sr, as, tc, dc)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		fmt.Printf("callback finished\n")
		return c.Redirect(http.StatusMovedPermanently, "https://portal.reud.net/")
	}
}

func (ah *authHandler) GetStop(usecase handler.AuthUseCase, si repository2.SessionInitializer) echo.HandlerFunc {
	return func(c echo.Context) error {
		sr, err := si.NewSessionRepository("request", &c)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		result, err := usecase.Stop(c, sr)
		if err != nil {
			fmt.Printf("err!: %+v", err)
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, result)
	}
}
