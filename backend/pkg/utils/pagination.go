package utils

import (
	"encoding/base64"
	"encoding/json"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	Limit      int    `json:"limit"`
}

func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func EncodeCursor(payload any) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(raw), nil
}

func DecodeCursor(cursor string, dst any) error {
	raw, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}
