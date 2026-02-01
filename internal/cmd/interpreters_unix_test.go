//go:build !windows

package cmd

import (
	"testing"
)

func TestNewDefaultInterpreters_PS1(t *testing.T) {
	tests := []struct {
		name           string
		findExecutable func([]string, []string) (string, error)
		wantCommand    string
		wantArgs       []string
	}{
		{
			name: "pwsh available",
			findExecutable: func(names, paths []string) (string, error) {
				if names[0] == "pwsh" {
					return "/usr/bin/pwsh", nil
				}
				return "", nil
			},
			wantCommand: "pwsh",
			wantArgs:    []string{"-NoLogo", "-File"},
		},
		{
			name: "pwsh not available",
			findExecutable: func(names, paths []string) (string, error) {
				return "", nil
			},
			wantCommand: "",
			wantArgs:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interpreters := NewDefaultInterpreters(tt.findExecutable)
			got := interpreters["ps1"]

			if got.Command != tt.wantCommand {
				t.Errorf("Command: got %q, want %q", got.Command, tt.wantCommand)
			}

			if len(got.Args) != len(tt.wantArgs) {
				t.Errorf("Args length: got %d, want %d", len(got.Args), len(tt.wantArgs))
			} else {
				for i := range got.Args {
					if got.Args[i] != tt.wantArgs[i] {
						t.Errorf("Args[%d]: got %q, want %q", i, got.Args[i], tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestNewDefaultInterpreters_OtherInterpreters(t *testing.T) {
	// Verify that other interpreters are not affected by ps1 changes
	interpreters := NewDefaultInterpreters(func([]string, []string) (string, error) {
		return "", nil
	})

	tests := []struct {
		ext         string
		wantCommand string
	}{
		{"nu", "nu"},
		{"pl", "perl"},
		{"py", "python3"},
		{"rb", "ruby"},
		{"bat", ""},
		{"cmd", ""},
		{"com", ""},
		{"exe", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := interpreters[tt.ext].Command
			if got != tt.wantCommand {
				t.Errorf("%s: got %q, want %q", tt.ext, got, tt.wantCommand)
			}
		})
	}
}
