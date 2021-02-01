package model

type WorkerData struct {
	ID         string `json:"id_str"`
	ScreenName string `json:"screen_name"`
}

type GetUsersIdTweetsPublicMetrics struct {
	RetweetCount int `json:"retweet_count"`
	ReplyCount   int `json:"reply_count"`
	LikeCount    int `json:"like_count"`
	QuoteCount   int `json:"quote_count"`
}

type GetUsersIdTweetsMeta struct {
	OldestId    string `json:"oldest_id"`
	NewestId    string `json:"newest_id"`
	ResultCount int    `json:"result_count"`
	NextToken   string `json:"next_token"`
}

type GetUsersIdTweetsResponse struct {
	Data []Tweet              `json:"data"`
	Meta GetUsersIdTweetsMeta `json:"meta"`
}

type Tweet struct {
	ID            string                        `json:"id"`
	Text          string                        `json:"text,omitempty"`
	PublicMetrics GetUsersIdTweetsPublicMetrics `json:"public_metrics"`
	CreatedAt     string                        `json:"created_at"`
}
