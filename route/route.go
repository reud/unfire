package route

import (
	"log"
	"unfire/api"
	"unfire/config"
	"unfire/domain/service"
	"unfire/infrastructure/client"
	"unfire/infrastructure/datastore"
	"unfire/infrastructure/repository"
	"unfire/interface/handler"
	admin2 "unfire/interface/handler/admin"
	handler2 "unfire/usecase/handler"
	"unfire/usecase/handler/admin"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

func Init(as service.AuthService, au handler2.AuthUseCase, si repository.SessionInitializer, ru admin.RestartUseCase) *echo.Echo {
	e := echo.New()

	shared := config.GetInstance()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(shared.SecurePhrase))))

	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding},
	}))
	e.Use(echoMw.Logger())

	auth := e.Group("/auth")
	{
		ah := handler.NewAuthHandler()
		tc := client.NewTwitterClientInitializer()
		ds, err := datastore.NewRedisDatastore()
		if err != nil {
			log.Fatal(err)
		}
		dc := service.NewDatastoreController(ds)

		auth.GET("/login", ah.GetLogin(au, as, si))
		auth.GET("/callback", ah.GetCallback(au, as, si, tc, dc))
		auth.GET("/stop", ah.GetStop(au, si))
	}

	e.GET("/health", api.Health())

	v1 := e.Group("/api/v1")
	{

		v1.GET("/health", api.Health())
	}

	adm := e.Group("/admin")
	{
		restart := adm.Group("/restart")
		{
			rh := admin2.NewRestartHandler()
			restart.GET("/delete", rh.GetDelete(ru))
			restart.GET("/reload", rh.GetReload(ru))
		}
	}

	return e
}
