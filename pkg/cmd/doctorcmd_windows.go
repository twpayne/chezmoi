package cmd

import "github.com/twpayne/chezmoi/v2/pkg/chezmoi"

func (c *umaskCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	return checkResultSkipped, ""
}
