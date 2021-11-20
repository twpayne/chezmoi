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
func Kernel(fileSystem vfs.FS) (map[string]interface{}, error) {
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

	kernel := make(map[string]interface{})
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
func OSRelease(system System) (map[string]interface{}, error) {
	for _, filename := range []AbsPath{
		NewAbsPath("/etc/os-release"),
		NewAbsPath("/usr/lib/os-release"),
	} {
		data, err := system.ReadFile(filename)
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
func parseOSRelease(r io.Reader) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	s := bufio.NewScanner(r)
	for s.Scan() {
		// Trim all leading whitespace, but not necessarily trailing whitespace.
		token := strings.TrimLeftFunc(s.Text(), unicode.IsSpace)
		// If the line is empty or starts with #, skip.
		if len(token) == 0 || token[0] == '#' {
			continue
		}
		fields := strings.SplitN(token, "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("%s: parse error", token)
		}
		key := fields[0]
		value := maybeUnquote(fields[1])
		result[key] = value
	}
	return result, s.Err()
}
