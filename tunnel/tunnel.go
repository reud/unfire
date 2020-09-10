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
	un, err := client.GetUsername(token, secret)
	if err != nil {
		return err
	}
	c <- worker.User{
		Username:    *un,
		Token:       token,
		TokenSecret: secret,
	}
	log.Printf("joined: %+v", *un)
	return nil
}
