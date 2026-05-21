package errors

var (
	ErrInvalidPreferredDate = BaseError{Msg: "preferred date is invalid", ErrC: "invalid_preferred_date"}
	ErrInvalidPreferredTime = BaseError{Msg: "preferred time is invalid", ErrC: "invalid_preferred_time"}
	ErrUserNotFound         = BaseError{Msg: "user not found", ErrC: "user_not_found"}
)
