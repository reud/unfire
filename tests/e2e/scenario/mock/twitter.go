package mock

import (
	"unfire/domain/client"
	"unfire/domain/model"

	"github.com/garyburd/go-oauth/oauth"
)

// TODO: モックTwitterClientの実装
type TwitterClientInitializerImpl struct {
	FetchMeFuncResult *model.WorkerData
	TweetPool         []model.Tweet
}

func (tcii *TwitterClientInitializerImpl) NewTwitterClient(at *oauth.Credentials) (client.TwitterClient, error) {

}
