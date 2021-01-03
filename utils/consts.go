package utils

type UserStatus int

const (
	Initializing UserStatus = iota
	Deleted
	Working
)

func (u UserStatus) String() string {
	switch u {
	case Initializing:
		return "Initializing"
	case Deleted:
		return "Deleted"
	case Working:
		return "Working"
	default:
		return "Unknown"
	}
}

const (
	TimeLine       = "timeline" // Valueは time_userIDで持つ memberはvalueで持つ。
	TimeLinePrefix = 9007199254740992
	StatusSuffix   = "_status"
	IgnoresSuffix  = "_ignore"
	TweetPrefix    = "tweet_"
	TweetsSuffix   = "_tweets"
	Users          = "users"
	TokenSuffix    = "token_"
)
