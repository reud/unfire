package service

import (
	"fmt"
	"strings"
	"unfire/config"

	"github.com/garyburd/go-oauth/oauth"
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

var callbackURL string

func init() {
	cfg := config.GetInstance()
	if strings.Contains(cfg.Domain, "localhost") {
		callbackURL = fmt.Sprintf("http://%+v/auth/callback", cfg.Domain)
	} else {
		callbackURL = fmt.Sprintf("https://%+v/auth/callback", cfg.Domain)
	}
}

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
		// TODO: ここでひっかかる。調査必要
		return nil, "", err
	}
	return rt, as.oauthClient.AuthorizationURL(rt, nil), nil
}

// GetAccessToken リクエストトークンとベリファイアからアクセストークンを取得する。
func (as *authService) GetAccessToken(rt *oauth.Credentials, oauthVerifier string) (*oauth.Credentials, error) {
	at, _, err := as.oauthClient.RequestToken(nil, rt, oauthVerifier)
	if err != nil {
		return nil, err
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
