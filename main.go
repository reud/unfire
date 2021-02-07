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

func startBatchService() {
	ds, err := datastore.NewRedisDatastore()
	if err != nil {
		panic(err)
	}
	dc := service.NewDatastoreController(ds)
	// start reload batch
	{
		bth := batch.NewReloadBatchService(time.Minute*3, dc)
		bth.Start()
	}
	// start delete batch
	{
		bth := batch.NewDeleteBatchService(time.Minute*3, dc)
		bth.Start()
	}
}

func main() {
	cfg := config.GetInstance()
	fmt.Printf("%+v", *cfg)
	startBatchService()
	as := service.NewAuthService()
	au := handler.NewAuthUseCase()
	si := repository.NewSessionInitializer()
	e := route.Init(as, au, si)
	if err := e.Start(":" + strconv.Itoa(cfg.Port)); err != nil {
		panic(err)
	}
}
