package cmd

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dustin/go-humanize/english"

	"chezmoi.io/chezmoi/v2/internal/chezmoiset"
)

var (
	wellKnownAbbreviations = chezmoiset.New(
		"ANSI",
		"CPE",
		"ID",
		"URL",
	)

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
	choicesOverwrite = []string{
		"overwrite",
		"all-overwrite",
		"skip",
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
	words := make([]string, len(wordBoundaries))
	prevWordBoundary := 0
	for i, wordBoundary := range wordBoundaries {
		word := string(runes[prevWordBoundary:wordBoundary])
		words[i] = strings.ToUpper(word)
		prevWordBoundary = wordBoundary
	}
	return strings.Join(words, "_")
}

// englishListWithNoun returns ss formatted as an English list, including an
// Oxford comma.
func englishListWithNoun(ss []string, singular, plural string) string {
	switch n := len(ss); n {
	case 0:
		return "no " + english.PluralWord(0, singular, plural)
	case 1:
		return ss[0] + " " + singular
	default:
		return english.OxfordWordSeries(ss, "and") + " " + english.PluralWord(n, singular, plural)
	}
}

// stringersToStrings converts a slice of fmt.Stringers to a list of strings.
func stringersToStrings[T fmt.Stringer](ss []T) []string {
	result := make([]string, len(ss))
	for i, s := range ss {
		result[i] = s.String()
	}
	return result
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
		} else if !wellKnownAbbreviations.Contains(word) {
			words[i] = titleFirst(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// upperSnakeCaseToCamelCaseMap returns m with all keys converted from
// UPPER_SNAKE_CASE to camelCase.
func upperSnakeCaseToCamelCaseMap[V any](m map[string]V) map[string]V {
	result := make(map[string]V)
	for k, v := range m {
		result[upperSnakeCaseToCamelCase(k)] = v
	}
	return result
}
