package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type ageKeygenCmdConfig struct {
	convert bool
}

func (c *Config) newAgeKeygenCmd() *cobra.Command {
	ageKeygenCommand := &cobra.Command{
		Use:     "age-keygen",
		Args:    cobra.MaximumNArgs(1),
		Short:   "Generate an age identity or convert an age identity to an age recipient",
		Long:    mustLongHelp("age-keygen"),
		Example: example("age-keygen"),
		RunE:    c.runAgeKeygenCmd,
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}
	ageKeygenCommand.Flags().
		BoolVarP(&c.ageKeygen.convert, "convert", "y", c.ageKeygen.convert, "convert identities to recipients")

	return ageKeygenCommand
}

func (c *Config) runAgeKeygenCmd(cmd *cobra.Command, args []string) error {
	if c.ageKeygen.convert {
		return c.runAgeKeygenConvertCmd(args)
	}
	return c.runAgeKeygenGenerateCmd(cmd, args)
}

func (c *Config) runAgeKeygenConvertCmd(args []string) error {
	input := c.stdin
	if len(args) > 0 {
		inputFile, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer inputFile.Close()
		input = inputFile
	}

	identities, err := age.ParseIdentities(input)
	switch {
	case err != nil:
		return err
	case len(identities) == 0:
		return errors.New("no identities found in input")
	}

	var builder strings.Builder
	for _, identity := range identities {
		id, ok := identity.(*age.X25519Identity)
		if !ok {
			return fmt.Errorf("internal error: unexpected identity type: %T", id)
		}
		builder.WriteString(id.Recipient().String())
		builder.WriteByte('\n')
	}
	return c.writeOutputString(builder.String(), 0o666)
}

func (c *Config) runAgeKeygenGenerateCmd(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return err
	}

	if stdout, ok := c.stdout.(*os.File); ok && term.IsTerminal(int(stdout.Fd())) {
		fmt.Fprintf(c.stderr, "Public key: %s\n", identity.Recipient())
	}

	if !c.outputAbsPath.IsEmpty() && c.outputAbsPath != chezmoi.NewAbsPath("-") {
		switch fileInfo, err := c.baseSystem.Stat(c.outputAbsPath); {
		case errors.Is(err, fs.ErrNotExist):
			// Do nothing.
		case err != nil:
			return err
		case fileInfo.Mode().IsRegular() && fileInfo.Mode().Perm()&0o004 != 0:
			c.errorf("writing secret key to a world-readable file\n")
		}
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&builder, "# public key: %s\n", identity.Recipient())
	fmt.Fprintf(&builder, "%s\n", identity)
	return c.writeOutputString(builder.String(), 0o660)
}
