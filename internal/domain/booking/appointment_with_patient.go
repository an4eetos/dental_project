package booking

import "github.com/anuarkuanysh/dental_project/internal/domain/identity"

// AppointmentWithPatient is an appointment plus patient identity for clinic staff.
type AppointmentWithPatient struct {
	Appointment Appointment
	Patient     identity.PatientSummary
}
