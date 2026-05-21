package identity

// Role defines what the user can do in the system.
type Role string

const (
	RolePatient Role = "patient"
	RoleDoctor  Role = "doctor"
)

func (r Role) Valid() bool {
	switch r {
	case RolePatient, RoleDoctor:
		return true
	default:
		return false
	}
}

func (r Role) String() string { return string(r) }

func (r Role) IsDoctor() bool { return r == RoleDoctor }
