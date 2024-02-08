package format

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

func String(format Format, ref plumbing.ReferenceName, output string) (string, error) {
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
