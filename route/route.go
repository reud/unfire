package route

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
	"unfire/api"
)

func Init() *echo.Echo {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding},
	}))
	e.Use(echoMw.Logger())

	// routes
	v1 := e.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.GET("/login", api.LoginByTwitter())
			auth.GET("/callback", api.TwitterCallback())
			auth.GET("/force", api.ForceLoginByTwitter())
		}
		v1.GET("/health", api.Health())
	}
	return e
}
