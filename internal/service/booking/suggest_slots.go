package booking

import (
	"fmt"
	"strings"
	"time"
)

const suggestSlotCount = 6

var suggestWeekdays = []time.Weekday{
	time.Monday,
	time.Tuesday,
	time.Wednesday,
	time.Thursday,
	time.Friday,
}

var suggestHours = []int{10, 12, 14, 16, 18}

// SuggestAvailableSlots builds a Russian text block with upcoming clinic slots.
func SuggestAvailableSlots(now time.Time, patientDate time.Time) string {
	lines := make([]string, 0, suggestSlotCount+1)
	lines = append(lines, "Доступные варианты для записи:")

	cursor := truncateUTC(now.UTC())
	if patientDate.After(cursor) {
		cursor = patientDate
	}

	added := 0
	for dayOffset := 1; added < suggestSlotCount && dayOffset <= MaxDaysAhead; dayOffset++ {
		day := cursor.AddDate(0, 0, dayOffset)
		if !isSuggestWeekday(day.Weekday()) {
			continue
		}
		for _, hour := range suggestHours {
			if added >= suggestSlotCount {
				break
			}
			lines = append(lines, fmt.Sprintf("- %s в %02d:00", day.Format(DateLayout), hour))
			added++
		}
	}

	if added == 0 {
		return "Доступные варианты для записи:\n- Свяжитесь с клиникой для согласования времени."
	}
	return strings.Join(lines, "\n")
}

func isSuggestWeekday(wd time.Weekday) bool {
	for _, allowed := range suggestWeekdays {
		if wd == allowed {
			return true
		}
	}
	return false
}
