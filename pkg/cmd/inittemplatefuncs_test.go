package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPromptBoolInitTemplateFunc(t *testing.T) {
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
				withNoTTY(true),
				withStdin(stdin),
				withStdout(stdout),
			)
			assert.NoError(t, err)
			if tc.expectedErr {
				assert.Panics(t, func() {
					config.promptBoolInitTemplateFunc(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptBoolInitTemplateFunc(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}

func TestPromptIntInitTemplateFunc(t *testing.T) {
	for _, tc := range []struct {
		name              string
		prompt            string
		args              []int64
		stdinStr          string
		expectedStdoutStr string
		expected          int64
		expectedErr       bool
	}{
		{
			name:              "response_without_default",
			prompt:            "int",
			stdinStr:          "1\n",
			expectedStdoutStr: "int? ",
			expected:          1,
		},
		{
			name:              "response_with_default",
			prompt:            "int",
			args:              []int64{1},
			stdinStr:          "2\n",
			expectedStdoutStr: "int (default 1)? ",
			expected:          2,
		},
		{
			name:              "no_response_with_default",
			prompt:            "int",
			args:              []int64{1},
			stdinStr:          "\n",
			expectedStdoutStr: "int (default 1)? ",
			expected:          1,
		},
		{
			name:        "invalid_response",
			stdinStr:    "invalid\n",
			expectedErr: true,
		},
		{
			name:        "invalid_response_with_default",
			args:        []int64{1},
			stdinStr:    "invalid\n",
			expectedErr: true,
		},
		{
			name:        "too_many_args",
			prompt:      "bool",
			args:        []int64{0, 0},
			stdinStr:    "\n",
			expectedErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdin := strings.NewReader(tc.stdinStr)
			stdout := &strings.Builder{}
			config, err := newConfig(
				withNoTTY(true),
				withStdin(stdin),
				withStdout(stdout),
			)
			assert.NoError(t, err)
			if tc.expectedErr {
				assert.Panics(t, func() {
					config.promptIntInitTemplateFunc(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptIntInitTemplateFunc(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}

func TestPromptStringInitTemplateFunc(t *testing.T) {
	for _, tc := range []struct {
		name              string
		prompt            string
		args              []string
		stdinStr          string
		expectedStdoutStr string
		expected          string
		expectedErr       bool
	}{
		{
			name:              "response_without_default",
			prompt:            "string",
			stdinStr:          "one\n",
			expectedStdoutStr: "string? ",
			expected:          "one",
		},
		{
			name:              "response_with_default",
			prompt:            "string",
			args:              []string{"one"},
			stdinStr:          "two\n",
			expectedStdoutStr: `string (default "one")? `,
			expected:          "two",
		},
		{
			name:              "response_with_space_with_default",
			prompt:            "string",
			args:              []string{"one"},
			stdinStr:          " two \n",
			expectedStdoutStr: `string (default "one")? `,
			expected:          "two",
		},
		{
			name:              "no_response_with_default_with_space",
			prompt:            "string",
			args:              []string{" one "},
			stdinStr:          "\n",
			expectedStdoutStr: `string (default "one")? `,
			expected:          "one",
		},
		{
			name:              "no_response_with_default",
			prompt:            "string",
			args:              []string{"one"},
			stdinStr:          "\n",
			expectedStdoutStr: `string (default "one")? `,
			expected:          "one",
		},
		{
			name:              "whitespace_response_with_default",
			prompt:            "string",
			args:              []string{"one"},
			stdinStr:          " \r\n",
			expectedStdoutStr: `string (default "one")? `,
			expected:          "one",
		},
		{
			name:        "too_many_args",
			prompt:      "bool",
			args:        []string{"", ""},
			stdinStr:    "\n",
			expectedErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdin := strings.NewReader(tc.stdinStr)
			stdout := &strings.Builder{}
			config, err := newConfig(
				withNoTTY(true),
				withStdin(stdin),
				withStdout(stdout),
			)
			assert.NoError(t, err)
			if tc.expectedErr {
				assert.Panics(t, func() {
					config.promptStringInitTemplateFunc(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptStringInitTemplateFunc(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}
