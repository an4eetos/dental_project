package prediction

import (
	"encoding/json"
	"strings"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
)

// ParseOutputJSON decodes a Gemini JSON response into an output row.
func ParseOutputJSON(raw string) (domain.Row, error) {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var decoded map[string]string
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		return nil, err
	}

	out := domain.NewOutputRow()
	for _, key := range domain.OutputKeys() {
		value, ok := decoded[key]
		if !ok {
			return nil, errMissingOutputKey(key)
		}
		out[key] = strings.TrimSpace(value)
		if out[key] == "" {
			return nil, errMissingOutputKey(key)
		}
	}

	return out, nil
}

type missingOutputKeyError struct {
	key string
}

func (e missingOutputKeyError) Error() string {
	return "missing output key: " + e.key
}

func errMissingOutputKey(key string) error {
	return missingOutputKeyError{key: key}
}
