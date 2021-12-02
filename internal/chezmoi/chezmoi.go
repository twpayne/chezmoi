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

	"github.com/coreos/go-semver/semver"
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
	dataName,
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

var modeTypeNames = map[fs.FileMode]string{
	0:                 "file",
	fs.ModeDir:        "dir",
	fs.ModeSymlink:    "symlink",
	fs.ModeNamedPipe:  "named pipe",
	fs.ModeSocket:     "socket",
	fs.ModeDevice:     "device",
	fs.ModeCharDevice: "char device",
}

// A TooOldErrror is returned when the source state requires a newer version of
// chezmoi.
type TooOldError struct {
	Have semver.Version
	Need semver.Version
}

func (e *TooOldError) Error() string {
	return fmt.Sprintf("source state requires version %s or later, chezmoi is version %s", e.Need, e.Have)
}

type inconsistentStateError struct {
	targetRelPath RelPath
	origins       []string
}

func (e *inconsistentStateError) Error() string {
	return fmt.Sprintf("%s: inconsistent state (%s)", e.targetRelPath, strings.Join(e.origins, ", "))
}

type notInAbsDirError struct {
	pathAbsPath AbsPath
	dirAbsPath  AbsPath
}

func (e *notInAbsDirError) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathAbsPath, e.dirAbsPath)
}

type notInRelDirError struct {
	pathRelPath RelPath
	dirRelPath  RelPath
}

func (e *notInRelDirError) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathRelPath, e.dirRelPath)
}

type unsupportedFileTypeError struct {
	absPath AbsPath
	mode    fs.FileMode
}

func (e *unsupportedFileTypeError) Error() string {
	return fmt.Sprintf("%s: unsupported file type %s", e.absPath, modeTypeName(e.mode))
}

// SHA256Sum returns the SHA256 sum of data.
func SHA256Sum(data []byte) []byte {
	sha256SumArr := sha256.Sum256(data)
	return sha256SumArr[:]
}

// SuspiciousSourceDirEntry returns true if base is a suspicious dir entry.
func SuspiciousSourceDirEntry(base string, fileInfo fs.FileInfo) bool {
	switch fileInfo.Mode().Type() {
	case 0:
		return strings.HasPrefix(base, Prefix) && !knownPrefixedFiles.contains(base)
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
