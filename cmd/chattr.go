package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var chattrCmd = &cobra.Command{
	Use:     "chattr attributes targets...",
	Args:    cobra.MinimumNArgs(2),
	Short:   "Change the attributes of a target in the source state",
	Long:    mustGetLongHelp("chattr"),
	Example: getExample("chattr"),
	PreRunE: config.ensureNoError,
	RunE:    config.runChattrCmd,
}

type boolModifier int

type attributeModifiers struct {
	empty      boolModifier
	encrypt    boolModifier
	exact      boolModifier
	executable boolModifier
	private    boolModifier
	template   boolModifier
}

func init() {
	rootCmd.AddCommand(chattrCmd)

	attributes := []string{
		"empty", "e",
		"encrypt",
		"exact",
		"executable", "x",
		"private", "p",
		"template", "t",
	}
	words := make([]string, 0, 4*len(attributes))
	for _, attribute := range attributes {
		words = append(words, attribute, "-"+attribute, "+"+attribute, "no"+attribute)
	}
	panicOnError(chattrCmd.MarkZshCompPositionalArgumentWords(1, words...))
	markRemainingZshCompPositionalArgumentsAsFiles(chattrCmd, 2)
}

func (c *Config) runChattrCmd(cmd *cobra.Command, args []string) error {
	ams, err := parseAttributeModifiers(args[0])
	if err != nil {
		return err
	}

	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}

	entries, err := c.getEntries(ts, args[1:])
	if err != nil {
		return err
	}

	updates := make(map[string]func() error)
	for _, entry := range entries {
		dir, oldBase := filepath.Split(entry.SourceName())
		oldpath := filepath.Join(ts.SourceDir, dir, oldBase)
		switch entry := entry.(type) {
		case *chezmoi.Dir:
			da := chezmoi.ParseDirAttributes(oldBase)
			da.Exact = ams.exact.modify(entry.Exact)
			perm := os.FileMode(0777)
			if private := ams.private.modify(entry.Private()); private {
				perm &= 0700
			}
			da.Perm = perm
			newBase := da.SourceName()
			if newBase != oldBase {
				newpath := filepath.Join(ts.SourceDir, dir, newBase)
				updates[oldpath] = func() error {
					return c.mutator.Rename(oldpath, newpath)
				}
			}
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
			fa.Encrypted = ams.encrypt.modify(entry.Encrypted)
			fa.Empty = ams.empty.modify(entry.Empty)
			fa.Template = ams.template.modify(entry.Template)
			newpath := filepath.Join(ts.SourceDir, dir, fa.SourceName())
			if fa.Encrypted != entry.Encrypted {
				oldContents, err := c.fs.ReadFile(filepath.Join(c.SourceDir, entry.SourceName()))
				if err != nil {
					return err
				}
				var newContents []byte
				if fa.Encrypted {
					newContents, err = ts.GPG.Encrypt(entry.TargetName(), oldContents)
				} else {
					newContents, err = ts.GPG.Decrypt(entry.TargetName(), oldContents)
				}
				if err != nil {
					return err
				}
				updates[oldpath] = func() error {
					// FIXME replace file and contents atomically, see
					// https://github.com/google/renameio/issues/16.
					if err := c.mutator.WriteFile(newpath, newContents, 0644, oldContents); err != nil {
						return err
					}
					return c.mutator.RemoveAll(oldpath)
				}
			} else if newpath != oldpath {
				updates[oldpath] = func() error {
					return c.mutator.Rename(oldpath, newpath)
				}
			}
		case *chezmoi.Symlink:
			fa := chezmoi.ParseFileAttributes(oldBase)
			fa.Template = ams.template.modify(entry.Template)
			newBase := fa.SourceName()
			if newBase != oldBase {
				newpath := filepath.Join(ts.SourceDir, dir, newBase)
				updates[oldpath] = func() error {
					return c.mutator.Rename(oldpath, newpath)
				}
			}
		}
	}

	// Sort oldpaths in reverse so we update files before their parent
	// directories.
	oldpaths := make([]string, 0, len(updates))
	for oldpath := range updates {
		oldpaths = append(oldpaths, oldpath)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(oldpaths)))

	// Apply all updates.
	for _, oldpath := range oldpaths {
		if err := updates[oldpath](); err != nil {
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
		case "encrypt":
			ams.encrypt = modifier
		case "exact":
			ams.exact = modifier
		case "executable", "x":
			ams.executable = modifier
		case "private", "p":
			ams.private = modifier
		case "template", "t":
			ams.template = modifier
		default:
			return nil, fmt.Errorf("%s: unknown attribute", attribute)
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
