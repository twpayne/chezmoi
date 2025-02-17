package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPromptBoolInteractiveTemplateFunc(t *testing.T) {
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
					config.promptBoolInteractiveTemplateFunc(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptBoolInteractiveTemplateFunc(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}

func TestPromptChoiceInteractiveTemplateFunc(t *testing.T) {
	for _, tc := range []struct {
		name              string
		prompt            string
		choices           []any
		args              []string
		stdinStr          string
		expectedStdoutStr string
		expected          string
		expectedErr       bool
	}{
		{
			name:              "response_without_default",
			prompt:            "choice",
			choices:           []any{"one", "two", "three"},
			stdinStr:          "one\n",
			expectedStdoutStr: "choice (one/two/three)? ",
			expected:          "one",
		},
		{
			name:              "response_with_default",
			prompt:            "choice",
			choices:           []any{"one", "two", "three"},
			args:              []string{"one"},
			stdinStr:          "two\n",
			expectedStdoutStr: "choice (one/two/three, default one)? ",
			expected:          "two",
		},
		{
			name:              "no_response_with_default",
			prompt:            "choice",
			choices:           []any{"one", "two", "three"},
			args:              []string{"three"},
			stdinStr:          "\n",
			expectedStdoutStr: "choice (one/two/three, default three)? ",
			expected:          "three",
		},
		{
			name:        "invalid_response",
			prompt:      "choice",
			choices:     []any{"one", "two", "three"},
			stdinStr:    "invalid\n",
			expectedErr: true,
		},
		{
			name:        "invalid_response_with_default",
			prompt:      "choice",
			choices:     []any{"one", "two", "three"},
			args:        []string{"one"},
			stdinStr:    "invalid\n",
			expectedErr: true,
		},
		{
			name:        "too_many_args",
			prompt:      "choice",
			choices:     []any{"one", "two", "three"},
			args:        []string{"two", "three"},
			stdinStr:    "\n",
			expectedErr: true,
		},
		{
			name:        "invalid_default",
			prompt:      "choice",
			choices:     []any{"one", "two", "three"},
			args:        []string{"four"},
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
					config.promptChoiceInteractiveTemplateFunc(tc.prompt, tc.choices, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptChoiceInteractiveTemplateFunc(tc.prompt, tc.choices, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}

func TestPromptIntInteractiveTemplateFunc(t *testing.T) {
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
					config.promptIntInteractiveTemplateFunc(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptIntInteractiveTemplateFunc(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}

func TestPromptMultichoiceInteractiveTemplateFunc(t *testing.T) {
	for _, tc := range []struct {
		name              string
		prompt            string
		multichoices      []any
		args              []any
		stdinStr          string
		expectedStdoutStr string
		expected          []string
		expectedErr       bool
	}{
		{
			name:              "response_without_default",
			prompt:            "multichoice",
			multichoices:      []any{"one", "two", "three"},
			stdinStr:          "one\n\n",
			expectedStdoutStr: "multichoice?\nChoices (one/two/three)\nChoice (ENTER to stop)> Choice (ENTER to stop)> ",
			expected:          []string{"one"},
		},
		{
			name:              "response_with_default",
			prompt:            "multichoice",
			multichoices:      []any{"one", "two", "three"},
			args:              []any{[]string{"one"}},
			stdinStr:          "two\n\n",
			expectedStdoutStr: "multichoice?\nChoices (one/two/three)\nDefault [one]\nChoice (ENTER to stop)> Choice (ENTER to stop)> ",
			expected:          []string{"two"},
		},
		{
			name:              "no_response_with_default",
			prompt:            "multichoice",
			multichoices:      []any{"one", "two", "three"},
			args:              []any{[]string{"three", "two"}},
			stdinStr:          "\n",
			expectedStdoutStr: "multichoice?\nChoices (one/two/three)\nDefault [three, two]\nChoice (ENTER to stop)> ",
			expected:          []string{"three", "two"},
		},
		{
			name:              "empty_response_without_default",
			prompt:            "multichoice",
			multichoices:      []any{"one", "two", "three"},
			stdinStr:          "[]\n",
			expectedStdoutStr: "multichoice?\nChoices (one/two/three)\nChoice (ENTER to stop)> ",
			expected:          []string{},
		},
		{
			name:              "empty_response_without_default",
			prompt:            "multichoice",
			multichoices:      []any{"one", "two", "three"},
			args:              []any{[]string{"three", "two"}},
			stdinStr:          "[]\n",
			expectedStdoutStr: "multichoice?\nChoices (one/two/three)\nDefault [three, two]\nChoice (ENTER to stop)> ",
			expected:          []string{},
		},
		{
			name:         "invalid_response",
			prompt:       "multichoice",
			multichoices: []any{"one", "two", "three"},
			stdinStr:     "invalid\n",
			expectedErr:  true,
		},
		{
			name:         "invalid_response_with_default",
			prompt:       "multichoice",
			multichoices: []any{"one", "two", "three"},
			args:         []any{[]string{"one"}},
			stdinStr:     "invalid\n",
			expectedErr:  true,
		},
		{
			name:         "invalid_default",
			prompt:       "multichoice",
			multichoices: []any{"one", "two", "three"},
			args:         []any{[]string{"four"}},
			stdinStr:     "\n",
			expectedErr:  true,
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
					config.promptMultichoiceInteractiveTemplateFunc(tc.prompt, tc.multichoices, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptMultichoiceInteractiveTemplateFunc(tc.prompt, tc.multichoices, tc.args...))

				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}

func TestPromptStringInteractiveTemplateFunc(t *testing.T) {
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
					config.promptStringInteractiveTemplateFunc(tc.prompt, tc.args...)
				})
			} else {
				assert.Equal(t, tc.expected, config.promptStringInteractiveTemplateFunc(tc.prompt, tc.args...))
				assert.Equal(t, tc.expectedStdoutStr, stdout.String())
			}
		})
	}
}
