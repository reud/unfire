package client

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"unfire/config"
)

const (
	refreshTokenURL  = "https://api.twitter.com/oauth/request_token"
	authorizationURL = "https://api.twitter.com/oauth/authenticate"
	accessTokenURL   = "https://api.twitter.com/oauth/access_token"
	accountURL       = "https://api.twitter.com/1.1/account/verify_credentials.json"
	searchTweetURL   = "https://api.twitter.com/1.1/statuses/user_timeline.json"
)

type MyData struct {
	ID         string `json:"id_str"`
	ScreenName string `json:"screen_name"`
}

func NewTWClient() *oauth.Client {
	configInstance := config.GetInstance()
	oc := &oauth.Client{
		TemporaryCredentialRequestURI: refreshTokenURL,
		ResourceOwnerAuthorizationURI: authorizationURL,
		TokenRequestURI:               accessTokenURL,
		Credentials: oauth.Credentials{
			Token:  configInstance.TwitterConsumerKey,
			Secret: configInstance.TwitterConsumerSecret,
		},
	}

	return oc
}

func GetAccessToken(rt *oauth.Credentials, oauthVerifier string) (int, *oauth.Credentials, error) {
	oc := NewTWClient()
	at, _, err := oc.RequestToken(nil, rt, oauthVerifier)
	if err != nil {
		err := errors.Wrap(err, "Failed to get access token.")
		return http.StatusBadRequest, nil, err
	}
	return http.StatusOK, at, nil
}

func GetUsername(token string, secret string) (*string, error) {
	at := &oauth.Credentials{
		Token:  token,
		Secret: secret,
	}
	d, err := GetMe(at)
	if err != nil {
		return nil, err
	}
	return &d.ScreenName, err
}

func GetMe(at *oauth.Credentials) (*MyData, error) {
	oc := NewTWClient()
	resp, err := oc.Get(nil, at, accountURL, nil)
	if err != nil {
		err = errors.Wrap(err, "Failed to send twitter request.")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		err = errors.New("Twitter is unavailable")
		return nil, err
	}

	if resp.StatusCode >= 400 {
		err = errors.New("Twitter request is invalid")
		return nil, err
	}

	data := &MyData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = errors.Wrap(err, "Failed to decode user account response.")
		return nil, err
	}

	return data, nil

}

// YYYY-MM-DD の形式で前日の日付を取得する。
func getUntilQuery() string {
	return time.Now().Add(-time.Duration(24) * time.Hour).String()[:10]
}

func GetSearchTweets(at *oauth.Credentials, username string) error {
	u, err := url.Parse(searchTweetURL)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("screen_name", username)
	oc := NewTWClient()
	resp, err := oc.Get(nil, at, u.String(), q)
	if err != nil {
		return err
	}

	// TODO Tweetを構造体に直し、idを抽出する。
	// body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", string(body))
	return nil
}
