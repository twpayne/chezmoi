package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptBool(t *testing.T) {
	for _, tc := range []struct {
		name              string
		prompt            string
		args              []bool
		stdinStr          string
		expectedStdoutStr string
		expected          bool
		expectedErr       bool
	}{
		{
			name:              "response_without_default",
			prompt:            "bool",
			stdinStr:          "false\n",
			expectedStdoutStr: "bool? ",
			expected:          false,
		},
		{
			name:              "response_with_default",
			prompt:            "bool",
			args:              []bool{true},
			stdinStr:          "no\n",
			expectedStdoutStr: "bool (default true)? ",
			expected:          false,
		},
		{
			name:              "no_response_with_default",
			prompt:            "bool",
			args:              []bool{true},
			stdinStr:          "\n",
			expectedStdoutStr: "bool (default true)? ",
			expected:          true,
		},
		{
			name:        "invalid_response",
			stdinStr:    "invalid\n",
			expectedErr: true,
		},
		{
			name:        "invalid_response_with_default",
			args:        []bool{false},
			stdinStr:    "invalid\n",
			expectedErr: true,
		},
		{
			name:        "too_many_args",
			prompt:      "bool",
			args:        []bool{false, false},
			stdinStr:    "\n",
			expectedErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdin := strings.NewReader(tc.stdinStr)
			stdout := &strings.Builder{}
			config, err := newConfig(
				withStdin(stdin),
				withStdout(stdout),
			)
			require.NoError(t, err)
			if tc.expectedErr {
				assert.Panics(t, func() {
					config.promptBool(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptBool(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}
