package tunnel

import (
	"log"
	"unfire/client"
	"unfire/worker"
)

var c chan worker.User
var w chan worker.User

func init() {
	c = make(chan worker.User)
	w = make(chan worker.User)
	go worker.RunTaskChannel(c, w)
	go worker.WaitingTaskChannel(c, w)
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
