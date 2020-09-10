package main

import (
	"strconv"
	"unfire/config"
	"unfire/route"
)

func main() {
	cfg := config.GetInstance()
	e := route.Init()
	if err := e.Start(":" + strconv.Itoa(cfg.Port)); err != nil {
		panic(err)
	}
}
