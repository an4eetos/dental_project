package postgres

import (
	"fmt"
	"time"
)

func parsePGTime(raw string) (time.Time, error) {
	layouts := []string{"15:04:05", "15:04:05.999999", "15:04"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", raw)
}
