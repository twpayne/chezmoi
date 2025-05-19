package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var (
	mackupCommentRx  = regexp.MustCompile(`\A#.*\z`)
	mackupKeyValueRx = regexp.MustCompile(`\A(\w+)\s*=\s*(.*)\z`)
	mackupSectionRx  = regexp.MustCompile(`\A\[(.*)\]\z`)
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
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
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
	mackupAddCmd.Flags().Var(c.Add.Secrets, "secrets", "Scan for secrets when adding unencrypted files")
	must(mackupAddCmd.RegisterFlagCompletionFunc("secrets", c.Add.Secrets.FlagCompletionFunc()))
	mackupCmd.AddCommand(mackupAddCmd)

	return mackupCmd
}

func (c *Config) runMackupAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	mackupApplicationsDir, err := c.mackupApplicationsDir()
	if err != nil {
		return err
	}

	mackupDirAbsPath := c.homeDirAbsPath.JoinString(".mackup")
	var addArgs []string
	for _, arg := range args {
		configRelPath := chezmoi.NewRelPath(arg + ".cfg")
		data, err := c.baseSystem.ReadFile(mackupDirAbsPath.Join(configRelPath))
		if errors.Is(err, fs.ErrNotExist) {
			data, err = c.baseSystem.ReadFile(mackupApplicationsDir.Join(configRelPath))
		}
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
		onIgnoreFunc:   c.defaultOnIgnoreFunc,
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
			Errorf:       c.errorf,
			Filter:       chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			OnIgnoreFunc: c.defaultOnIgnoreFunc,
			PreAddFunc:   c.defaultPreAddFunc,
			ReplaceFunc:  c.defaultReplaceFunc,
		},
	)
}

func (c *Config) mackupApplicationsDir() (chezmoi.AbsPath, error) {
	mackupBinaryPath, err := chezmoi.LookPath("mackup")
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	mackupBinaryPathResolved, err := filepath.EvalSymlinks(mackupBinaryPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	mackupBinaryAbsPath := chezmoi.NewAbsPath(mackupBinaryPathResolved)

	libDirAbsPath := mackupBinaryAbsPath.Dir().Dir().JoinString("lib")
	dirEntries, err := c.baseSystem.ReadDir(libDirAbsPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() || !strings.HasPrefix(dirEntry.Name(), "python") {
			continue
		}
		mackupApplicationsDirAbsPath := libDirAbsPath.JoinString(dirEntry.Name(), "site-packages", "mackup", "applications")
		if fileInfo, err := c.baseSystem.Stat(mackupApplicationsDirAbsPath); err == nil && fileInfo.IsDir() {
			return mackupApplicationsDirAbsPath, nil
		}
	}

	return chezmoi.EmptyAbsPath, fmt.Errorf("%s: mackup application directory not found", libDirAbsPath)
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
		switch section {
		case "application":
			if m := mackupKeyValueRx.FindStringSubmatch(text); m != nil {
				if m[1] == "name" {
					config.Application.Name = m[2]
				}
			}
		case "configuration_files":
			config.ConfigurationFiles = append(config.ConfigurationFiles, chezmoi.NewRelPath(text))
		case "xdg_configuration_files":
			config.XDGConfigurationFiles = append(config.XDGConfigurationFiles, chezmoi.NewRelPath(text))
		}
	}
	return config, s.Err()
}
