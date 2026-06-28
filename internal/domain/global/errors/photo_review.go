package errors

var (
	ErrSubmissionNotFound        = BaseError{Msg: "Заявка не найдена", ErrC: "submission_not_found"}
	ErrSubmissionAlreadyAnswered = BaseError{Msg: "На эту заявку уже дан ответ", ErrC: "submission_already_answered"}
	ErrEmptyDoctorResponse       = BaseError{Msg: "Требуется ответ врача", ErrC: "empty_doctor_response"}
	ErrDraftGenerationFailed     = BaseError{Msg: "Не удалось сгенерировать черновик ИИ", ErrC: "draft_generation_failed"}
	ErrSubmissionMediaTooLarge   = BaseError{Msg: "Файл слишком большой", ErrC: "submission_media_too_large"}
)
