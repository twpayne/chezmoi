package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/go-github/v36/github"
	"golang.org/x/oauth2"
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

// englishList returns ss formatted as a list, including an Oxford comma.
func englishList(ss []string) string {
	switch n := len(ss); n {
	case 0:
		return ""
	case 1:
		return ss[0]
	case 2:
		return ss[0] + " and " + ss[1]
	default:
		return strings.Join(ss[:n-1], ", ") + ", and " + ss[n-1]
	}
}

// englishListWithNoun returns ss formatted as an English list, including an Oxford
// comma.
func englishListWithNoun(ss []string, singular, plural string) string {
	if len(ss) == 1 {
		return ss[0] + " " + singular
	}
	if plural == "" {
		plural = pluralize(singular)
	}
	switch n := len(ss); n {
	case 0:
		return "no " + plural
	default:
		return englishList(ss) + " " + plural
	}
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

// newGitHubClient returns a new github.Client configured with an access token,
// if available.
func newGitHubClient(ctx context.Context) *github.Client {
	var httpClient *http.Client
	for _, key := range []string{
		"CHEZMOI_GITHUB_ACCESS_TOKEN",
		"GITHUB_ACCESS_TOKEN",
		"GITHUB_TOKEN",
	} {
		if accessToken := os.Getenv(key); accessToken != "" {
			httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: accessToken,
			}))
			break
		}
	}
	return github.NewClient(httpClient)
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

// pluralize returns the English plural form of singular.
func pluralize(singular string) string {
	if strings.HasSuffix(singular, "y") {
		return strings.TrimSuffix(singular, "y") + "ies"
	}
	return singular + "s"
}

// titleize returns s with its first rune titleized.
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
func upperSnakeCaseToCamelCaseMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
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
