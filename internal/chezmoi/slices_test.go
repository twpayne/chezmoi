package chezmoi_test

import (
	"reflect"
	"testing"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func TestDistinct(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "empty",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "no duplicates",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with duplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "all duplicates",
			input: []string{"a", "a", "a"},
			want:  []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chezmoi.Distinct(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Distinct(%v) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}
