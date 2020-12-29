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
	TimeLine      = "timeline"
	StatusSuffix  = "_status"
	IgnoresSuffix = "_ignore"
)
