package model

type User struct {
	UserID      string
	Token       string
	TokenSecret string
	Options     Options
}

type Options struct {
	DeleteLike                 bool
	DeleteLikeCount            int
	KeepLegendaryTweetV1Enable bool
	KeepLegendaryTweetV1Border int
}

func NewUser(userID string, token string, tokenSecret string, options Options) *User {
	return &User{
		UserID:      userID,
		Token:       token,
		TokenSecret: tokenSecret,
		Options:     options,
	}
}
