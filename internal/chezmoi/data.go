package chezmoi

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"

	"github.com/twpayne/go-vfs/v5"
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
		m, err := parseOSRelease(data)
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
func parseOSRelease(data []byte) (map[string]any, error) {
	result := make(map[string]any)
	for line := range bytes.Lines(data) {
		token := bytes.TrimSpace(line)
		if len(token) == 0 || token[0] == '#' {
			continue
		}
		key, value, ok := bytes.Cut(token, []byte{'='})
		if !ok {
			return nil, fmt.Errorf("%s: parse error", token)
		}
		result[string(key)] = maybeUnquote(string(value))
	}
	return result, nil
}
