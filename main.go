package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"unfire/config"
	"unfire/route"
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

func main() {
	cfg := config.GetInstance()
	e := route.Init()
	if err := e.Start(":" + strconv.Itoa(cfg.Port)); err != nil {
		panic(err)
	}
}
