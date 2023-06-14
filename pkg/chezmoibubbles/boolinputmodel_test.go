package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBoolInputModel(t *testing.T) {
	for _, tc := range []struct {
		name             string
		defaultValue     *bool
		input            string
		expectedCanceled bool
		expectedValue    bool
	}{
		{
			name:          "empty_with_default",
			defaultValue:  newBool(true),
			input:         "\r",
			expectedValue: true,
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
			name:          "true",
			input:         "t",
			expectedValue: true,
		},
		{
			name:          "false",
			input:         "f",
			expectedValue: false,
		},
		{
			name:          "yes",
			input:         "y",
			expectedValue: true,
		},
		{
			name:          "no",
			input:         "n",
			expectedValue: false,
		},
		{
			name:          "one",
			input:         "1",
			expectedValue: true,
		},
		{
			name:          "zero",
			input:         "0",
			expectedValue: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualModel := testRunModelWithInput(
				t,
				NewBoolInputModel("prompt", tc.defaultValue),
				tc.input,
			)
			assert.Equal(t, tc.expectedCanceled, actualModel.Canceled())
			assert.Equal(t, tc.expectedValue, actualModel.Value())
		})
	}
}
