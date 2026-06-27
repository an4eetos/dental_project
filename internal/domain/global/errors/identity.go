package errors

var (
	ErrUserBlocked      = BaseError{Msg: "user is blocked", ErrC: "user_blocked"}
	ErrCannotBlockSelf  = BaseError{Msg: "cannot block yourself", ErrC: "cannot_block_self"}
	ErrCannotBlockAdmin = BaseError{Msg: "cannot block an admin", ErrC: "cannot_block_admin"}
	ErrInvalidRole      = BaseError{Msg: "invalid role", ErrC: "invalid_role"}
	ErrInvalidProfile   = BaseError{Msg: "invalid profile data", ErrC: "invalid_profile"}
)
