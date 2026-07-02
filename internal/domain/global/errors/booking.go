package errors

var (
	ErrInvalidPreferredDate     = BaseError{Msg: "preferred date is invalid", ErrC: "invalid_preferred_date"}
	ErrInvalidPreferredTime     = BaseError{Msg: "preferred time is invalid", ErrC: "invalid_preferred_time"}
	ErrInvalidPreferredVisitType = BaseError{Msg: "preferred visit type is invalid", ErrC: "invalid_preferred_visit_type"}
	ErrInvalidZoomLink        = BaseError{Msg: "zoom link is invalid", ErrC: "invalid_zoom_link"}
	ErrInvalidDecision        = BaseError{Msg: "decision is invalid", ErrC: "invalid_decision"}
	ErrDoctorNotesRequired    = BaseError{Msg: "doctor notes are required for rejection", ErrC: "doctor_notes_required"}
	ErrAppointmentNotVideo    = BaseError{Msg: "appointment is not a video consultation", ErrC: "appointment_not_video"}
	ErrAppointmentNotPending  = BaseError{Msg: "appointment is not pending", ErrC: "appointment_not_pending"}
	ErrUserNotFound           = BaseError{Msg: "user not found", ErrC: "user_not_found"}
	ErrAppointmentNotFound    = BaseError{Msg: "appointment not found", ErrC: "appointment_not_found"}
	ErrAppointmentCancelled   = BaseError{Msg: "appointment is cancelled", ErrC: "appointment_cancelled"}
)
