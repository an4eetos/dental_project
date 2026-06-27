package booking

import (
	"fmt"
	"strings"
)

const nonDiagnosticNotice = "\n\n⚠️ Онлайн-консультация носит информационный характер и не заменяет очный приём у стоматолога."

func zoomSection(zoomLink string) string {
	if strings.TrimSpace(zoomLink) != "" {
		return fmt.Sprintf("\n\nСсылка на Zoom:\n%s", zoomLink)
	}
	return "\n\nСсылка на Zoom будет добавлена позже — мы пришлём её в этом чате."
}

// FormatNewRequestDoctorMessage notifies doctors about a pending appointment.
func FormatNewRequestDoctorMessage(patientName, date, timeStr string) string {
	return fmt.Sprintf(
		"Новая заявка на приём.\n\nПациент: %s\nЗапрошено: %s в %s\n\nОткройте мини-приложение → вкладка «Записи», чтобы подтвердить, предложить видеоконсультацию или перенести.",
		patientName,
		date,
		timeStr,
	)
}

// FormatInPersonConfirmedMessage builds the patient message for an in-person visit.
func FormatInPersonConfirmedMessage(date, timeStr, notes string) string {
	msg := fmt.Sprintf(
		"Ваш очный приём подтверждён!\n\nДата: %s\nВремя: %s",
		date,
		timeStr,
	)
	if strings.TrimSpace(notes) != "" {
		msg += fmt.Sprintf("\n\nКомментарий врача:\n%s", strings.TrimSpace(notes))
	}
	return msg
}

// FormatVideoConfirmedMessage builds the patient message for a video consultation.
func FormatVideoConfirmedMessage(date, timeStr, zoomLink string) string {
	return fmt.Sprintf(
		"Ваша видеоконсультация подтверждена!\n\nДата: %s\nВремя: %s%s%s",
		date,
		timeStr,
		zoomSection(zoomLink),
		nonDiagnosticNotice,
	)
}

// FormatRejectedMessage builds the patient message when a doctor rejects / reschedules.
func FormatRejectedMessage(date, timeStr, doctorNotes string) string {
	return fmt.Sprintf(
		"К сожалению, запрошенное время (%s в %s) недоступно.\n\nВрач предлагает другие варианты:\n%s\n\nПожалуйста, отправьте новую заявку на удобное время через мини-приложение.",
		date,
		timeStr,
		strings.TrimSpace(doctorNotes),
	)
}

// FormatZoomLinkAddedMessage notifies the patient when a doctor adds a Zoom link.
func FormatZoomLinkAddedMessage(date, timeStr, zoomLink string) string {
	return fmt.Sprintf(
		"Ссылка на видеоконсультацию готова!\n\nДата: %s\nВремя: %s\n\nСсылка на Zoom:\n%s%s",
		date,
		timeStr,
		zoomLink,
		nonDiagnosticNotice,
	)
}

// FormatDoctorZoomMissingMessage reminds the doctor to add a Zoom link.
func FormatDoctorZoomMissingMessage(patientName, date, timeStr string) string {
	return fmt.Sprintf(
		"Нужна ссылка на Zoom.\n\nПациент: %s\nКонсультация: %s в %s\n\nДобавьте ссылку в мини-приложении → «Записи» → выберите запись.",
		patientName,
		date,
		timeStr,
	)
}

// FormatPatientReminderOneDay builds the patient reminder one day before.
func FormatPatientReminderOneDay(date, timeStr, visitType, zoomLink string) string {
	switch visitType {
	case "in_person":
		return fmt.Sprintf(
			"Напоминание: завтра у вас очный приём.\n\nДата: %s\nВремя: %s",
			date,
			timeStr,
		)
	default:
		return fmt.Sprintf(
			"Напоминание: завтра у вас видеоконсультация.\n\nДата: %s\nВремя: %s%s%s",
			date,
			timeStr,
			zoomSection(zoomLink),
			nonDiagnosticNotice,
		)
	}
}

// FormatPatientReminderOneHour builds the patient reminder one hour before.
func FormatPatientReminderOneHour(date, timeStr, visitType, zoomLink string) string {
	switch visitType {
	case "in_person":
		return fmt.Sprintf(
			"Напоминание: через 1 час у вас очный приём.\n\nДата: %s\nВремя: %s",
			date,
			timeStr,
		)
	default:
		return fmt.Sprintf(
			"Напоминание: через 1 час у вас видеоконсультация.\n\nДата: %s\nВремя: %s%s%s",
			date,
			timeStr,
			zoomSection(zoomLink),
			nonDiagnosticNotice,
		)
	}
}

// FormatDoctorReminderOneDay builds the doctor reminder one day before.
func FormatDoctorReminderOneDay(patientName, date, timeStr, visitType, zoomLink string) string {
	kind := "очный приём"
	if visitType == "video" {
		kind = "видеоконсультация"
	}
	msg := fmt.Sprintf(
		"Напоминание: завтра %s.\n\nПациент: %s\nДата: %s\nВремя: %s",
		kind,
		patientName,
		date,
		timeStr,
	)
	if visitType == "video" && strings.TrimSpace(zoomLink) == "" {
		msg += "\n\n⚠️ Ссылка на Zoom ещё не добавлена — укажите её в мини-приложении."
	}
	return msg
}

// FormatDoctorReminderOneHour builds the doctor reminder one hour before.
func FormatDoctorReminderOneHour(patientName, date, timeStr, visitType, zoomLink string) string {
	kind := "очный приём"
	if visitType == "video" {
		kind = "видеоконсультация"
	}
	msg := fmt.Sprintf(
		"Напоминание: через 1 час %s.\n\nПациент: %s\nДата: %s\nВремя: %s",
		kind,
		patientName,
		date,
		timeStr,
	)
	if visitType == "video" && strings.TrimSpace(zoomLink) == "" {
		msg += "\n\n⚠️ Ссылка на Zoom ещё не добавлена — укажите её в мини-приложении."
	}
	return msg
}

// FormatOfferMessage is kept for backward compatibility in tests.
func FormatOfferMessage(date, timeStr, zoomLink string) string {
	return FormatVideoConfirmedMessage(date, timeStr, zoomLink)
}

// FormatReminderOneDayMessage is kept for backward compatibility.
func FormatReminderOneDayMessage(date, timeStr, zoomLink string) string {
	return FormatPatientReminderOneDay(date, timeStr, "video", zoomLink)
}

// FormatReminderOneHourMessage is kept for backward compatibility.
func FormatReminderOneHourMessage(date, timeStr, zoomLink string) string {
	return FormatPatientReminderOneHour(date, timeStr, "video", zoomLink)
}
