package errors

var (
	ErrInvalidInitData = BaseError{Msg: "invalid telegram init data", ErrC: "invalid_init_data"}
	ErrUnauthorized    = BaseError{Msg: "unauthorized", ErrC: "unauthorized"}
	ErrInvalidToken    = BaseError{Msg: "invalid or expired token", ErrC: "invalid_token"}
	ErrForbidden       = BaseError{Msg: "forbidden", ErrC: "forbidden"}
)
