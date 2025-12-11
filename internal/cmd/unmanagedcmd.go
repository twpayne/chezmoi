package cmd

import (
	"fmt"
	"io/fs"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
	"chezmoi.io/chezmoi/internal/chezmoiset"
)

type unmanagedCmdConfig struct {
	Ignore           []string `json:"ignore" mapstructure:"ignore" yaml:"ignore"`
	nulPathSeparator bool
	pathStyle        *choiceFlag
	tree             bool
}

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		GroupID: groupIDAdvanced,
		Use:     "unmanaged [path]...",
		Short:   "List the unmanaged files in the destination directory",
		Long:    mustLongHelp("unmanaged"),
		Example: example("unmanaged"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runUnmanagedCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}

	unmanagedCmd.Flags().
		BoolVarP(&c.Unmanaged.nulPathSeparator, "nul-path-separator", "0", c.Unmanaged.nulPathSeparator, "Use the NUL character as a path separator")
	unmanagedCmd.Flags().VarP(c.Unmanaged.pathStyle, "path-style", "p", "Path style")
	must(unmanagedCmd.RegisterFlagCompletionFunc("path-style", c.Unmanaged.pathStyle.FlagCompletionFunc()))
	unmanagedCmd.Flags().BoolVarP(&c.Unmanaged.tree, "tree", "t", c.Unmanaged.tree, "Print paths as a tree")

	return unmanagedCmd
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var absPaths []chezmoi.AbsPath
	if len(args) == 0 {
		absPaths = append(absPaths, c.DestDirAbsPath)
	} else {
		argsAbsPaths := chezmoiset.New[chezmoi.AbsPath]()
		for _, arg := range args {
			argAbsPath, err := chezmoi.NormalizePath(arg)
			if err != nil {
				return err
			}
			argsAbsPaths.Add(argAbsPath)
		}
		absPaths = argsAbsPaths.Elements()
		slices.Sort(absPaths)
	}

	unmanagedRelPaths := chezmoiset.New[chezmoi.RelPath]()
	walkFunc := func(destAbsPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			c.errorf("%s: %v\n", destAbsPath, err)
			if fileInfo == nil || fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if destAbsPath == c.DestDirAbsPath {
			return nil
		}
		targetRelPath, err := destAbsPath.TrimDirPrefix(c.DestDirAbsPath)
		if err != nil {
			return err
		}
		sourceStateEntry := sourceState.Get(targetRelPath)
		managed := sourceStateEntry != nil
		ignored := sourceState.Ignore(targetRelPath)
		if !managed && !ignored {
			unmanagedRelPaths.Add(targetRelPath)
		}
		if fileInfo.IsDir() {
			switch {
			case !managed:
				return fs.SkipDir
			case ignored:
				return fs.SkipDir
			case sourceStateEntry != nil:
				if external, ok := sourceStateEntry.Origin().(*chezmoi.External); ok {
					if external.Type == chezmoi.ExternalTypeGitRepo {
						return fs.SkipDir
					}
				}
			}
		}
		return nil
	}
	for _, absPath := range absPaths {
		if err := chezmoi.Walk(c.destSystem, absPath, walkFunc); err != nil {
			return err
		}
	}

UNMANAGED_REL_PATH:
	for unmanagedRelPath := range unmanagedRelPaths {
		for _, ignore := range c.Unmanaged.Ignore {
			switch ok, err := doublestar.Match(ignore, unmanagedRelPath.String()); {
			case err != nil:
				return err
			case ok:
				delete(unmanagedRelPaths, unmanagedRelPath)
				// FIXME both the gocritic and revive linters report the errors
				// for the following line, suggesting that "continue LABEL"
				// should be changed to "break". This is incorrect as the break
				// applies to the enclosing switch, not the loop.
				continue UNMANAGED_REL_PATH //nolint:gocritic,revive
			}
		}
	}

	paths := make([]fmt.Stringer, 0, len(unmanagedRelPaths.Elements()))
	for relPath := range unmanagedRelPaths {
		var path fmt.Stringer
		switch pathStyle := c.Unmanaged.pathStyle.String(); pathStyle {
		case pathStyleAbsolute:
			path = c.DestDirAbsPath.Join(relPath)
		case pathStyleRelative:
			path = relPath
		default:
			return fmt.Errorf("%s: invalid path style", pathStyle)
		}
		paths = append(paths, path)
	}

	return c.writePaths(stringersToStrings(paths), writePathsOptions{
		nulPathSeparator: c.Unmanaged.nulPathSeparator,
		tree:             c.Unmanaged.tree,
	})
}
