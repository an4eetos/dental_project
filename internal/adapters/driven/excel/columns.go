package excel

import (
	"strconv"
	"strings"
)

// isIgnorableColumn reports whether a header names a row-index column that should be skipped.
func isIgnorableColumn(header string) bool {
	key := strings.TrimSpace(header)
	if key == "" {
		return true
	}

	lower := strings.ToLower(key)
	switch lower {
	case "№", "#", "n", "no", "no.", "num", "number", "номер", "н", "н.":
		return true
	}

	if strings.HasPrefix(lower, "unnamed:") {
		return true
	}

	if _, err := strconv.Atoi(key); err == nil {
		return true
	}

	return false
}
