package errors

var (
	ErrSubmissionNotFound      = BaseError{Msg: "photo submission not found", ErrC: "submission_not_found"}
	ErrSubmissionAlreadyAnswered = BaseError{Msg: "photo submission already answered", ErrC: "submission_already_answered"}
	ErrEmptyDoctorResponse     = BaseError{Msg: "doctor response is required", ErrC: "empty_doctor_response"}
	ErrDraftGenerationFailed   = BaseError{Msg: "failed to generate AI draft", ErrC: "draft_generation_failed"}
)
