package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	vfs "github.com/twpayne/go-vfs"
)

var (
	wellKnownAbbreviations = map[string]struct{}{
		"ANSI": struct{}{},
		"CPE":  struct{}{},
		"URL":  struct{}{},
	}
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
		m, err := parseOSRelease(f)
		if err != nil {
			return nil, err
		}
		return upperSnakeCaseToCamelCaseMap(m), nil
	}
	return nil, os.ErrNotExist
}

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
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

// titilize returns s, titilized.
func titilize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	return string(append([]rune{unicode.ToTitle(runes[0])}, runes[1:]...))
}

// upperSnakeCaseToCamelCase converts a string in UPPER_SNAKE_CASE to
// camelCase.
func upperSnakeCaseToCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else if !isWellKnownAbbreviation(word) {
			words[i] = titilize(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// upperSnakeCaseToCamelCaseKeys returns m with all keys converted from UPPER_SNAKE_CASE to camelCase.
func upperSnakeCaseToCamelCaseMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[upperSnakeCaseToCamelCase(k)] = v
	}
	return result
}
