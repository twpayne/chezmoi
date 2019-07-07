// +build !windows

package cmd

func getSecretTestConfig() (*Config, []string) {
	return &Config{
			GenericSecret: genericSecretCmdConfig{
				Command: "date",
			},
		},
		[]string{"+%Y-%M-%DT%H:%M:%SZ"}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return &Config{
			GenericSecret: genericSecretCmdConfig{
				Command: "date",
			},
		},
		[]string{`+{"date":"%Y-%M-%DT%H:%M:%SZ"}`}
}
