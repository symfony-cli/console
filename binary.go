package console

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func CurrentBinaryName() string {
	argv0, err := os.Executable()
	if nil != err {
		return ""
	}

	return filepath.Base(argv0)
}

func CurrentBinaryPath() (string, error) {
	argv0, err := os.Executable()
	if nil != err {
		return argv0, errors.WithStack(err)
	}
	return argv0, nil
}
