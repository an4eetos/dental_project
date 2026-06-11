package errors

import (
	"errors"
	"net/http"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Write maps domain errors to HTTP status and JSON body.
func Write(c interface {
	JSON(code int, obj any)
	AbortWithStatus(code int)
}, err error) {
	status, body := Translate(err)
	if body == nil {
		c.AbortWithStatus(status)
		return
	}
	c.JSON(status, body)
}

func Translate(err error) (int, *APIError) {
	var de domainerrors.DomainError
	if errors.As(err, &de) {
		return statusForCode(de.Code()), &APIError{Code: de.Code(), Message: de.Error()}
	}
	return http.StatusInternalServerError, &APIError{Code: "internal_error", Message: "internal server error"}
}

func Unauthorized() domainerrors.BaseError {
	return domainerrors.ErrUnauthorized
}

func statusForCode(code string) int {
	switch code {
	case domainerrors.ErrUnauthorized.Code(),
		domainerrors.ErrInvalidToken.Code(),
		domainerrors.ErrInvalidInitData.Code():
		return http.StatusUnauthorized
	case domainerrors.ErrInvalidPreferredDate.Code(),
		domainerrors.ErrInvalidPreferredTime.Code(),
		domainerrors.ErrInvalidZoomLink.Code(),
		domainerrors.ErrAppointmentCancelled.Code():
		return http.StatusBadRequest
	case domainerrors.ErrUserNotFound.Code(),
		domainerrors.ErrSubmissionNotFound.Code(),
		domainerrors.ErrAppointmentNotFound.Code():
		return http.StatusNotFound
	case domainerrors.ErrForbidden.Code():
		return http.StatusForbidden
	case domainerrors.ErrSubmissionAlreadyAnswered.Code(),
		domainerrors.ErrEmptyDoctorResponse.Code():
		return http.StatusBadRequest
	case domainerrors.ErrNoPredictionExamples.Code(),
		domainerrors.ErrPredictionFailed.Code(),
		domainerrors.ErrDraftGenerationFailed.Code():
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
