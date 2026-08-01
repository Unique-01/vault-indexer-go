package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Cursor struct {
	BlockNumber uint64
	LogIndex    uint
}

func EncodeCursor(c Cursor) string {
	raw := fmt.Sprintf("%d:%d", c.BlockNumber, c.LogIndex)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(s string) (Cursor, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("Invalid cursor encoding: %w", err)
	}
	var (
		blockNumber uint64
		logIndex    uint
	)
	_, err = fmt.Sscanf(string(raw), "%d:%d", &blockNumber, &logIndex)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor format: %w", err)
	}

	return Cursor{BlockNumber: blockNumber, LogIndex: logIndex}, nil
}

func (c Cursor) MarshalJSON() ([]byte, error) {
	return json.Marshal(EncodeCursor(c))
}
