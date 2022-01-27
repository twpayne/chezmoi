package cmd

import "github.com/twpayne/chezmoi/v2/pkg/chezmoi"

func (umaskCheck) Name() string {
	return "umask"
}

func (umaskCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	return checkResultSkipped, ""
}

func (unameCheck) Name() string {
	return "uname"
}

func (unameCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	return checkResultSkipped, ""
}
