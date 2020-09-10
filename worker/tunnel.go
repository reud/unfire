package worker

import (
	"github.com/garyburd/go-oauth/oauth"
	"log"
	"time"
	"unfire/client"
	"unfire/model"
)

type User struct {
	Username    string
	Token       string
	TokenSecret string
}

const (
	DaysBefore = 32
)

func RunTaskChannel(cl chan User, wa chan User) {
	for u := range cl {
		go runTask(&u, wa)
	}
}

func WaitingTaskChannel(cl chan User, wa chan User) {
	for u := range wa {
		time.Sleep(time.Minute * 15)
		cl <- u
	}
}

func isOldTweet(ms *model.TweetSimple) (bool, error) {
	t, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", ms.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}
	ago := time.Since(t)

	if ago > time.Hour*DaysBefore {
		return true, nil
	}
	return false, nil
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
			ok, err := isOldTweet(&tweet)
			if err != nil {
				log.Fatal(err)
				return
			}
			if !ok {
				return
			}
			if err := client.DestroyTweet(at, tweet.IDStr); err != nil {
				log.Fatal(err)
			}
			log.Printf("destroied tweet : %+v", tweet.Text)
		}()
	}
}
