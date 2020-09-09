package worker

import (
	"github.com/garyburd/go-oauth/oauth"
	"unfire/client"
)

type User struct {
	Username    string
	Token       string
	TokenSecret string
}

func RunTask(cl chan User) {
	for u := range cl {
		at := &oauth.Credentials{
			Token:  u.Token,
			Secret: u.TokenSecret,
		}
		if err := client.GetSearchTweets(at, u.Username); err != nil {
			panic(err)
		}
	}
}
