package identity

// Role defines what the user can do in the system.
type Role string

const (
	RolePatient Role = "patient"
	RoleDoctor  Role = "doctor"
	RoleAdmin   Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RolePatient, RoleDoctor, RoleAdmin:
		return true
	default:
		return false
	}
}

func (r Role) String() string { return string(r) }

func (r Role) IsDoctor() bool { return r == RoleDoctor }

func (r Role) IsAdmin() bool { return r == RoleAdmin }
