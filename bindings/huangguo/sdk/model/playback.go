package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func (p *PlayInfo) UnmarshalJSON(data []byte) error {
	type plain PlayInfo
	var decoded plain
	value := struct {
		*plain
		ID  json.RawMessage `json:"id"`
		EP  int             `json:"ep"`
		CDN struct {
			Enabled   bool   `json:"m3u8Enabled"`
			AuthKey   string `json:"authKey"`
			AuthValue string `json:"authValue"`
			App       string `json:"app"`
		} `json:"cdn"`
	}{plain: &decoded}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if len(value.ID) > 0 {
		raw := string(value.ID)
		if raw[0] == '"' {
			if err := json.Unmarshal(value.ID, &raw); err != nil {
				return err
			}
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playback identity: %w", err)
		}
		decoded.ID = id
	}
	if decoded.EpNum == 0 {
		decoded.EpNum = value.EP
	}
	if value.CDN.Enabled && value.CDN.AuthKey != "" && value.CDN.AuthValue != "" && value.CDN.App != "" {
		if decoded.MediaHeaders == nil {
			decoded.MediaHeaders = map[string]string{}
		}
		decoded.MediaHeaders[value.CDN.AuthKey] = value.CDN.AuthValue
		decoded.MediaHeaders["x-cdn-app"] = value.CDN.App
	}
	*p = PlayInfo(decoded)
	return nil
}
