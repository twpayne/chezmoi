package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

const ignoreFileName = chezmoi.Prefix + "ignore"

type ignorePattern struct {
	pattern string
	include bool
}

// parseIgnoreData parses the contents of .chezmoiignore into lines and patterns.
func parseIgnoreData(data []byte) ([]string, []ignorePattern) {
	lines := []string{}
	patterns := []ignorePattern{}
	if len(data) > 0 {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if trimmed == "" {
			continue
		}
		include := true
		if strings.HasPrefix(trimmed, "!") {
			include = false
			trimmed = strings.TrimPrefix(trimmed, "!")
		}
		patterns = append(patterns, ignorePattern{pattern: trimmed, include: include})
	}
	return lines, patterns
}

// matchIgnore reports whether name is ignored by patterns.
func matchIgnore(patterns []ignorePattern, name string) bool {
	includeMatch := false
	for _, p := range patterns {
		if ok, _ := doublestar.Match(p.pattern, name); ok {
			if !p.include {
				return false
			}
			includeMatch = true
		}
	}
	return includeMatch
}

func (c *Config) newIgnoreCmd() *cobra.Command {
	ignoreCmd := &cobra.Command{
		Use:   "ignore",
		Short: "Manage .chezmoiignore",
		Long:  mustLongHelp("ignore"),
		Annotations: newAnnotations(
			persistentStateModeEmpty,
		),
	}

	ignoreCmd.Flags().BoolP("force", "f", false, "Force the operation")

	ignoreAddCmd := &cobra.Command{
		Use:   "add [pattern]...",
		Short: "Add patterns to .chezmoiignore",
		Args:  cobra.ArbitraryArgs,
		RunE:  c.runIgnoreAddCmd,
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}
	ignoreCmd.AddCommand(ignoreAddCmd)

	ignoreRemoveCmd := &cobra.Command{
		Use:   "remove [pattern]...",
		Short: "Remove patterns from .chezmoiignore",
		Args:  cobra.ArbitraryArgs,
		RunE:  c.runIgnoreRemoveCmd,
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}
	ignoreCmd.AddCommand(ignoreRemoveCmd)

	ignoreActivateCmd := &cobra.Command{
		Use:     "activate [pattern]...",
		Aliases: []string{"-A"},
		Short:   "Activate patterns in .chezmoiignore",
		Args:    cobra.ArbitraryArgs,
		RunE:    c.runIgnoreActivateCmd,
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}
	ignoreCmd.AddCommand(ignoreActivateCmd)

	ignoreDeactivateCmd := &cobra.Command{
		Use:     "deactivate [pattern]...",
		Aliases: []string{"-D"},
		Short:   "Deactivate patterns in .chezmoiignore",
		Args:    cobra.ArbitraryArgs,
		RunE:    c.runIgnoreDeactivateCmd,
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}
	ignoreCmd.AddCommand(ignoreDeactivateCmd)

	ignoreQueryCmd := &cobra.Command{
		Use:   "query [name]...",
		Short: "Print active patterns matching name",
		Args:  cobra.ArbitraryArgs,
		RunE:  c.runIgnoreQueryCmd,
	}
	ignoreCmd.AddCommand(ignoreQueryCmd)

	return ignoreCmd
}

func (c *Config) readIgnorePatterns(args []string) ([]string, error) {
	var patterns []string
	switch {
	case len(args) == 1 && args[0] == "-":
		scanner := bufio.NewScanner(c.stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				patterns = append(patterns, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	default:
		patterns = append([]string(nil), args...)
	}
	return patterns, nil
}

func (c *Config) runIgnoreAddCmd(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	patterns, err := c.readIgnorePatterns(args)
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		return nil
	}

	ignoreAbsPath := c.sourceDirAbsPath.JoinString(ignoreFileName)

	data, err := c.sourceSystem.ReadFile(ignoreAbsPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	lines, ignorePatterns := parseIgnoreData(data)

	for _, p := range patterns {
		direct := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if trimmed == p {
				direct = true
				break
			}
		}
		if matchIgnore(ignorePatterns, p) && !force {
			return fmt.Errorf("%s is already ignored (use -f to force)", p)
		}
		if direct && !force {
			continue
		}
		lines = append(lines, p)
		include := true
		trimmed := p
		if strings.HasPrefix(p, "!") {
			include = false
			trimmed = strings.TrimPrefix(p, "!")
		}
		ignorePatterns = append(ignorePatterns, ignorePattern{pattern: trimmed, include: include})
	}

	output := strings.Join(lines, "\n")
	if output != "" && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	if err := chezmoi.MkdirAll(c.sourceSystem, c.sourceDirAbsPath, fs.ModePerm); err != nil &&
		!errors.Is(err, fs.ErrExist) {
		return err
	}
	return c.sourceSystem.WriteFile(ignoreAbsPath, []byte(output), 0o666&^c.Umask)
}

func (c *Config) runIgnoreRemoveCmd(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	patterns, err := c.readIgnorePatterns(args)
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		return nil
	}

	sourceDirAbsPath, err := c.getSourceDirAbsPath(nil)
	if err != nil {
		return err
	}
	ignoreAbsPath := sourceDirAbsPath.JoinString(ignoreFileName)

	data, err := c.sourceSystem.ReadFile(ignoreAbsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	lines, ignorePatterns := parseIgnoreData(data)

	for _, p := range patterns {
		var remaining []string
		removed := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if trimmed == p {
				removed = true
				continue
			}
			remaining = append(remaining, line)
		}

		if !removed {
			if matchIgnore(ignorePatterns, p) && !force {
				return fmt.Errorf("%s ignored by other patterns (use -f to force)", p)
			}
			lines = remaining
			continue
		}

		newPatterns := []ignorePattern{}
		for _, line := range remaining {
			trimmed := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if trimmed == "" {
				continue
			}
			include := true
			if strings.HasPrefix(trimmed, "!") {
				include = false
				trimmed = strings.TrimPrefix(trimmed, "!")
			}
			newPatterns = append(newPatterns, ignorePattern{pattern: trimmed, include: include})
		}
		if matchIgnore(newPatterns, p) && !force {
			return fmt.Errorf("%s would remain ignored (use -f to force)", p)
		}
		lines = remaining
		ignorePatterns = newPatterns
	}

	output := strings.Join(lines, "\n")
	if output != "" && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	return c.sourceSystem.WriteFile(ignoreAbsPath, []byte(output), 0o666&^c.Umask)
}

