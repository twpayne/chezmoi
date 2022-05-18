// Package chezmoi contains chezmoi's core logic.
package chezmoi

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// DefaultTemplateOptions are the default template options.
	DefaultTemplateOptions = []string{"missingkey=error"}

	// Break indicates that a walk should be stopped.
	Break = io.EOF

	// Skip indicates that entry should be skipped.
	Skip = filepath.SkipDir

	// Umask is the process's umask.
	Umask = fs.FileMode(0)
)

// Prefixes and suffixes.
const (
	ignorePrefix     = "."
	afterPrefix      = "after_"
	beforePrefix     = "before_"
	createPrefix     = "create_"
	dotPrefix        = "dot_"
	emptyPrefix      = "empty_"
	encryptedPrefix  = "encrypted_"
	exactPrefix      = "exact_"
	executablePrefix = "executable_"
	literalPrefix    = "literal_"
	modifyPrefix     = "modify_"
	oncePrefix       = "once_"
	onChangePrefix   = "onchange_"
	privatePrefix    = "private_"
	readOnlyPrefix   = "readonly_"
	removePrefix     = "remove_"
	runPrefix        = "run_"
	symlinkPrefix    = "symlink_"
	literalSuffix    = ".literal"
	TemplateSuffix   = ".tmpl"
)

// Special file names.
const (
	Prefix = ".chezmoi"

	RootName         = Prefix + "root"
	VersionName      = Prefix + "version"
	dataName         = Prefix + "data"
	externalName     = Prefix + "external"
	ignoreName       = Prefix + "ignore"
	removeName       = Prefix + "remove"
	scriptsDirName   = Prefix + "scripts"
	templatesDirName = Prefix + "templates"
)

var (
	dirPrefixRegexp  = regexp.MustCompile(`\A(dot|exact|literal|readonly|private)_`)
	filePrefixRegexp = regexp.MustCompile(
		`\A(after|before|create|dot|empty|encrypted|executable|literal|modify|once|private|readonly|remove|run|symlink)_`,
	)
	fileSuffixRegexp = regexp.MustCompile(`\.(literal|tmpl)\z`)
)

// knownPrefixedFiles is a set of known filenames with the .chezmoi prefix.
var knownPrefixedFiles = newStringSet(
	Prefix+".json"+TemplateSuffix,
	Prefix+".toml"+TemplateSuffix,
	Prefix+".yaml"+TemplateSuffix,
	RootName,
	dataName+".json",
	dataName+".toml",
	dataName+".yaml",
	externalName+".json",
	externalName+".toml",
	externalName+".yaml",
	ignoreName,
	removeName,
	VersionName,
)

// knownPrefixedDirs is a set of known dirnames with the .chezmoi prefix.
var knownPrefixedDirs = newStringSet(
	scriptsDirName,
	templatesDirName,
)

// knownTargetFiles is a set of known target files that should not be managed
// directly.
var knownTargetFiles = newStringSet(
	"chezmoi.json",
	"chezmoi.toml",
	"chezmoi.yaml",
	"chezmoistate.boltdb",
)

var modeTypeNames = map[fs.FileMode]string{
	0:                 "file",
	fs.ModeDir:        "dir",
	fs.ModeSymlink:    "symlink",
	fs.ModeNamedPipe:  "named pipe",
	fs.ModeSocket:     "socket",
	fs.ModeDevice:     "device",
	fs.ModeCharDevice: "char device",
}

// FlagCompletionFunc returns a flag completion function.
func FlagCompletionFunc(allCompletions []string) func(*cobra.Command, []string, string) (
	[]string, cobra.ShellCompDirective,
) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for _, completion := range allCompletions {
			if strings.HasPrefix(completion, toComplete) {
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// SHA256Sum returns the SHA256 sum of data.
func SHA256Sum(data []byte) []byte {
	sha256SumArr := sha256.Sum256(data)
	return sha256SumArr[:]
}

// SuspiciousSourceDirEntry returns true if base is a suspicious dir entry.
func SuspiciousSourceDirEntry(base string, fileInfo fs.FileInfo, encryptedSuffixes []string) bool {
	switch fileInfo.Mode().Type() {
	case 0:
		if strings.HasPrefix(base, Prefix) && !knownPrefixedFiles.contains(base) {
			return true
		}
		for _, encryptedSuffix := range encryptedSuffixes {
			if fileAttr := parseFileAttr(fileInfo.Name(), encryptedSuffix); knownTargetFiles.contains(fileAttr.TargetName) {
				return true
			}
		}
		return false
	case fs.ModeDir:
		return strings.HasPrefix(base, Prefix) && !knownPrefixedDirs.contains(base)
	case fs.ModeSymlink:
		return strings.HasPrefix(base, Prefix)
	default:
		return true
	}
}

// isEmpty returns true if data is empty after trimming whitespace from both
// ends.
func isEmpty(data []byte) bool {
	return len(bytes.TrimSpace(data)) == 0
}

// modeTypeName returns a string representation of mode.
func modeTypeName(mode fs.FileMode) string {
	if name, ok := modeTypeNames[mode.Type()]; ok {
		return name
	}
	return fmt.Sprintf("0o%o: unknown type", mode.Type())
}

// mustTrimPrefix is like strings.TrimPrefix but panics if s is not prefixed by
// prefix.
func mustTrimPrefix(s, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		panic(fmt.Sprintf("%s: not prefixed by %s", s, prefix))
	}
	return s[len(prefix):]
}

// mustTrimSuffix is like strings.TrimSuffix but panics if s is not suffixed by
// suffix.
func mustTrimSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		panic(fmt.Sprintf("%s: not suffixed by %s", s, suffix))
	}
	return s[:len(s)-len(suffix)]
}

// ensureSuffix adds suffix to s if s is not suffixed by suffix.
func ensureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}
