package worker

import (
	"github.com/garyburd/go-oauth/oauth"
	"log"
	"time"
	"unfire/client"
	"unfire/model"
	"unfire/worker/hook"
)

const (
	DaysBefore = 32
)

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

func RunTask(u *model.User, waiting chan model.User) {

	defer func() {
		waiting <- *u
	}()

	var err error
	err = hook.PreRunTaskHook(u)
	if err != nil {
		log.Fatal(err)
	}

	at := &oauth.Credentials{
		Token:  u.Token,
		Secret: u.TokenSecret,
	}

	tts, err := client.GetSearchTweets(at, u.UserID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("tweet picked len: %+v", len(tts))
	go runDeleteTweetTask(&tts, at)

	likes, err := client.GetUserFavoriteTweets(at, u.UserID)
	if err != nil {
		log.Fatal(err)
	}
	go runDeleteFavoritesTask(&likes, at)
}

func runDeleteTweetTask(tts *[]model.TweetSimple, at *oauth.Credentials) {
	for _, tw := range *tts {
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

func runDeleteFavoritesTask(tts *[]model.TweetSimple, at *oauth.Credentials) {
	if len(*tts) < 30 {
		log.Printf("いいねが許容範囲内のため削除を中止します。")
		return
	}
	for _, tw := range *tts {
		tweet := tw
		go func() {
			if err := client.DestroyFavorites(at, tweet.IDStr); err != nil {
				log.Fatal(err)
			}
			log.Printf("destroied fav : %+v", tweet.Text)
		}()
	}
}
