package main

import (
	"strconv"
	"time"
	"unfire/config"
	"unfire/route"
)

const location = "Asia/Tokyo"

// fix time
func init() {
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}
	time.Local = loc
}

func main() {
	cfg := config.GetInstance()
	e := route.Init()
	if err := e.Start(":" + strconv.Itoa(cfg.Port)); err != nil {
		panic(err)
	}
}
