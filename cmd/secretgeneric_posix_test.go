// +build !windows

package cmd

import "github.com/twpayne/chezmoi/internal/chezmoi"

func getSecretTestConfig() (*Config, []string) {
	return &Config{
			mutator: chezmoi.NullMutator{},
			GenericSecret: genericSecretCmdConfig{
				Command: "date",
			},
		},
		[]string{"+%Y-%M-%DT%H:%M:%SZ"}
}

func getSecretJSONTestConfig() (*Config, []string) {
	return &Config{
			mutator: chezmoi.NullMutator{},
			GenericSecret: genericSecretCmdConfig{
				Command: "date",
			},
		},
		[]string{`+{"date":"%Y-%M-%DT%H:%M:%SZ"}`}
}
