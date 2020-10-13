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
	CallbackURL                string
}
