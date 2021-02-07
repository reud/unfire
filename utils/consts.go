package utils

type UserStatus int

const (
	Initializing UserStatus = iota
	Deleted
	Working
	Waiting
)

func (u UserStatus) String() string {
	switch u {
	case Initializing:
		return "Initializing"
	case Deleted:
		return "Deleted"
	case Working:
		return "Working"
	case Waiting:
		return "Waiting"
	default:
		return "Unknown"
	}
}

// 存在しない場合はエラーを返す。
func StrToUserStatus(statusStr string) UserStatus {
	switch statusStr {
	case Waiting.String():
		return Waiting
	case Initializing.String():
		return Initializing
	case Deleted.String():
		return Deleted
	case Working.String():
		return Working
	default:
		panic("failed to convert user status() (StrToUserStatus)")
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
