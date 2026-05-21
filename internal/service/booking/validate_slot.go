package booking

import (
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

const (
	MinHour        = 9
	MaxHour        = 20
	MaxDaysAhead   = 90
	TimeLayout     = "15:04"
	DateLayout     = "2006-01-02"
)

// ParsePreferredDate parses YYYY-MM-DD and validates it is today or later within MaxDaysAhead.
func ParsePreferredDate(raw string, now time.Time) (time.Time, error) {
	t, err := time.ParseInLocation(DateLayout, raw, time.UTC)
	if err != nil {
		return time.Time{}, domainerrors.ErrInvalidPreferredDate
	}
	day := truncateUTC(t)
	today := truncateUTC(now.UTC())
	maxDay := today.AddDate(0, 0, MaxDaysAhead)
	if day.Before(today) || day.After(maxDay) {
		return time.Time{}, domainerrors.ErrInvalidPreferredDate
	}
	return day, nil
}

// ParsePreferredTime parses HH:MM (24h) and validates clinic hours.
func ParsePreferredTime(raw string) (time.Time, error) {
	parsed, err := time.Parse(TimeLayout, raw)
	if err != nil {
		return time.Time{}, domainerrors.ErrInvalidPreferredTime
	}
	hour := parsed.Hour()
	minute := parsed.Minute()
	if hour < MinHour || hour > MaxHour {
		return time.Time{}, domainerrors.ErrInvalidPreferredTime
	}
	if hour == MaxHour && minute > 0 {
		return time.Time{}, domainerrors.ErrInvalidPreferredTime
	}
	if minute < 0 || minute > 59 {
		return time.Time{}, domainerrors.ErrInvalidPreferredTime
	}
	return time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC), nil
}

func truncateUTC(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
