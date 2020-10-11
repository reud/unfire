package tunnel

import (
	"log"
	"time"
	"unfire/client"
	"unfire/model"
	"unfire/worker"
)

var c chan model.User
var w chan model.User

func init() {
	c = make(chan model.User)
	w = make(chan model.User)
	go runTaskChannel(c, w)
	go waitingTaskChannel(c, w)
}

func runTaskChannel(cl chan model.User, wa chan model.User) {
	for u := range cl {
		go worker.RunTask(&u, wa)
	}
}

func waitingTaskChannel(cl chan model.User, wa chan model.User) {
	for u := range wa {
		user := u
		go func() {
			time.Sleep(time.Minute * 15)
			cl <- user
		}()
	}
}

func AddUserByCredentials(token string, secret string, options model.Options) error {
	ui, err := client.GetUserID(token, secret)
	if err != nil {
		return err
	}
	w <- model.User{
		UserID:      *ui,
		Token:       token,
		TokenSecret: secret,
		Options:     options,
	}
	log.Printf("joined: %+v", *ui)
	return nil
}
