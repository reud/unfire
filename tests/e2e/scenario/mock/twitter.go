package mock

import (
	"math/rand"
	"strconv"
	"time"
	"unfire/domain/client"
	"unfire/domain/model"
	client2 "unfire/infrastructure/client"
	"unfire/utils"

	"github.com/garyburd/go-oauth/oauth"
)

// TODO: モックTwitterClientの実装
// TODO: gomockとかを利用して後で引数テスト出来る様にする。
type TwitterClientInitializerImpl struct{}

func NewTwitterClientInitializer() client2.TwitterClientInitializer {
	return &TwitterClientInitializerImpl{}
}

type TwitterClientImpl struct {
	FetchMeFuncResult *model.WorkerData
	TweetPool         []model.Tweet
	FavoritesPool     []model.Tweet
	TweetIDToMap      map[string]int // tweetID -> TweetPool int
}

// idが空文字列の場合はランダムなidになる。二つ目の帰り値はtweetID -> TweetPool int
func generateRandTweets(n int) ([]model.Tweet, map[string]int) {
	var tweetIDToMap map[string]int
	var tweetPool []model.Tweet

	format := "2006-01-02T15:04:05.000Z"

	i := 0
	for i < n {
		var tweetID string
		rand.Seed(time.Now().Unix())
		uid := strconv.Itoa(rand.Int())
		tweetID = uid

		// 一日一ツイートで計算していく。
		now := time.Now()
		now = now.Add(time.Hour * time.Duration(-24*i))

		t := model.Tweet{
			ID:   tweetID,
			Text: "text_" + utils.RandString(15),
			PublicMetrics: model.GetUsersIdTweetsPublicMetrics{
				RetweetCount: rand.Intn(1000),
				ReplyCount:   rand.Intn(1000),
				LikeCount:    rand.Intn(1000),
				QuoteCount:   rand.Intn(20000),
			},
			CreatedAt: now.Format(format),
		}

		tweetPool = append(tweetPool, t)

		i++
	}

	return tweetPool, tweetIDToMap
}

// TwitterCLinetのMockを実装する。
func (tcii *TwitterClientInitializerImpl) NewTwitterClient(at *oauth.Credentials) (client.TwitterClient, error) {
	rand.Seed(time.Now().Unix())
	uid := strconv.Itoa(rand.Int())

	pool, mp := generateRandTweets(1000)
	fpool, _ := generateRandTweets(1000)

	return &TwitterClientImpl{
		FetchMeFuncResult: &model.WorkerData{
			ID:         uid,
			ScreenName: "screen_name_" + utils.RandString(5),
		},
		TweetPool:     pool,
		FavoritesPool: fpool,
		TweetIDToMap:  mp,
	}, nil
}

func (tc *TwitterClientImpl) FetchMe() *model.WorkerData {
	return tc.FetchMeFuncResult
}

func (tc *TwitterClientImpl) FetchTweets(options ...client.FetchTweetOptionFunc) ([]model.Tweet, error) {
	option := &client.FetchTweetOption{GetAll: false}

	for _, f := range options {
		f(option)
	}

	if option.GetAll {
		return tc.TweetPool, nil
	}

	return tc.TweetPool[:100], nil
}

func (tc *TwitterClientImpl) FetchFavorites() ([]model.Tweet, error) {
	return tc.FavoritesPool[:150], nil
}

func (tc *TwitterClientImpl) DestroyTweet(tweetID string) error {
	return nil
}

func (tc *TwitterClientImpl) DestroyFavorite(tweetID string) error {
	return nil
}

func (tc *TwitterClientImpl) FetchTweetFromIDStr(tweetID string) (*model.Tweet, error) {
	return &tc.TweetPool[tc.TweetIDToMap[tweetID]], nil
}
