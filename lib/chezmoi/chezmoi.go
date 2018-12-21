package chezmoi

import (
	"archive/tar"
	"os"
	"path/filepath"
	"sort"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

const (
	symlinkPrefix    = "symlink_"
	privatePrefix    = "private_"
	emptyPrefix      = "empty_"
	executablePrefix = "executable_"
	dotPrefix        = "dot_"
	templateSuffix   = ".tmpl"
)

// A templateFuncError is an error encountered while executing a template
// function.
type templateFuncError struct {
	err error
}

// An Entry is either a Dir, a File, or a Symlink.
type Entry interface {
	Apply(fs vfs.FS, targetDir string, umask os.FileMode, mutator Mutator) error
	ConcreteValue(targetDir, sourceDir string, recursive bool) (interface{}, error)
	Evaluate() error
	SourceName() string
	TargetName() string
	archive(w *tar.Writer, headerTemplate *tar.Header, umask os.FileMode) error
}

type parsedSourceFilePath struct {
	FileAttributes
	dirNames []string
}

// ReturnTemplateFuncError causes template execution to return an error.
func ReturnTemplateFuncError(err error) {
	panic(templateFuncError{
		err: err,
	})
}

// parseDirNameComponents parses multiple directory name components. It returns
// the target directory names, target permissions, and any error.
func parseDirNameComponents(components []string) ([]string, []os.FileMode) {
	dirNames := []string{}
	perms := []os.FileMode{}
	for _, component := range components {
		da := ParseDirAttributes(component)
		dirNames = append(dirNames, da.Name)
		perms = append(perms, da.Perm)
	}
	return dirNames, perms
}

// parseSourceFilePath parses a single source file path.
func parseSourceFilePath(path string) parsedSourceFilePath {
	components := splitPathList(path)
	dirNames, _ := parseDirNameComponents(components[0 : len(components)-1])
	fa := ParseFileAttributes(components[len(components)-1])
	return parsedSourceFilePath{
		FileAttributes: fa,
		dirNames:       dirNames,
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
