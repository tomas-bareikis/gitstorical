package files

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, errors.Wrap(err, "failed to open dir")
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, errors.Wrap(err, "failed to read dir")
}
