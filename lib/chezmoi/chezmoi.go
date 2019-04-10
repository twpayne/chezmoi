package chezmoi

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

// Suffixes and prefixes.
const (
	symlinkPrefix    = "symlink_"
	privatePrefix    = "private_"
	emptyPrefix      = "empty_"
	encryptedPrefix  = "encrypted_"
	exactPrefix      = "exact_"
	executablePrefix = "executable_"
	dotPrefix        = "dot_"
	TemplateSuffix   = ".tmpl"
)

// A templateFuncError is an error encountered while executing a template
// function.
type templateFuncError struct {
	err error
}

// An Entry is either a Dir, a File, or a Symlink.
type Entry interface {
	Apply(fs vfs.FS, destDir string, ignore func(string) bool, umask os.FileMode, mutator Mutator) error
	ConcreteValue(destDir string, ignore func(string) bool, sourceDir string, umask os.FileMode, recursive bool) (interface{}, error)
	Evaluate(ignore func(string) bool) error
	SourceName() string
	TargetName() string
	archive(w *tar.Writer, ignore func(string) bool, headerTemplate *tar.Header, umask os.FileMode) error
}

type parsedSourceFilePath struct {
	FileAttributes
	dirAttributes []DirAttributes
}

// ReturnTemplateFuncError causes template execution to return an error.
func ReturnTemplateFuncError(err error) {
	panic(templateFuncError{
		err: err,
	})
}

// dirNames returns the dir names from dirAttributes.
func dirNames(dirAttributes []DirAttributes) []string {
	dns := make([]string, len(dirAttributes))
	for i, da := range dirAttributes {
		dns[i] = da.Name
	}
	return dns
}

// isEmpty returns true if b should be considered empty.
func isEmpty(b []byte) bool {
	return len(bytes.TrimSpace(b)) == 0
}

// parseDirNameComponents parses multiple directory name components.
func parseDirNameComponents(components []string) []DirAttributes {
	das := []DirAttributes{}
	for _, component := range components {
		da := ParseDirAttributes(component)
		das = append(das, da)
	}
	return das
}

// parseSourceFilePath parses a single source file path.
func parseSourceFilePath(path string) parsedSourceFilePath {
	components := splitPathList(path)
	das := parseDirNameComponents(components[0 : len(components)-1])
	fa := ParseFileAttributes(components[len(components)-1])
	return parsedSourceFilePath{
		FileAttributes: fa,
		dirAttributes:  das,
	}
}

// sortedEntryNames returns a sorted slice of all entry names.
func sortedEntryNames(entries map[string]Entry) []string {
	entryNames := []string{}
	for entryName := range entries {
		entryNames = append(entryNames, entryName)
	}
	sort.Strings(entryNames)
	return entryNames
}

func splitPathList(path string) []string {
	if strings.HasPrefix(path, string(filepath.Separator)) {
		path = strings.TrimPrefix(path, string(filepath.Separator))
	}
	return strings.Split(path, string(filepath.Separator))
}
