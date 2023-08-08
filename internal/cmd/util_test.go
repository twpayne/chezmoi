package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestCamelCaseToUpperSnakeCase(t *testing.T) {
	for _, tc := range []struct {
		s        string
		expected string
	}{
		{
			"",
			"",
		},
		{
			"camel",
			"CAMEL",
		},
		{
			"camelCase",
			"CAMEL_CASE",
		},
		{
			"bugReportURL",
			"BUG_REPORT_URL",
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			actual := camelCaseToUpperSnakeCase(tc.s)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestEnglishList(t *testing.T) {
	for _, tc := range []struct {
		ss       []string
		expected string
	}{
		{
			expected: "",
		},
		{
			ss:       []string{"first"},
			expected: "first",
		},
		{
			ss:       []string{"first", "second"},
			expected: "first and second",
		},
		{
			ss:       []string{"first", "second", "third"},
			expected: "first, second, and third",
		},
		{
			ss:       []string{"first", "second", "third", "fourth"},
			expected: "first, second, third, and fourth",
		},
	} {
		actual := englishList(tc.ss)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestEnglishListWithNoun(t *testing.T) {
	for _, tc := range []struct {
		ss       []string
		singular string
		plural   string
		expected string
	}{
		{
			singular: "item",
			expected: "no items",
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
			singular: "entry",
			expected: "no entries",
		},
		{
			ss:       []string{"first"},
			singular: "entry",
			expected: "first entry",
		},
		{
			ss:       []string{"first", "second"},
			singular: "entry",
			expected: "first and second entries",
		},
		{
			ss:       []string{"first", "second", "third"},
			singular: "entry",
			expected: "first, second, and third entries",
		},
		{
			ss:       []string{"first", "second", "third", "fourth"},
			singular: "entry",
			expected: "first, second, third, and fourth entries",
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
		actual := englishListWithNoun(tc.ss, tc.singular, tc.plural)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestUpperSnakeCaseToCamelCaseMap(t *testing.T) {
	actual := upperSnakeCaseToCamelCaseMap(map[string]any{
		"BUG_REPORT_URL": "",
		"ID":             "",
	})
	assert.Equal(t, map[string]any{
		"bugReportURL": "",
		"id":           "",
	}, actual)
}

func Test_flattenStringList(t *testing.T) {
	tests := []struct {
		name   string
		vpaths []any
		want   []string
	}{
		{
			name: "Nothing",
		},
		{
			name:   "Just a string",
			vpaths: []any{"1"},
			want:   []string{"1"},
		},
		{
			name:   "Just a array of string",
			vpaths: []any{[]string{"1", "2"}},
			want:   []string{"1", "2"},
		},
		{
			name:   "Just a array of any containing string",
			vpaths: []any{[]any{"1", "2"}},
			want:   []string{"1", "2"},
		},
		{
			name:   "Just a array of any containing string",
			vpaths: []any{[]any{"1", "2"}},
			want:   []string{"1", "2"},
		},
		{
			name:   "Hybrid",
			vpaths: []any{"0", []any{"1", "2"}, []any{[]string{"3", "4"}}, []any{[]any{"5", "6"}}},
			want:   strings.Split("0123456", ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenStringList(tt.vpaths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}
