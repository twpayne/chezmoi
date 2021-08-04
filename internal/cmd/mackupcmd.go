package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var (
	mackupCommentRx  = regexp.MustCompile(`\A#.*\z`)
	mackupKeyValueRx = regexp.MustCompile(`\A(\w+)\s*=\s*(.*)\z`)
	mackupSectionRx  = regexp.MustCompile(`\A\[(.*)\]\z`)
	mackupVersionRx  = regexp.MustCompile(`\AMackup\s+(\d+\.\d+\.\d+)\s*\z`)
)

type mackupApplicationApplicationConfig struct {
	Name string
}

type mackupApplicationConfig struct {
	Application           mackupApplicationApplicationConfig
	ConfigurationFiles    []chezmoi.RelPath
	XDGConfigurationFiles []chezmoi.RelPath
}

func (c *Config) newMackupCmd() *cobra.Command {
	if runtime.GOOS != "darwin" {
		return nil
	}

	mackupCmd := &cobra.Command{
		Use:    "mackup",
		Short:  "Interact with Mackup",
		Hidden: true,
	}

	mackupAddCmd := &cobra.Command{
		Use:   "add application...",
		Short: "Add an application's configuration from its Mackup configuration",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.makeRunEWithSourceState(c.runMackupAddCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			persistentStateMode:     persistentStateModeReadWrite,
			requiresSourceDirectory: "true",
		},
	}
	mackupCmd.AddCommand(mackupAddCmd)

	// FIXME add other subcommands like
	// mackup list
	// mackup forget

	return mackupCmd
}

func (c *Config) runMackupAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	mackupApplicationsDir, err := c.mackupApplicationsDir()
	if err != nil {
		return err
	}

	var addArgs []string
	for _, arg := range args {
		data, err := c.baseSystem.ReadFile(mackupApplicationsDir.Join(chezmoi.RelPath(arg + ".cfg")))
		if err != nil {
			return err
		}
		config, err := parseMackupApplication(data)
		if err != nil {
			return err
		}
		for _, filename := range config.ConfigurationFiles {
			addArg := c.DestDirAbsPath.Join(filename)
			addArgs = append(addArgs, addArg.String())
		}
		configHomeAbsPath := chezmoi.AbsPath(c.bds.ConfigHome)
		for _, filename := range config.XDGConfigurationFiles {
			addArg := configHomeAbsPath.Join(filename)
			addArgs = append(addArgs, addArg.String())
		}
	}

	destAbsPathInfos, err := c.destAbsPathInfos(sourceState, addArgs, destAbsPathInfosOptions{
		follow:         c.Add.follow,
		ignoreNotExist: true,
		recursive:      c.Add.recursive,
	})
	if err != nil {
		return err
	}

	return sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
		Include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
	})
}

func (c *Config) mackupApplicationsDir() (chezmoi.AbsPath, error) {
	brewPrefixCmd := exec.Command("brew", "--prefix")
	brewPrefixData, err := c.baseSystem.IdempotentCmdOutput(brewPrefixCmd)
	if err != nil {
		return "", err
	}
	brewPrefix := chezmoi.AbsPath(strings.TrimRight(string(brewPrefixData), "\n"))

	mackupVersionCmd := exec.Command("mackup", "--version")
	mackupVersionData, err := c.baseSystem.IdempotentCmdCombinedOutput(mackupVersionCmd)
	if err != nil {
		return "", err
	}
	m := mackupVersionRx.FindSubmatch(mackupVersionData)
	if m == nil {
		return "", fmt.Errorf("%q: cannot determine Mackup version", mackupVersionData)
	}
	mackupVersion := string(m[1])

	// FIXME the following line is not robust. It assumes that Python 3.9 is
	// being used. Replace it with something more robust.
	return brewPrefix.Join("Cellar", "mackup", chezmoi.RelPath(mackupVersion), "libexec", "lib", "python3.9", "site-packages", "mackup", "applications"), nil
}

func parseMackupApplication(data []byte) (mackupApplicationConfig, error) {
	var config mackupApplicationConfig
	var section string
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		text := s.Text()
		if mackupCommentRx.MatchString(text) {
			continue
		}
		if m := mackupSectionRx.FindStringSubmatch(s.Text()); m != nil {
			section = m[1]
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		//nolint:gocritic
		switch section {
		case "application":
			if m := mackupKeyValueRx.FindStringSubmatch(text); m != nil {
				switch m[1] {
				case "name":
					config.Application.Name = m[2]
				}
			}
		case "configuration_files":
			config.ConfigurationFiles = append(config.ConfigurationFiles, chezmoi.RelPath(text))
		case "xdg_configuration_files":
			config.XDGConfigurationFiles = append(config.XDGConfigurationFiles, chezmoi.RelPath(text))
		}
	}
	return config, s.Err()
}
