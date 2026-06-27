package admin

import "encoding/json"

func marshalBlockData(data any) (json.RawMessage, error) {
	if data == nil {
		return json.RawMessage("{}"), nil
	}
	return json.Marshal(data)
}
