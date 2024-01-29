package format

import (
	"encoding/json"
	"fmt"
)

type jsonElement struct {
	Ref    string `json:"ref"`
	Output string `json:"output"`
}

func jsonLines(ref, output string) (string, error) {
	element := jsonElement{
		Ref:    ref,
		Output: output,
	}

	bytes, err := json.Marshal(element)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", bytes), nil
}
