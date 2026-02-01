package cmd

import "chezmoi.io/chezmoi/internal/chezmoi"

// NewDefaultInterpreters returns the default interpreters map, using the
// provided findExecutable function.
func NewDefaultInterpreters(findExecutable func([]string, []string) (string, error)) map[string]chezmoi.Interpreter {
	interpreters := map[string]chezmoi.Interpreter{
		"bat": {},
		"cmd": {},
		"com": {},
		"exe": {},
		"nu": {
			Command: "nu",
		},
		"pl": {
			Command: "perl",
		},
		// select the platform-appropriate interpreter for .ps1 scripts - prefer pwsh for UTF-8 and cross-platform support
		// if available with a fallback to powershell on Windows
		"ps1": func() chezmoi.Interpreter {
			i := getPS1Interpreter(findExecutable)
			return chezmoi.Interpreter{
				Command: i.Command,
				Args:    i.Args,
			}
		}(),
		"py": {
			Command: "python3",
		},
		"rb": {
			Command: "ruby",
		},
	}
	return interpreters
}

// DefaultInterpreters is the default interpreters map for the current platform.
var DefaultInterpreters = NewDefaultInterpreters(chezmoi.FindExecutable)
