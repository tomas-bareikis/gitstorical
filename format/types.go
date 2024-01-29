package format

import (
	"fmt"
	"strings"
)

type Format string

const (
	Plain Format = "plain"
	JSONL Format = "jsonl"
)

func ParseType(s string) (Format, error) {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case string(Plain):
		return Plain, nil
	case string(JSONL):
		return JSONL, nil
	default:
		return Plain, fmt.Errorf("cannot parse format %s", s)
	}
}
