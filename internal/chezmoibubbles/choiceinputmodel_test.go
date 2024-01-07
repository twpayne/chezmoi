package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestChoiceInputModel(t *testing.T) {
	choicesYesNoAll := []string{"yes", "no", "all"}
	for _, tc := range []struct {
		defaultValue     *string
		name             string
		input            string
		expectedValue    string
		choices          []string
		expectedCanceled bool
	}{
		{
			name:          "empty_with_default",
			choices:       choicesYesNoAll,
			defaultValue:  newString("all"),
			input:         "\r",
			expectedValue: "all",
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
			name:          "y",
			choices:       choicesYesNoAll,
			input:         "y",
			expectedValue: "yes",
		},
		{
			name:          "n",
			choices:       choicesYesNoAll,
			input:         "n",
			expectedValue: "no",
		},
		{
			name:          "a",
			choices:       choicesYesNoAll,
			input:         "a",
			expectedValue: "all",
		},
		{
			name:    "ambiguous_a",
			choices: []string{"aaa", "abb", "bbb"},
			input:   "a",
		},
		{
			name:          "unambiguous_b",
			choices:       []string{"aaa", "abb", "bbb"},
			input:         "b",
			expectedValue: "bbb",
		},
		{
			name:          "ambiguous_resolved",
			choices:       []string{"aaa", "abb", "bbb"},
			input:         "aa",
			expectedValue: "aaa",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualModel := testRunModelWithInput(t, NewChoiceInputModel("prompt", tc.choices, tc.defaultValue), tc.input)
			assert.Equal(t, tc.expectedCanceled, actualModel.Canceled())
			assert.Equal(t, tc.expectedValue, actualModel.Value())
		})
	}
}
