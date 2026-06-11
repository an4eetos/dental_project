package photoreview

// Status tracks whether a clinic doctor has responded to a patient photo.
type Status string

const (
	StatusPending  Status = "pending"
	StatusAnswered Status = "answered"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusAnswered:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }
