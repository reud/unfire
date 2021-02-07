package route

import (
	"unfire/api"
	"unfire/domain/service"
	"unfire/infrastructure/repository"
	"unfire/interface/handler"
	handler2 "unfire/usecase/handler"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

func Init(as service.AuthService, au handler2.AuthUseCase, si repository.SessionInitializer) *echo.Echo {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding},
	}))
	e.Use(echoMw.Logger())

	ah := handler.NewAuthHandler()

	auth := e.Group("/auth")
	{
		auth.GET("/login", ah.GetLogin(au, as, si))
		auth.GET("/callback", ah.GetCallback(au, as, si))
	}

	e.GET("/health", api.Health())

	// routes
	v1 := e.Group("/api/v1")
	{

		v1.GET("/health", api.Health())
	}
	return e
}
