package booking

// Status represents appointment lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusRejected  Status = "rejected"
	StatusCancelled Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusRejected, StatusCancelled:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }
