package booking

// DoctorDecision is how a doctor responds to a pending appointment request.
type DoctorDecision string

const (
	DecisionInPerson DoctorDecision = "in_person"
	DecisionVideo    DoctorDecision = "video"
	DecisionReject   DoctorDecision = "reject"
)

func (d DoctorDecision) Valid() bool {
	switch d {
	case DecisionInPerson, DecisionVideo, DecisionReject:
		return true
	default:
		return false
	}
}

func (d DoctorDecision) String() string { return string(d) }
