package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var chattrCommand = &cobra.Command{
	Use:   "chattr",
	Args:  cobra.MinimumNArgs(2),
	Short: "Change the exact, private, empty, executable, or template attributes of a target",
	RunE:  makeRunE(config.runChattrCommand),
}

type boolModifier int

type attributeModifiers struct {
	empty      boolModifier
	exact      boolModifier
	executable boolModifier
	private    boolModifier
	template   boolModifier
}

func init() {
	rootCommand.AddCommand(chattrCommand)
}

func (c *Config) runChattrCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	ams, err := parseAttributeModifiers(args[0])
	if err != nil {
		return err
	}
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args[1:])
	if err != nil {
		return err
	}
	renames := make(map[string]string)
	for _, entry := range entries {
		dir, oldBase := filepath.Split(entry.SourceName())
		var newBase string
		switch entry := entry.(type) {
		case *chezmoi.Dir:
			da := chezmoi.ParseDirAttributes(oldBase)
			da.Exact = ams.exact.modify(entry.Exact)
			perm := os.FileMode(0777)
			if private := ams.private.modify(entry.Private()); private {
				perm &= 0700
			}
			da.Perm = perm
			newBase = da.SourceName()
		case *chezmoi.File:
			fa := chezmoi.ParseFileAttributes(oldBase)
			mode := os.FileMode(0666)
			if executable := ams.executable.modify(entry.Executable()); executable {
				mode |= 0111
			}
			if private := ams.private.modify(entry.Private()); private {
				mode &= 0700
			}
			fa.Mode = mode
			fa.Empty = ams.empty.modify(entry.Empty)
			fa.Template = ams.template.modify(entry.Template)
			newBase = fa.SourceName()
		case *chezmoi.Symlink:
			fa := chezmoi.ParseFileAttributes(oldBase)
			fa.Template = ams.template.modify(entry.Template)
			newBase = fa.SourceName()
		}
		if newBase != oldBase {
			renames[filepath.Join(ts.SourceDir, dir, oldBase)] = filepath.Join(ts.SourceDir, dir, newBase)
		}
	}

	mutator := c.getDefaultMutator(fs)

	// Sort oldpaths in reverse so we rename files before their parent
	// directories.
	var oldpaths []string
	for oldpath := range renames {
		oldpaths = append(oldpaths, oldpath)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(oldpaths)))
	for _, oldpath := range oldpaths {
		if err := mutator.Rename(oldpath, renames[oldpath]); err != nil {
			return err
		}
	}
	return nil
}

func parseAttributeModifiers(s string) (*attributeModifiers, error) {
	ams := &attributeModifiers{}
	for _, attributeModifier := range strings.Split(s, ",") {
		attributeModifier = strings.TrimSpace(attributeModifier)
		if attributeModifier == "" {
			continue
		}
		var modifier boolModifier
		var attribute string
		switch {
		case attributeModifier[0] == '-':
			modifier = boolModifier(-1)
			attribute = attributeModifier[1:]
		case attributeModifier[0] == '+':
			modifier = boolModifier(1)
			attribute = attributeModifier[1:]
		case strings.HasPrefix(attributeModifier, "no"):
			modifier = boolModifier(-1)
			attribute = attributeModifier[2:]
		default:
			modifier = boolModifier(1)
			attribute = attributeModifier
		}
		switch attribute {
		case "empty", "e":
			ams.empty = modifier
		case "exact":
			ams.exact = modifier
		case "executable", "x":
			ams.executable = modifier
		case "private", "p":
			ams.private = modifier
		case "template", "t":
			ams.template = modifier
		default:
			return nil, fmt.Errorf("unknown attribute: %s", attribute)
		}
	}
	return ams, nil
}

func (bm boolModifier) modify(x bool) bool {
	switch {
	case bm < 0:
		return false
	case bm > 0:
		return true
	default:
		return x
	}
}
