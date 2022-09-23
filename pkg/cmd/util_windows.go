package cmd

import (
	"io/fs"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

const defaultEditor = "notepad.exe"

var defaultInterpreters = map[string]*chezmoi.Interpreter{
	"bat": {},
	"cmd": {},
	"com": {},
	"exe": {},
	"pl": {
		Command: "perl",
	},
	"ps1": {
		Command: "powershell",
		Args:    []string{"-NoLogo"},
	},
	"py": {
		Command: "python3",
	},
	"rb": {
		Command: "ruby",
	},
}

func fileInfoUID(fs.FileInfo) int {
	return 0
}
