package service

import (
	"github.com/garyburd/go-oauth/oauth"
	"github.com/pkg/errors"
	"unfire/config"
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
	callbackURL         = "https://unfire.reud.app/api/v1/auth/callback"
)

type AuthService interface {
	RequestTemporaryCredentialsAuthorizationURL() (*oauth.Credentials, string, error)
	GetAccessToken(rt *oauth.Credentials, oauthVerifier string) (*oauth.Credentials, error)
}

type authService struct {
	oauthClient *oauth.Client
}

// RequestTemporaryCredentialsAuthorizationURL 誘導するURLとリクエストトークンを生成する。
func (as *authService) RequestTemporaryCredentialsAuthorizationURL() (*oauth.Credentials, string, error) {
	rt, err := as.oauthClient.RequestTemporaryCredentials(nil, callbackURL, nil)
	if err != nil {
		return nil, "", err
	}
	return rt, as.oauthClient.AuthorizationURL(rt, nil), nil
}

// GetAccessToken リクエストトークンとベリファイアからアクセストークンを取得する。
func (as *authService) GetAccessToken(rt *oauth.Credentials, oauthVerifier string) (*oauth.Credentials, error) {
	at, _, err := as.oauthClient.RequestToken(nil, rt, oauthVerifier)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get access totken")
	}
	return at, nil
}

func NewAuthService() AuthService {
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

	return &authService{oauthClient: oc}
}
