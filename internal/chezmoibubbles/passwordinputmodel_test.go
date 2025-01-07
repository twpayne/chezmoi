package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPasswordInputModel(t *testing.T) {
	for _, tc := range []struct {
		name             string
		input            string
		expectedCanceled bool
		expectedValue    string
	}{
		{
			name:  "empty",
			input: "\r",
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
			name:          "password_enter",
			input:         "password\r",
			expectedValue: "password",
		},
		{
			name:             "password_ctrlc",
			input:            "password\x03",
			expectedCanceled: true,
			expectedValue:    "password",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualModel := testRunModelWithInput(t, NewPasswordInputModel("prompt", "placeholder"), tc.input)
			assert.Equal(t, tc.expectedCanceled, actualModel.Canceled())
			assert.Equal(t, tc.expectedValue, actualModel.Value())
		})
	}
}
