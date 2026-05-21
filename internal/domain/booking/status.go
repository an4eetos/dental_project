package booking

// Status represents appointment lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusCancelled:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }
