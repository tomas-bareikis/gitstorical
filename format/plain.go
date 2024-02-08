package format

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
)

func plain(ref plumbing.ReferenceName, output string) string {
	return fmt.Sprintf("%s\n%s\n", ref.Short(), output)
}
