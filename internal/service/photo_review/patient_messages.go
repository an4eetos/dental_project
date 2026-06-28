package photoreview

import (
	"strings"

	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
)

const doctorReplySuffix = "\n\n⚠️ Справочная информация, не медицинская консультация. " +
	"При симптомах, боли или тревоге за здоровье обратитесь к стоматологу лично."

const patientAckPhotoMessage = "Фото получено.\n\n" +
	"Врач клиники рассмотрит его и пришлёт ответ в этот чат в течение 48 часов.\n\n" +
	"⚠️ Ответ врача — справочная информация, не медицинская консультация. " +
	"При симптомах, боли или тревоге за здоровье обратитесь к стоматологу лично."

const patientAckVideoMessage = "Видео получено.\n\n" +
	"Врач клиники рассмотрит его и пришлёт ответ в этот чат в течение 48 часов.\n\n" +
	"⚠️ Ответ врача — справочная информация, не медицинская консультация. " +
	"При симптомах, боли или тревоге за здоровье обратитесь к стоматологу лично."

// PatientAckMessage confirms receipt of a photo or video submission.
func PatientAckMessage(mediaType photoreview.MediaType) string {
	if mediaType == photoreview.MediaTypeVideo {
		return patientAckVideoMessage
	}
	return patientAckPhotoMessage
}

// FormatDoctorReply builds the Telegram message sent to the patient after review.
func FormatDoctorReply(mediaType photoreview.MediaType, response string) string {
	return doctorReplyPrefix(mediaType) + response + doctorReplySuffix
}

func doctorReplyPrefix(mediaType photoreview.MediaType) string {
	if mediaType == photoreview.MediaTypeVideo {
		return "Ответ врача клиники по вашему видео:\n\n"
	}
	return "Ответ врача клиники по вашему фото:\n\n"
}

// MediaTypeLabelRU returns a short Russian label for UI lists.
func MediaTypeLabelRU(mediaType photoreview.MediaType) string {
	if mediaType == photoreview.MediaTypeVideo {
		return "Видео"
	}
	return "Фото"
}

// MediaTypeDetailRU returns a Russian type line for detail views.
func MediaTypeDetailRU(mediaType photoreview.MediaType) string {
	return "Тип: " + strings.ToLower(MediaTypeLabelRU(mediaType))
}
