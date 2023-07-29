package chezmoi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/twpayne/go-vfs/v4"
)

// Kernel returns the kernel information parsed from /proc/sys/kernel.
func Kernel(fileSystem vfs.FS) (map[string]any, error) {
	const procSysKernel = "/proc/sys/kernel"

	switch fileInfo, err := fileSystem.Stat(procSysKernel); {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case errors.Is(err, fs.ErrPermission):
		return nil, nil
	case err != nil:
		return nil, err
	case !fileInfo.Mode().IsDir():
		return nil, nil
	}

	kernel := make(map[string]any)
	for _, filename := range []string{
		"osrelease",
		"ostype",
		"version",
	} {
		switch data, err := fileSystem.ReadFile(filepath.Join(procSysKernel, filename)); {
		case err == nil:
			kernel[filename] = string(bytes.TrimSpace(data))
		case errors.Is(err, fs.ErrNotExist):
			continue
		case errors.Is(err, fs.ErrPermission):
			continue
		default:
			return nil, err
		}
	}
	return kernel, nil
}

// OSRelease returns the operating system identification data as defined by the
// os-release specification.
func OSRelease(fileSystem vfs.FS) (map[string]any, error) {
	for _, filename := range []string{
		"/etc/os-release",
		"/usr/lib/os-release",
	} {
		data, err := fileSystem.ReadFile(filename)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		m, err := parseOSRelease(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	return nil, fs.ErrNotExist
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
// by the os-release specification.
func parseOSRelease(r io.Reader) (map[string]any, error) {
	result := make(map[string]any)
	s := bufio.NewScanner(r)
	for s.Scan() {
		// Trim all leading whitespace, but not necessarily trailing whitespace.
		token := strings.TrimLeftFunc(s.Text(), unicode.IsSpace)
		// If the line is empty or starts with #, skip.
		if len(token) == 0 || token[0] == '#' {
			continue
		}
		key, value, ok := strings.Cut(token, "=")
		if !ok {
			return nil, fmt.Errorf("%s: parse error", token)
		}
		result[key] = maybeUnquote(value)
	}
	return result, s.Err()
}
