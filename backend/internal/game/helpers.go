package game

import "encoding/json"

func marshalJSON(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}"), err
	}
	return b, nil
}
