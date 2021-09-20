package cmd

import "github.com/twpayne/chezmoi/v2/internal/chezmoi"

func (c *umaskCheck) Run(system chezmoi.System) (checkResult, string) {
	return checkResultSkipped, ""
}
