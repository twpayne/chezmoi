package cmd

import (
	"strings"
	"unicode"
)

var (
	wellKnownAbbreviations = map[string]struct{}{
		"ANSI": {},
		"CPE":  {},
		"ID":   {},
		"URL":  {},
	}

	choicesYesNoAllQuit = []string{
		"yes",
		"no",
		"all",
		"quit",
	}
	choicesYesNoQuit = []string{
		"yes",
		"no",
		"quit",
	}
)

// camelCaseToUpperSnakeCase converts a string in camelCase to UPPER_SNAKE_CASE.
func camelCaseToUpperSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	var wordBoundaries []int
	for i, r := range runes[1:] {
		if unicode.IsLower(runes[i]) && unicode.IsUpper(r) {
			wordBoundaries = append(wordBoundaries, i+1)
		}
	}

	if len(wordBoundaries) == 0 {
		return strings.ToUpper(s)
	}

	wordBoundaries = append(wordBoundaries, len(runes))
	words := make([]string, 0, len(wordBoundaries))
	prevWordBoundary := 0
	for _, wordBoundary := range wordBoundaries {
		word := string(runes[prevWordBoundary:wordBoundary])
		words = append(words, strings.ToUpper(word))
		prevWordBoundary = wordBoundary
	}
	return strings.Join(words, "_")
}

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

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
}

// pluralize returns the English plural form of singular.
func pluralize(singular string) string {
	if strings.HasSuffix(singular, "y") {
		return strings.TrimSuffix(singular, "y") + "ies"
	}
	return singular + "s"
}

// titleFirst returns s with its first rune converted to title case.
func titleFirst(s string) string {
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
			words[i] = titleFirst(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// upperSnakeCaseToCamelCaseKeys returns m with all keys converted from
// UPPER_SNAKE_CASE to camelCase.
func upperSnakeCaseToCamelCaseMap[V any](m map[string]V) map[string]V {
	result := make(map[string]V)
	for k, v := range m {
		result[upperSnakeCaseToCamelCase(k)] = v
	}
	return result
}
