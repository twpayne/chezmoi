// +build windows

package cmd

import "github.com/twpayne/chezmoi/internal/chezmoi"

func getSecretTestConfig() (*Config, []string) {
	// Windows doesn't (usually) have "date", but PowerShell is included with
	// all versions of Windows v7 or newer.
	return newConfig(
			withMutator(chezmoi.NullMutator{}),
			withGenericSecretCmdConfig(genericSecretCmdConfig{
				Command: "powershell.exe",
			}),
		),
		[]string{"-NoProfile", "-NonInteractive", "-Command", "Get-Date"}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return newConfig(
			withMutator(chezmoi.NullMutator{}),
			withGenericSecretCmdConfig(genericSecretCmdConfig{
				Command: "powershell.exe",
			}),
		),
		[]string{"-NoProfile", "-NonInteractive", "-Command", "Get-Date | ConvertTo-Json"}
}
