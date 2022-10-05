// Package shell returns the current user's shell across multiple platforms.
package shell

import (
	"runtime"
)

// DefaultShell returns the default shell depending on runtime.GOOS.
func DefaultShell() string {
	switch runtime.GOOS {
	case "darwin":
		return "/bin/zsh"
	case "openbsd":
		return "/bin/ksh"
	case "plan9":
		return "/bin/rc"
	case "windows":
		return "powershell.exe"
	default:
		return "/bin/sh"
	}
}