func (c *Config) runIgnoreToggleCmd(args []string, activate bool) error {
	patterns, err := c.readIgnorePatterns(args)
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		return nil
	}

	sourceDirAbsPath, err := c.getSourceDirAbsPath(nil)
	if err != nil {
		return err
	}
	ignoreAbsPath := sourceDirAbsPath.JoinString(ignoreFileName)

	data, err := c.sourceSystem.ReadFile(ignoreAbsPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	lines := []string{}
	if len(data) > 0 {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	for _, p := range patterns {
		pTrimmed := strings.TrimSpace(p)
		pBase := strings.TrimSpace(strings.SplitN(pTrimmed, "#", 2)[0])
		lastIdx := -1
		for i, line := range lines {
			lTrimmed := strings.TrimSpace(line)
			lCommented := strings.HasPrefix(lTrimmed, "#")
			lBase := strings.TrimSpace(strings.SplitN(pTrimmed, "#", 2)[0])
			if lBase == pBase && (!activate && !lCommented || activate && lCommented) {
				lastIdx = i
			}
		}

		switch {
		case lastIdx == -1:
			return fmt.Errorf("%s couldn't be found", pTrimmed)
		case activate:
			lines[lastIdx] = pBase
		case !activate:
			lines[lastIdx] = "# " + pBase
		}
	}

	output := strings.Join(lines, "\n")
	if output != "" && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	if err := chezmoi.MkdirAll(c.sourceSystem, ignoreAbsPath.Dir(), fs.ModePerm); err != nil {
		return err
	}
	return c.sourceSystem.WriteFile(ignoreAbsPath, []byte(output), 0o666&^c.Umask)
}

func (c *Config) runIgnoreActivateCmd(cmd *cobra.Command, args []string) error {
	return c.runIgnoreToggleCmd(args, true)
}

func (c *Config) runIgnoreDeactivateCmd(cmd *cobra.Command, args []string) error {
	return c.runIgnoreToggleCmd(args, false)
}

// readQueryNames reads names from args or interactively.
func (c *Config) readQueryNames(args []string) ([]string, error) {
	var names []string
	switch {
	case len(args) == 1 && args[0] == "-":
		scanner := bufio.NewScanner(c.stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				names = append(names, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	default:
		names = append([]string(nil), args...)
	}
	return names, nil
}

// runIgnoreQueryCmd prints active patterns matching any name.
func (c *Config) runIgnoreQueryCmd(cmd *cobra.Command, args []string) error {
	names, err := c.readQueryNames(args)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}

	sourceDirAbsPath, err := c.getSourceDirAbsPath(nil)
	if err != nil {
		return err
	}
	ignoreAbsPath := sourceDirAbsPath.JoinString(ignoreFileName)

	data, err := c.sourceSystem.ReadFile(ignoreAbsPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	lines := []string{}
	if len(data) > 0 {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	matches := []string{}
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		trimmed := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if trimmed == "" {
			continue
		}
		pattern := trimmed
		pattern = strings.TrimPrefix(pattern, "!")
		for _, name := range names {
			if ok, _ := doublestar.Match(pattern, name); ok {
				matches = append(matches, trimmed)
				break
			}
		}
	}

	if len(matches) == 0 {
		return chezmoi.ExitCodeError(1)
	}
	for _, m := range matches {
		fmt.Fprintln(c.stdout, m)
	}
	return nil
}
