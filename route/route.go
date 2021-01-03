package route

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
	"unfire/api"
	"unfire/domain/service"
	"unfire/interface/handler"
	"unfire/usecase"
)

func Init(as service.AuthService, au usecase.AuthUseCase) *echo.Echo {
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
		auth.GET("/login", ah.GetLogin(au, as))
		auth.GET("/callback", ah.GetCallback(au, as))
	}

	e.GET("/health", api.Health())

	// routes
	v1 := e.Group("/api/v1")
	{

		v1.GET("/health", api.Health())
	}
	return e
}
