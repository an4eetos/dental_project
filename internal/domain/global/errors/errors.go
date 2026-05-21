package errors

// DomainError is implemented by all typed domain errors.
type DomainError interface {
	error
	Code() string
}

// BaseError provides Code() for domain errors with a stable message.
type BaseError struct {
	Msg  string
	ErrC string
}

func (e BaseError) Error() string { return e.Msg }
func (e BaseError) Code() string  { return e.ErrC }
