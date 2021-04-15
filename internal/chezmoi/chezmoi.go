// Package chezmoi contains chezmoi's core logic.
package chezmoi

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// DefaultTemplateOptions are the default template options.
	DefaultTemplateOptions = []string{"missingkey=error"}

	// Skip indicates that entry should be skipped.
	Skip = filepath.SkipDir

	// Umask is the process's umask.
	Umask = os.FileMode(0)
)

// Suffixes and prefixes.
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
	modifyPrefix     = "modify_"
	oncePrefix       = "once_"
	privatePrefix    = "private_"
	runPrefix        = "run_"
	symlinkPrefix    = "symlink_"
	TemplateSuffix   = ".tmpl"
)

// Special file names.
const (
	Prefix = ".chezmoi"

	dataName         = Prefix + "data"
	ignoreName       = Prefix + "ignore"
	removeName       = Prefix + "remove"
	templatesDirName = Prefix + "templates"
	versionName      = Prefix + "version"
)

var knownPrefixedFiles = map[string]bool{
	Prefix + ".json" + TemplateSuffix: true,
	Prefix + ".toml" + TemplateSuffix: true,
	Prefix + ".yaml" + TemplateSuffix: true,
	dataName:                          true,
	ignoreName:                        true,
	removeName:                        true,
	versionName:                       true,
}

var modeTypeNames = map[os.FileMode]string{
	0:                 "file",
	os.ModeDir:        "dir",
	os.ModeSymlink:    "symlink",
	os.ModeNamedPipe:  "named pipe",
	os.ModeSocket:     "socket",
	os.ModeDevice:     "device",
	os.ModeCharDevice: "char device",
}

type errDuplicateTarget struct {
	targetRelPath  RelPath
	sourceRelPaths SourceRelPaths
}

func (e *errDuplicateTarget) Error() string {
	sourceRelPathStrs := make([]string, 0, len(e.sourceRelPaths))
	for _, sourceRelPath := range e.sourceRelPaths {
		sourceRelPathStrs = append(sourceRelPathStrs, sourceRelPath.String())
	}
	return fmt.Sprintf("%s: duplicate source state entries (%s)", e.targetRelPath, strings.Join(sourceRelPathStrs, ", "))
}

type errNotInAbsDir struct {
	pathAbsPath AbsPath
	dirAbsPath  AbsPath
}

func (e *errNotInAbsDir) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathAbsPath, e.dirAbsPath)
}

type errNotInRelDir struct {
	pathRelPath RelPath
	dirRelPath  RelPath
}

func (e *errNotInRelDir) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathRelPath, e.dirRelPath)
}

type errUnsupportedFileType struct {
	absPath AbsPath
	mode    os.FileMode
}

func (e *errUnsupportedFileType) Error() string {
	return fmt.Sprintf("%s: unsupported file type %s", e.absPath, modeTypeName(e.mode))
}

// SHA256Sum returns the SHA256 sum of data.
func SHA256Sum(data []byte) []byte {
	sha256SumArr := sha256.Sum256(data)
	return sha256SumArr[:]
}

// SuspiciousSourceDirEntry returns true if base is a suspicious dir entry.
func SuspiciousSourceDirEntry(base string, info os.FileInfo) bool {
	//nolint:exhaustive
	switch info.Mode() & os.ModeType {
	case 0:
		return strings.HasPrefix(base, Prefix) && !knownPrefixedFiles[base]
	case os.ModeDir:
		return strings.HasPrefix(base, Prefix) && base != templatesDirName
	case os.ModeSymlink:
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

func modeTypeName(mode os.FileMode) string {
	if name, ok := modeTypeNames[mode&os.ModeType]; ok {
		return name
	}
	return fmt.Sprintf("0o%o: unknown type", mode&os.ModeType)
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
