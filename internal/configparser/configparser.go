package configparser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/twpayne/go-vfs"
)

type parser func(io.Reader, interface{}) error

var parsers = make(map[string]parser)

// Extensions returns all the supported extensions in alphabetical order.
func Extensions() []string {
	extensions := make([]string, 0, len(parsers))
	for extension := range parsers {
		extensions = append(extensions, extension)
	}
	sort.Strings(extensions)
	return extensions
}

// FindConfig finds the first config file named filename or with basename
// filename.
func FindConfig(fs vfs.Stater, filename string) (string, error) {
	info, err := fs.Stat(filename)
	switch {
	case err == nil && info.Mode().IsRegular():
		return filename, err
	case !os.IsNotExist(err):
		return "", err
	}

	for _, extension := range Extensions() {
		info, err := fs.Stat(filename + extension)
		switch {
		case err == nil && info.Mode().IsRegular():
			return filename + extension, err
		case !os.IsNotExist(err):
			return "", err
		}
	}

	return "", nil
}

// ParseConfig parses r into value.
func ParseConfig(f *os.File, value interface{}) error {
	if f == nil {
		return nil
	}
	parser, ok := parsers[filepath.Ext(f.Name())]
	if !ok {
		return fmt.Errorf("%s: unsupported format", f.Name())
	}
	return parser(f, value)
}
