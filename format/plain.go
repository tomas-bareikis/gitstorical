package format

import (
	"fmt"
)

func plain(ref, output string) string {
	return fmt.Sprintf("%s\n%s\n", ref, output)
}
