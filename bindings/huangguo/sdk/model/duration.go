package model

import (
	"encoding/json"
	"strconv"
)

// Detail's duration is a JSON number on older endpoints and a numeric string on
// the current endpoint. Keep the public seconds value and canonical numeric JSON.
func (d *Detail) UnmarshalJSON(data []byte) error {
	type plain Detail
	var decoded plain
	value := struct {
		*plain
		Duration json.RawMessage `json:"duration"`
	}{plain: &decoded}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if len(value.Duration) > 0 && string(value.Duration) != "null" {
		raw := string(value.Duration)
		if raw[0] == '"' {
			if err := json.Unmarshal(value.Duration, &raw); err != nil {
				return err
			}
		}
		duration, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		decoded.Duration = duration
	}
	*d = Detail(decoded)
	return nil
}
