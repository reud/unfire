package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"unfire/config"
	"unfire/domain/service"
	"unfire/infrastructure/datastore"
	"unfire/infrastructure/repository"
	"unfire/route"
	"unfire/usecase/batch"
	"unfire/usecase/handler"
	"unfire/usecase/handler/admin"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + " [DEBUG] " + string(bytes))
}

const location = "Asia/Tokyo"

// fix time
func init() {
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}
	time.Local = loc
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	log.Println("Unfire Started!")
}

func initBatchService() batch.Services {
	ds, err := datastore.NewRedisDatastore()
	if err != nil {
		panic(err)
	}
	dc := service.NewDatastoreController(ds)

	// start reload batch
	// TODO: ここを環境変数で設定可能にする。
	rbth := batch.NewReloadBatchService(time.Hour*3, dc)
	// TODO: ここを環境変数で設定可能にする。
	dbth := batch.NewDeleteBatchService(time.Hour*3, dc)
	return batch.NewServices(rbth, dbth)
}

func main() {
	cfg := config.GetInstance()
	fmt.Printf("%+v", *cfg)
	sv := initBatchService()

	sv.Launch()

	as := service.NewAuthService()

	au := handler.NewAuthUseCase()
	ru := admin.NewRestartUseCase(sv)

	si := repository.NewSessionInitializer()
	e := route.Init(as, au, si, ru)
	if err := e.Start(":" + strconv.Itoa(cfg.Port)); err != nil {
		panic(err)
	}
}
