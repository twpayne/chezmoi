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

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type ageKeygenCmdConfig struct {
	convert     bool
	postQuantum bool
}

func (c *Config) newAgeKeygenCmd() *cobra.Command {
	ageKeygenCommand := &cobra.Command{
		GroupID: groupIDEncryption,
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
	ageKeygenCommand.Flags().
		BoolVar(&c.ageKeygen.postQuantum, "pq", c.ageKeygen.postQuantum, "generate a post-quantum key pair")

	return ageKeygenCommand
}

func (c *Config) runAgeKeygenCmd(cmd *cobra.Command, args []string) error {
	switch {
	case c.ageKeygen.convert && c.ageKeygen.postQuantum:
		return errors.New("--pq cannot be used with --convert")
	case c.ageKeygen.convert:
		return c.runAgeKeygenConvertCmd(args)
	default:
		return c.runAgeKeygenGenerateCmd(cmd, args)
	}
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

	var identity age.Identity
	var recipient age.Recipient
	if c.ageKeygen.postQuantum {
		key, err := age.GenerateHybridIdentity()
		if err != nil {
			return err
		}
		identity = key
		recipient = key.Recipient()
	} else {
		key, err := age.GenerateX25519Identity()
		if err != nil {
			return err
		}
		identity = key
		recipient = key.Recipient()
	}

	if stdout, ok := c.stdout.(*os.File); ok && term.IsTerminal(int(stdout.Fd())) {
		fmt.Fprintf(c.stderr, "Public key: %s\n", recipient)
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
	fmt.Fprintf(&builder, "# public key: %s\n", recipient)
	fmt.Fprintf(&builder, "%s\n", identity)
	return c.writeOutputString(builder.String(), 0o660)
}
