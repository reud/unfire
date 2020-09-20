package tunnel

import (
	"log"
	"time"
	"unfire/client"
	"unfire/worker"
)

var c chan worker.User
var w chan worker.User

func init() {
	c = make(chan worker.User)
	w = make(chan worker.User)
	go runTaskChannel(c, w)
	go waitingTaskChannel(c, w)
}

func runTaskChannel(cl chan worker.User, wa chan worker.User) {
	for u := range cl {
		go worker.RunTask(&u, wa)
	}
}

func waitingTaskChannel(cl chan worker.User, wa chan worker.User) {
	for u := range wa {
		user := u
		go func() {
			time.Sleep(time.Minute * 15)
			cl <- user
		}()
	}
}

func AddUserByCredentials(token string, secret string) error {
	ui, err := client.GetUserID(token, secret)
	if err != nil {
		return err
	}
	w <- worker.User{
		UserID:      *ui,
		Token:       token,
		TokenSecret: secret,
	}
	log.Printf("joined: %+v", *ui)
	return nil
}
