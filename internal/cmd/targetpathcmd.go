package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func (c *Config) newTargetPathCmd() *cobra.Command {
	targetPathCmd := &cobra.Command{
		GroupID: groupIDInternal,
		Use:     "target-path [source-path]...",
		Short:   "Print the target path of a source path",
		Long:    mustLongHelp("target-path"),
		Example: example("target-path"),
		RunE:    c.runTargetPathCmd,
		Annotations: newAnnotations(
			persistentStateModeReadMockWrite,
		),
	}

	return targetPathCmd
}

func (c *Config) runTargetPathCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.writeOutputString(c.DestDirAbsPath.String()+"\n", 0o666)
	}

	builder := strings.Builder{}

	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return err
		}

		argRelPath, err := argAbsPath.TrimDirPrefix(c.sourceDirAbsPath)
		if err != nil {
			return err
		}

		var sourceRelPath chezmoi.SourceRelPath
		switch fileInfo, err := c.sourceSystem.Stat(argAbsPath); {
		case err != nil:
			return err
		case fileInfo.IsDir():
			sourceRelPath = chezmoi.NewSourceRelDirPath(argRelPath.String())
		default:
			sourceRelPath = chezmoi.NewSourceRelPath(argRelPath.String())
		}

		targetRelPath := sourceRelPath.TargetRelPath(c.encryption.EncryptedSuffix())

		if _, err := builder.WriteString(c.DestDirAbsPath.String()); err != nil {
			return err
		}
		if err := builder.WriteByte('/'); err != nil {
			return err
		}
		if _, err := builder.WriteString(targetRelPath.String()); err != nil {
			return err
		}
		if err := builder.WriteByte('\n'); err != nil {
			return err
		}
	}

	return c.writeOutputString(builder.String(), 0o666)
}
