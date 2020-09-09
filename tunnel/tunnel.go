package tunnel

import (
	"log"
	"unfire/client"
	"unfire/worker"
)

var c chan worker.User

func init() {
	c = make(chan worker.User)
	go worker.RunTask(c)
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
