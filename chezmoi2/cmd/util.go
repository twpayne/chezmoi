package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/viper"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-xdg/v3"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

var (
	wellKnownAbbreviations = map[string]struct{}{
		"ANSI": {},
		"CPE":  {},
		"ID":   {},
		"URL":  {},
	}

	yesNoAllQuit = []string{
		"yes",
		"no",
		"all",
		"quit",
	}
)

// defaultConfigFile returns the default config file according to the XDG Base
// Directory Specification.
func defaultConfigFile(fs vfs.Stater, bds *xdg.BaseDirectorySpecification) chezmoi.AbsPath {
	// Search XDG Base Directory Specification config directories first.
	for _, configDir := range bds.ConfigDirs {
		configDirAbsPath := chezmoi.AbsPath(configDir)
		for _, extension := range viper.SupportedExts {
			configFileAbsPath := configDirAbsPath.Join(chezmoi.RelPath("chezmoi"), chezmoi.RelPath("chezmoi."+extension))
			if _, err := fs.Stat(string(configFileAbsPath)); err == nil {
				return configFileAbsPath
			}
		}
	}
	// Fallback to XDG Base Directory Specification default.
	return chezmoi.AbsPath(bds.ConfigHome).Join(chezmoi.RelPath("chezmoi"), chezmoi.RelPath("chezmoi.toml"))
}

// defaultSourceDir returns the default source directory according to the XDG
// Base Directory Specification.
func defaultSourceDir(fs vfs.Stater, bds *xdg.BaseDirectorySpecification) chezmoi.AbsPath {
	// Check for XDG Base Directory Specification data directories first.
	for _, dataDir := range bds.DataDirs {
		dataDirAbsPath := chezmoi.AbsPath(dataDir)
		sourceDirAbsPath := dataDirAbsPath.Join(chezmoi.RelPath("chezmoi"))
		if _, err := fs.Stat(string(sourceDirAbsPath)); err == nil {
			return sourceDirAbsPath
		}
	}
	// Fallback to XDG Base Directory Specification default.
	return chezmoi.AbsPath(bds.DataHome).Join(chezmoi.RelPath("chezmoi"))
}

// firstNonEmptyString returns its first non-empty argument, or "" if all
// arguments are empty.
func firstNonEmptyString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
}

// parseBool is like strconv.ParseBool but also accepts on, ON, y, Y, yes, YES,
// n, N, no, NO, off, and OFF.
func parseBool(str string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "n", "no", "off":
		return false, nil
	case "on", "y", "yes":
		return true, nil
	default:
		return strconv.ParseBool(str)
	}
}

// serializationFormatNamesStr returns the list of serialization formats as a
// comma-separated list.
func serializationFormatNamesStr() string {
	names := make([]string, 0, len(chezmoi.Formats))
	for name := range chezmoi.Formats {
		names = append(names, name)
	}
	sort.Strings(names)
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " or " + names[1]
	default:
		names[len(names)-1] = "or " + names[len(names)-1]
		return strings.Join(names, ", ")
	}
}

// titleize returns s with its first rune titlized.
func titleize(s string) string {
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
			words[i] = titleize(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// uniqueAbbreviations returns a map of unique abbreviations of values to
// values. Values always map to themselves.
func uniqueAbbreviations(values []string) map[string]string {
	abbreviations := make(map[string][]string)
	for _, value := range values {
		for i := 1; i <= len(value); i++ {
			abbreviation := value[:i]
			abbreviations[abbreviation] = append(abbreviations[abbreviation], value)
		}
	}
	uniqueAbbreviations := make(map[string]string)
	for abbreviation, values := range abbreviations {
		if len(values) == 1 {
			uniqueAbbreviations[abbreviation] = values[0]
		}
	}
	for _, value := range values {
		uniqueAbbreviations[value] = value
	}
	return uniqueAbbreviations
}

// upperSnakeCaseToCamelCaseKeys returns m with all keys converted from
// UPPER_SNAKE_CASE to camelCase.
func upperSnakeCaseToCamelCaseMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[upperSnakeCaseToCamelCase(k)] = v
	}
	return result
}

// validateKeys ensures that all keys in data match re.
func validateKeys(data interface{}, re *regexp.Regexp) error {
	switch data := data.(type) {
	case map[string]interface{}:
		for key, value := range data {
			if !re.MatchString(key) {
				return fmt.Errorf("%s: invalid key", key)
			}
			if err := validateKeys(value, re); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, value := range data {
			if err := validateKeys(value, re); err != nil {
				return err
			}
		}
	}
	return nil
}
