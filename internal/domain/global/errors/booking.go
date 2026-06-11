package errors

var (
	ErrInvalidPreferredDate = BaseError{Msg: "preferred date is invalid", ErrC: "invalid_preferred_date"}
	ErrInvalidPreferredTime = BaseError{Msg: "preferred time is invalid", ErrC: "invalid_preferred_time"}
	ErrInvalidZoomLink      = BaseError{Msg: "zoom link is invalid", ErrC: "invalid_zoom_link"}
	ErrUserNotFound         = BaseError{Msg: "user not found", ErrC: "user_not_found"}
	ErrAppointmentNotFound  = BaseError{Msg: "appointment not found", ErrC: "appointment_not_found"}
	ErrAppointmentCancelled = BaseError{Msg: "appointment is cancelled", ErrC: "appointment_cancelled"}
)
