package cmd

// FIXME add documentation if we decide to keep this command

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
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
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
		),
	}
	mackupCmd.AddCommand(mackupAddCmd)

	// FIXME add other subcommands like
	// mackup list
	// mackup forget

	return mackupCmd
}

func (c *Config) runMackupAddCmd(
	cmd *cobra.Command,
	args []string,
	sourceState *chezmoi.SourceState,
) error {
	mackupApplicationsDir, err := c.mackupApplicationsDir()
	if err != nil {
		return err
	}

	var addArgs []string
	for _, arg := range args {
		data, err := c.baseSystem.ReadFile(
			mackupApplicationsDir.Join(chezmoi.NewRelPath(arg + ".cfg")),
		)
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
		configHomeAbsPath := chezmoi.NewAbsPath(c.bds.ConfigHome)
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

	return sourceState.Add(
		c.sourceSystem,
		c.persistentState,
		c.destSystem,
		destAbsPathInfos,
		&chezmoi.AddOptions{
			Filter:       chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			OnIgnoreFunc: c.defaultOnIgnoreFunc,
			PreAddFunc:   c.defaultPreAddFunc,
			ReplaceFunc:  c.defaultReplaceFunc,
		},
	)
}

func (c *Config) mackupApplicationsDir() (chezmoi.AbsPath, error) {
	brewPrefixCmd := exec.Command("brew", "--prefix")
	brewPrefixData, err := brewPrefixCmd.Output()
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	brewPrefix := chezmoi.NewAbsPath(strings.TrimRight(string(brewPrefixData), "\n"))

	mackupVersionCmd := exec.Command("mackup", "--version")
	mackupVersionData, err := mackupVersionCmd.Output()
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	mackupVersionMatch := mackupVersionRx.FindSubmatch(mackupVersionData)
	if mackupVersionMatch == nil {
		return chezmoi.EmptyAbsPath, fmt.Errorf(
			"%q: cannot determine Mackup version",
			mackupVersionData,
		)
	}
	mackupVersion := string(mackupVersionMatch[1])

	libDirAbsPath := brewPrefix.JoinString("Cellar", "mackup", mackupVersion, "libexec", "lib")
	dirEntries, err := c.baseSystem.ReadDir(libDirAbsPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	var pythonDirRelPath chezmoi.RelPath
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() && strings.HasPrefix(dirEntry.Name(), "python") {
			pythonDirRelPath = chezmoi.NewRelPath(dirEntry.Name())
			break
		}
	}
	if pythonDirRelPath.Empty() {
		return chezmoi.EmptyAbsPath, fmt.Errorf(
			"%s: could not find python directory",
			libDirAbsPath,
		)
	}

	return libDirAbsPath.Join(pythonDirRelPath).
			JoinString("site-packages", "mackup", "applications"),
		nil
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
			config.ConfigurationFiles = append(config.ConfigurationFiles, chezmoi.NewRelPath(text))
		case "xdg_configuration_files":
			config.XDGConfigurationFiles = append(
				config.XDGConfigurationFiles,
				chezmoi.NewRelPath(text),
			)
		}
	}
	return config, s.Err()
}
