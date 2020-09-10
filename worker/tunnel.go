package worker

import (
	"github.com/garyburd/go-oauth/oauth"
	"log"
	"time"
	"unfire/client"
)

type User struct {
	Username    string
	Token       string
	TokenSecret string
}

func RunTaskChannel(cl chan User, wa chan User) {
	for u := range cl {
		go runTask(&u, wa)
	}
}

func WaitingTaskChannel(cl chan User, wa chan User) {
	for u := range wa {
		time.Sleep(time.Minute * 3)
		cl <- u
	}
}

func runTask(u *User, waiting chan User) {

	defer func() {
		waiting <- *u
	}()

	at := &oauth.Credentials{
		Token:  u.Token,
		Secret: u.TokenSecret,
	}
	log.Printf("tweet pickking from: %+v", u.Username)
	tts, err := client.GetSearchTweets(at, u.Username)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("tweet picked len: %+v", len(tts))
	for _, tw := range tts {
		tweet := tw
		go func() {
			if err := client.DestroyTweet(at, tweet.IDStr); err != nil {
				log.Fatal(err)
			}
			log.Printf("destroy tweet : %+v", tweet.Text)
		}()
	}
}
