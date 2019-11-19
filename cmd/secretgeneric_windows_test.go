// +build windows

package cmd

import "github.com/twpayne/chezmoi/internal/chezmoi"

func getSecretTestConfig() (*Config, []string) {
	// Windows doesn't (usually) have "date", but powershell is included with
	// all versions of Windows v7 or newer.
	return &Config{
			mutator: chezmoi.NullMutator{},
			GenericSecret: genericSecretCmdConfig{
				Command: "powershell.exe",
			},
		},
		[]string{"-NoProfile", "-NonInteractive", "-Command", "Get-Date"}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return &Config{
			mutator: chezmoi.NullMutator{},
			GenericSecret: genericSecretCmdConfig{
				Command: "powershell.exe",
			},
		},
		[]string{"-NoProfile", "-NonInteractive", "-Command", "Get-Date | ConvertTo-Json"}
}
