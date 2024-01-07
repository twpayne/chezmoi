package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestStringInputModel(t *testing.T) {
	for _, tc := range []struct {
		defaultValue     *string
		name             string
		input            string
		expectedValue    string
		expectedCanceled bool
	}{
		{
			name:  "empty",
			input: "\r",
		},
		{
			name:          "empty_with_default",
			defaultValue:  newString("default"),
			input:         "\r",
			expectedValue: "default",
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
			name:          "value_enter",
			input:         "value\r",
			expectedValue: "value",
		},
		{
			name:             "value_ctrlc",
			input:            "value\x03",
			expectedCanceled: true,
			expectedValue:    "value",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualModel := testRunModelWithInput(t, NewStringInputModel("prompt", tc.defaultValue), tc.input)
			assert.Equal(t, tc.expectedCanceled, actualModel.Canceled())
			assert.Equal(t, tc.expectedValue, actualModel.Value())
		})
	}
}
