package cmd

import (
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// FIXME consider tagging the dotfiles repo with chezmoi if it is public

type thanksCmdConfig struct {
	owner string
	repo  string
}

func (c *Config) newThanksCmd() *cobra.Command {
	thanksCmd := &cobra.Command{
		Use:    "thanks",
		Short:  "Say thanks for chezmoi by starring the repo",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE:   c.runThanksCmd,
	}

	return thanksCmd
}

func (c *Config) runThanksCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	gitHubClient := newGitHubClient(ctx)

	switch isStarred, _, err := gitHubClient.Activity.IsStarred(ctx, c.thanks.owner, c.thanks.repo); {
	case err != nil:
		return err
	case !isStarred:
		if _, err := gitHubClient.Activity.Star(ctx, c.thanks.owner, c.thanks.repo); err != nil {
			return browser.OpenURL("https://github.com/" + c.thanks.owner + "/" + c.thanks.repo + "/stargazers")
		}
	}

	return nil
}
