package main

import (
	"unfire/route"
)


func main() {
	e := route.Init()
	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}
