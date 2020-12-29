package client

import (
	"encoding/json"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"unfire/config"
	"unfire/domain/client"
	"unfire/domain/model"
)

const (
	refreshTokenURL     = "https://api.twitter.com/oauth/request_token"
	authorizationURL    = "https://api.twitter.com/oauth/authenticate"
	accessTokenURL      = "https://api.twitter.com/oauth/access_token"
	accountURL          = "https://api.twitter.com/1.1/account/verify_credentials.json"
	searchTweetURL      = "https://api.twitter.com/1.1/statuses/user_timeline.json"
	destroyTweetURL     = "https://api.twitter.com/1.1/statuses/destroy"
	getFavoritesURL     = "https://api.twitter.com/1.1/favorites/list.json"
	destroyFavoritesURL = "https://api.twitter.com/1.1/favorites/destroy.json"
)

type twitterClient struct {
	UserData *model.WorkerData
	at       *oauth.Credentials
}

func NewTwitterClient(at *oauth.Credentials) (client.TwitterClient, error) {
	wda, err := fetchProfile(at)
	if err != nil {
		return nil, err
	}
	return &twitterClient{
		UserData: wda,
		at:       at,
	}, nil
}

func fetchProfile(at *oauth.Credentials) (*model.WorkerData, error) {
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

	data := &model.WorkerData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = errors.Wrap(err, "Failed to decode user account response.")
		return nil, err
	}

	return data, nil
}

func (tc *twitterClient) FetchMe() *model.WorkerData {
	return tc.UserData
}

func (tc *twitterClient) FetchTweets(options ...client.FetchTweetOptionFunc) ([]model.Tweet, error) {
	var tweets []model.Tweet
	u, err := url.Parse(searchTweetURL)
	if err != nil {
		return tweets, err
	}

	q := u.Query()
	q.Set("user_id", tc.UserData.ID)
	q.Set("count", "150")
	oc := NewTWClient()
	resp, err := oc.Get(nil, tc.at, u.String(), q)
	if err != nil {
		return tweets, err
	}

	// TODO Tweetを構造体に直し、idを抽出する。
	// body
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&tweets)
	if err != nil {
		return tweets, err
	}
	return tweets, nil
}

func (tc *twitterClient) FetchFavorites() ([]model.Tweet, error) {
	var tweets []model.Tweet
	u, err := url.Parse(getFavoritesURL)
	if err != nil {
		return tweets, err
	}

	q := u.Query()
	q.Set("user_id", tc.UserData.ID)
	q.Set("count", "150")
	oc := NewTWClient()
	resp, err := oc.Get(nil, tc.at, u.String(), q)
	if err != nil {
		return tweets, err
	}

	// TODO Tweetを構造体に直し、idを抽出する。
	// body
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&tweets)
	if err != nil {
		return tweets, err
	}
	return tweets, nil
}

func (tc *twitterClient) DestroyTweet(tweetID string) error {
	u, err := url.Parse(destroyTweetURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, tweetID+".json")
	oc := NewTWClient()
	resp, err := oc.Post(nil, tc.at, u.String(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("%+v", string(body))
	return nil
}

func (tc *twitterClient) DestroyFavorite(tweetID string) error {
	u, err := url.Parse(destroyFavoritesURL)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("id", tweetID)

	oc := NewTWClient()
	resp, err := oc.Post(nil, tc.at, u.String(), q)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("%+v", string(body))
	return nil
}

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