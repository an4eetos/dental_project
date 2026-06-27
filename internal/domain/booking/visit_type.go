package booking

// VisitType describes how a confirmed appointment takes place.
type VisitType string

const (
	VisitTypeInPerson VisitType = "in_person"
	VisitTypeVideo    VisitType = "video"
)

func (v VisitType) Valid() bool {
	switch v {
	case VisitTypeInPerson, VisitTypeVideo:
		return true
	default:
		return false
	}
}

func (v VisitType) String() string { return string(v) }
