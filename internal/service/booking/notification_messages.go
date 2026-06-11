package booking

import "fmt"

const nonDiagnosticNotice = "\n\n⚠️ Онлайн-консультация носит информационный характер и не заменяет очный приём у стоматолога."

// FormatOfferMessage builds the initial confirmation sent to the patient.
func FormatOfferMessage(date, timeStr, zoomLink string) string {
	return fmt.Sprintf(
		"Ваша онлайн-консультация подтверждена!\n\nДата: %s\nВремя: %s\n\nСсылка на Zoom:\n%s%s",
		date,
		timeStr,
		zoomLink,
		nonDiagnosticNotice,
	)
}

// FormatReminderOneDayMessage builds the reminder sent one day before the slot.
func FormatReminderOneDayMessage(date, timeStr, zoomLink string) string {
	return fmt.Sprintf(
		"Напоминание: завтра у вас онлайн-консультация.\n\nДата: %s\nВремя: %s\n\nСсылка на Zoom:\n%s%s",
		date,
		timeStr,
		zoomLink,
		nonDiagnosticNotice,
	)
}

// FormatReminderOneHourMessage builds the reminder sent one hour before the slot.
func FormatReminderOneHourMessage(date, timeStr, zoomLink string) string {
	return fmt.Sprintf(
		"Напоминание: через 1 час у вас онлайн-консультация.\n\nДата: %s\nВремя: %s\n\nСсылка на Zoom:\n%s%s",
		date,
		timeStr,
		zoomLink,
		nonDiagnosticNotice,
	)
}
