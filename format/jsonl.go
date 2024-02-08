package format

import (
	"encoding/json"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
)

type jsonElement struct {
	Ref    string `json:"ref"`
	Output string `json:"output"`
}

func jsonLines(ref plumbing.ReferenceName, output string) (string, error) {
	element := jsonElement{
		Ref:    ref.Short(),
		Output: output,
	}

	bytes, err := json.Marshal(element)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", bytes), nil
}
