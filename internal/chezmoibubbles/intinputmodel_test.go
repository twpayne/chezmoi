package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestIntInputModel(t *testing.T) {
	for _, tc := range []struct {
		name             string
		defaultValue     *int64
		input            string
		expectedCanceled bool
		expectedValue    int64
	}{
		{
			name:          "empty_with_default",
			defaultValue:  newInt64(1),
			input:         "\r",
			expectedValue: 1,
		},
		{
			name:             "cancel_ctrlc",
			input:            "\x03",
			expectedCanceled: true,
		},
		{
			name:             "cancel_esc",
			input:            "\x1b",
			expectedCanceled: true,
		},
		{
			name:          "one_enter",
			input:         "1\r",
			expectedValue: 1,
		},
		{
			name:          "minus_one_enter",
			input:         "-1\r",
			expectedValue: -1,
		},
		{
			name:          "minus_enter",
			input:         "-\r",
			expectedValue: 0,
		},
		{
			name:          "one_invalid_enter",
			input:         "1a\r",
			expectedValue: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualModel := testRunModelWithInput(t, NewIntInputModel("prompt", tc.defaultValue), tc.input)
			assert.Equal(t, tc.expectedCanceled, actualModel.Canceled())
			assert.Equal(t, tc.expectedValue, actualModel.Value())
		})
	}
}
