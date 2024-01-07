package cmd

import (
	"github.com/spf13/cobra"
)

type gitCmdConfig struct {
	Command                   string `json:"command"                   mapstructure:"command"                   yaml:"command"`
	CommitMessageTemplate     string `json:"commitMessageTemplate"     mapstructure:"commitMessageTemplate"     yaml:"commitMessageTemplate"`
	CommitMessageTemplateFile string `json:"commitMessageTemplateFile" mapstructure:"commitMessageTemplateFile" yaml:"commitMessageTemplateFile"`
	AutoAdd                   bool   `json:"autoadd"                   mapstructure:"autoadd"                   yaml:"autoadd"`
	AutoCommit                bool   `json:"autocommit"                mapstructure:"autocommit"                yaml:"autocommit"`
	AutoPush                  bool   `json:"autopush"                  mapstructure:"autopush"                  yaml:"autopush"`
}

func (c *Config) newGitCmd() *cobra.Command {
	gitCmd := &cobra.Command{
		Use:     "git [arg]...",
		Short:   "Run git in the source directory",
		Long:    mustLongHelp("git"),
		Example: example("git"),
		RunE:    c.runGitCmd,
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			requiresWorkingTree,
			runsCommands,
		),
	}

	return gitCmd
}

func (c *Config) runGitCmd(cmd *cobra.Command, args []string) error {
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, args)
}
