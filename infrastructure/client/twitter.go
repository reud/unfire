package client

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"strings"
	"unfire/config"
	"unfire/domain/client"
	"unfire/domain/model"
	"unfire/utils"

	"github.com/garyburd/go-oauth/oauth"
)

const (
	refreshTokenURL     = "https://api.twitter.com/oauth/request_token"
	authorizationURL    = "https://api.twitter.com/oauth/authenticate"
	accessTokenURL      = "https://api.twitter.com/oauth/access_token"
	accountURL          = "https://api.twitter.com/1.1/account/verify_credentials.json"
	_searchTweetURL     = "https://api.twitter.com/2/users/:id/tweets" // V2 Beta URL https://developer.twitter.com/en/docs/twitter-api/tweets/timelines/api-reference/get-users-id-tweets
	destroyTweetURL     = "https://api.twitter.com/1.1/statuses/destroy"
	getFavoritesURL     = "https://api.twitter.com/1.1/favorites/list.json"
	destroyFavoritesURL = "https://api.twitter.com/1.1/favorites/destroy.json"
	_getTweetsIDURL     = "https://api.twitter.com/2/tweets/:id" // V2 Beta URL https://developer.twitter.com/en/docs/twitter-api/tweets/lookup/api-reference/get-tweets-id
)

func generateSearchTweetURL(id string) string {
	return strings.ReplaceAll(_searchTweetURL, ":id", id)
}

func generateGetTweetsIDURL(id string) string {
	return strings.ReplaceAll(_getTweetsIDURL, ":id", id)
}

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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, err
	}

	data := &model.WorkerData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (tc *twitterClient) FetchMe() *model.WorkerData {
	return tc.UserData
}

func fetchTweetDefaultOption() client.FetchTweetOption {
	return client.FetchTweetOption{
		GetAll: false,
	}
}

// TODO: since_idにバグあり。 https://memo.furyutei.work/entry/20100124/1264342029 pagingで対応する。
// FetchTweets ツイート取得する。
func (tc *twitterClient) FetchTweets(options ...client.FetchTweetOptionFunc) ([]model.Tweet, error) {
	log.Printf("start fetch tweets \n")

	// パラメータの取得
	option := fetchTweetDefaultOption()
	for _, f := range options {
		f(&option)
	}

	// 全件取得フラグがtrueなら全件取得メソッドを別で呼び出して終了する。
	if option.GetAll {
		log.Printf("get all tweets mode on \n")
		return fetchAllTweets(tc)
	}

	var tweets []model.Tweet

	searchTweetURL := generateSearchTweetURL(tc.UserData.ID)
	u, err := url.Parse(searchTweetURL)
	if err != nil {
		return tweets, err
	}

	q := u.Query()

	q.Set("max_results", "100")
	q.Set("tweet.fields", "id,text,public_metrics,created_at")
	oc := NewTWClient()
	resp, err := oc.Get(nil, tc.at, u.String(), q)
	if err != nil {
		return tweets, err
	}

	var response model.GetUsersIdTweetsResponse

	// TODO Tweetを構造体に直し、idを抽出する。
	// body
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return tweets, err
	}

	for _, v := range response.Data {
		tweets = append(tweets, v)
		log.Printf("tweet data: %+v", v)
	}

	return tweets, nil
}

// 全件ツイートを取得する。
func fetchAllTweets(tc *twitterClient) ([]model.Tweet, error) {
	var tweets []model.Tweet

	// URLの作成
	searchTweetURL := generateSearchTweetURL(tc.UserData.ID)
	// URLのvalidate
	u, err := url.Parse(searchTweetURL)
	if err != nil {
		return tweets, err
	}
	q := u.Query()
	q.Set("max_results", "100")
	q.Set("tweet.fields", "id,text,public_metrics,created_at")
	oc := NewTWClient()

	var childFunc func(isFirst bool, nextToken string) error

	childFunc = func(isFirst bool, nextToken string) error {
		query := q
		if !isFirst {
			query.Set("pagination_token", nextToken)
		}

		resp, err := oc.Get(nil, tc.at, u.String(), q)
		if err != nil {
			log.Printf("error occuered: %+v", err)
			return err
		}

		var response model.GetUsersIdTweetsResponse

		// TODO Tweetを構造体に直し、idを抽出する。
		// body
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			utils.DebugResponse(resp.Body)
			log.Printf("error occuered: %+v", err)
			return err
		}

		for _, v := range response.Data {
			tweets = append(tweets, v)
		}

		log.Printf("get data success! meta: %+v", response.Meta)

		// 一番最後の場合はnext_tokenの値は空文字列になることを利用する。
		if response.Meta.NextToken != "" {
			return childFunc(false, response.Meta.NextToken)
		}
		return nil
	}

	if err := childFunc(true, ""); err != nil {
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

func (tc *twitterClient) FetchTweetFromIDStr(tweetID string) (*model.Tweet, error) {
	u, err := url.Parse(generateGetTweetsIDURL(tweetID))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("tweet.fields", "id,text,public_metrics,created_at")

	oc := NewTWClient()
	resp, err := oc.Get(nil, tc.at, u.String(), q)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody := struct {
		Data model.Tweet `json:"data"`
	}{}
	// utils.DebugResponse(resp.Body)
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, err
	}

	return &responseBody.Data, err
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
