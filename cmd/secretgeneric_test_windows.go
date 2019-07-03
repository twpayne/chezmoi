// +build windows

package cmd

func getSecretTestConfig() (*Config, []string) {
	// Windows doesn't (usually) have "date", but powershell is included with
	// all versions of Windows v7 or newer.
	return &Config{
			GenericSecret: genericSecretCmdConfig{
				Command: "powershell.exe",
			},
		},
		[]string{"-Command", "Get-Date"}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return &Config{
			GenericSecret: genericSecretCmdConfig{
				Command: "powershell.exe",
			},
		},
		[]string{"-Command", "Get-Date | ConvertTo-Json"}
}
