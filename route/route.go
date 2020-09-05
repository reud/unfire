package route

import (
	"github.com/labstack/echo"
	echoMw "github.com/labstack/echo/middleware"
	"unfire/api"
)

func Init() *echo.Echo {
	e := echo.New()
	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding},
	}))

	// routes
	v1 := e.Group("/api/v1")
	{
		v1.GET("/health",api.Health())
	}
	return e
}