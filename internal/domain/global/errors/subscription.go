package errors

var (
	ErrSubscriptionRequired   = BaseError{Msg: "subscription required", ErrC: "subscription_required"}
	ErrSubscriptionNotFound   = BaseError{Msg: "subscription not found", ErrC: "subscription_not_found"}
	ErrInvalidPaymentPayload  = BaseError{Msg: "invalid payment payload", ErrC: "invalid_payment_payload"}
	ErrPaymentAlreadyRecorded = BaseError{Msg: "payment already recorded", ErrC: "payment_already_recorded"}
)
