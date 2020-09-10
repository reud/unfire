package model

type TweetSimple struct {
	CreatedAt string `json:"created_at"`
	IDStr     string `json:"id_str"`
	Text      string `json:"text"`
}
