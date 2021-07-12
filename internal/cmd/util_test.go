package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnglishList(t *testing.T) {
	for _, tc := range []struct {
		ss       []string
		singular string
		plural   string
		expected string
	}{
		{
			singular: "item",
			expected: "zero items",
		},
		{
			ss:       []string{"first"},
			singular: "item",
			expected: "first item",
		},
		{
			ss:       []string{"first", "second"},
			singular: "item",
			expected: "first and second items",
		},
		{
			ss:       []string{"first", "second", "third"},
			singular: "item",
			expected: "first, second, and third items",
		},
		{
			ss:       []string{"first", "second", "third", "fourth"},
			singular: "item",
			expected: "first, second, third, and fourth items",
		},
		{
			ss:       []string{"first"},
			singular: "person",
			plural:   "people",
			expected: "first person",
		},
		{
			ss:       []string{"first", "second", "third"},
			singular: "person",
			plural:   "people",
			expected: "first, second, and third people",
		},
	} {
		actual := englishList(tc.ss, tc.singular, tc.plural)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestUniqueAbbreviations(t *testing.T) {
	for _, tc := range []struct {
		values   []string
		expected map[string]string
	}{
		{
			values:   nil,
			expected: map[string]string{},
		},
		{
			values: []string{
				"yes",
				"no",
				"all",
				"quit",
			},
			expected: map[string]string{
				"y":    "yes",
				"ye":   "yes",
				"yes":  "yes",
				"n":    "no",
				"no":   "no",
				"a":    "all",
				"al":   "all",
				"all":  "all",
				"q":    "quit",
				"qu":   "quit",
				"qui":  "quit",
				"quit": "quit",
			},
		},
		{
			values: []string{
				"ale",
				"all",
				"abort",
			},
			expected: map[string]string{
				"ale":   "ale",
				"all":   "all",
				"ab":    "abort",
				"abo":   "abort",
				"abor":  "abort",
				"abort": "abort",
			},
		},
		{
			values: []string{
				"no",
				"now",
				"nope",
			},
			expected: map[string]string{
				"no":   "no",
				"now":  "now",
				"nop":  "nope",
				"nope": "nope",
			},
		},
	} {
		t.Run(strings.Join(tc.values, "_"), func(t *testing.T) {
			actual := uniqueAbbreviations(tc.values)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestUpperSnakeCaseToCamelCaseMap(t *testing.T) {
	actual := upperSnakeCaseToCamelCaseMap(map[string]interface{}{
		"BUG_REPORT_URL": "",
		"ID":             "",
	})
	assert.Equal(t, map[string]interface{}{
		"bugReportURL": "",
		"id":           "",
	}, actual)
}
