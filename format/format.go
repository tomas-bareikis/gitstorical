package format

import (
	"fmt"
	"strings"
)

func String(format Format, ref, output string) (string, error) {
	output = strings.TrimSpace(output)

	switch format {
	case JSONL:
		return jsonLines(ref, output)
	case Plain:
		return plain(ref, output), nil
	default:
		return "", fmt.Errorf("unknown format %s", format)
	}
}
