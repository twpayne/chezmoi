package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

// getOSRelease returns the operating system identification data as defined by
// https://www.freedesktop.org/software/systemd/man/os-release.html.
func getOSRelease(fs vfs.FS) (map[string]string, error) {
	for _, filename := range []string{"/usr/lib/os-release", "/etc/os-release"} {
		f, err := fs.Open(filename)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		defer f.Close()
		return parseOSRelease(f)
	}
	return nil, os.ErrNotExist
}

// maybeUnquote removes quotation marks around s.
func maybeUnquote(s string) string {
	// Try to unquote.
	if s, err := strconv.Unquote(s); err == nil {
		return s
	}
	// Otherwise return s, unchanged.
	return s
}

// parseOSRelease parses operating system identification data from r as defined
// by https://www.freedesktop.org/software/systemd/man/os-release.html.
func parseOSRelease(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	s := bufio.NewScanner(r)
	for s.Scan() {
		fields := strings.SplitN(s.Text(), "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("cannot parse %q", s.Text())
		}
		key := fields[0]
		value := maybeUnquote(fields[1])
		result[key] = value
	}
	return result, s.Err()
}
