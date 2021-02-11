package mock

import "github.com/garyburd/go-oauth/oauth"

type mockAuthService struct {
}

func (mas *mockAuthService) RequestTemporaryCredentialsAuthorizationURL() (*oauth.Credentials, string, error) {
	return &oauth.Credentials{
		Token:  "mock token",
		Secret: "mock secret",
	}, "http://example.com/authorize", nil
}

func (mas *mockAuthService) GetAccessToken(rt *oauth.Credentials, oauthVerifier string) (*oauth.Credentials, error) {
	return &oauth.Credentials{
		Token:  "mock access token",
		Secret: "mock access secret",
	}, nil
}

func NewMockAuthService() *mockAuthService {
	return &mockAuthService{}
}
