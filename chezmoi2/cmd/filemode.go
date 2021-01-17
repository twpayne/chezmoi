package cmd

import (
	"fmt"
	"os"
	"strconv"
)

// A fileMode represents a file mode. It implements the
// github.com/spf13/pflag.Value interface for use as a command line flag.
type fileMode os.FileMode

func (m *fileMode) FileMode() os.FileMode {
	return os.FileMode(*m)
}

func (m *fileMode) Set(s string) error {
	mUint64, err := strconv.ParseUint(s, 8, 32)
	if err != nil || os.FileMode(mUint64)&os.ModePerm != os.FileMode(mUint64) {
		return fmt.Errorf("%s: invalid mode", s)
	}
	*m = fileMode(mUint64)
	return nil
}

func (m *fileMode) String() string {
	return fmt.Sprintf("%03o", *m)
}

func (m *fileMode) Type() string {
	return "file mode"
}
