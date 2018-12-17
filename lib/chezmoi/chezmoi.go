package chezmoi

import (
	"archive/tar"
	"fmt"
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

// DirAttributes is a parsed source dir name.
type DirAttributes struct {
	Name string
	Perm os.FileMode
}

// A ParsedSourceFileName is a parsed source file name.
type ParsedSourceFileName struct {
	FileName string
	Mode     os.FileMode
	Empty    bool
	Template bool
}

type parsedSourceFilePath struct {
	ParsedSourceFileName
	dirNames []string
}

// ReturnTemplateFuncError causes template execution to return an error.
func ReturnTemplateFuncError(err error) {
	panic(templateFuncError{
		err: err,
	})
}

// ParseDirAttributes parses a single directory name.
func ParseDirAttributes(sourceName string) DirAttributes {
	name := sourceName
	perm := os.FileMode(0777)
	if strings.HasPrefix(name, privatePrefix) {
		name = strings.TrimPrefix(name, privatePrefix)
		perm &= 0700
	}
	if strings.HasPrefix(name, dotPrefix) {
		name = "." + strings.TrimPrefix(name, dotPrefix)
	}
	return DirAttributes{
		Name: name,
		Perm: perm,
	}
}

// SourceName returns da's source name.
func (da DirAttributes) SourceName() string {
	sourceName := ""
	if da.Perm&os.FileMode(077) == os.FileMode(0) {
		sourceName = privatePrefix
	}
	if strings.HasPrefix(da.Name, ".") {
		sourceName += dotPrefix + strings.TrimPrefix(da.Name, ".")
	} else {
		sourceName += da.Name
	}
	return sourceName
}

// ParseSourceFileName parses a source file name.
func ParseSourceFileName(fileName string) ParsedSourceFileName {
	mode := os.FileMode(0666)
	empty := false
	template := false
	if strings.HasPrefix(fileName, symlinkPrefix) {
		fileName = strings.TrimPrefix(fileName, symlinkPrefix)
		mode |= os.ModeSymlink
	} else {
		private := false
		if strings.HasPrefix(fileName, privatePrefix) {
			fileName = strings.TrimPrefix(fileName, privatePrefix)
			private = true
		}
		if strings.HasPrefix(fileName, emptyPrefix) {
			fileName = strings.TrimPrefix(fileName, emptyPrefix)
			empty = true
		}
		if strings.HasPrefix(fileName, executablePrefix) {
			fileName = strings.TrimPrefix(fileName, executablePrefix)
			mode |= 0111
		}
		if private {
			mode &= 0700
		}
	}
	if strings.HasPrefix(fileName, dotPrefix) {
		fileName = "." + strings.TrimPrefix(fileName, dotPrefix)
	}
	if strings.HasSuffix(fileName, templateSuffix) {
		fileName = strings.TrimSuffix(fileName, templateSuffix)
		template = true
	}
	return ParsedSourceFileName{
		FileName: fileName,
		Mode:     mode,
		Empty:    empty,
		Template: template,
	}
}

// SourceFileName returns psfn's source file name.
func (psfn ParsedSourceFileName) SourceFileName() string {
	fileName := ""
	switch psfn.Mode & os.ModeType {
	case 0:
		if psfn.Mode.Perm()&os.FileMode(077) == os.FileMode(0) {
			fileName = privatePrefix
		}
		if psfn.Empty {
			fileName += emptyPrefix
		}
		if psfn.Mode.Perm()&os.FileMode(0111) != os.FileMode(0) {
			fileName += executablePrefix
		}
	case os.ModeSymlink:
		fileName = symlinkPrefix
	default:
		panic(fmt.Sprintf("%+v: unsupported type", psfn)) // FIXME return error instead of panicing
	}
	if strings.HasPrefix(psfn.FileName, ".") {
		fileName += dotPrefix + strings.TrimPrefix(psfn.FileName, ".")
	} else {
		fileName += psfn.FileName
	}
	if psfn.Template {
		fileName += templateSuffix
	}
	return fileName
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
	psfn := ParseSourceFileName(components[len(components)-1])
	return parsedSourceFilePath{
		ParsedSourceFileName: psfn,
		dirNames:             dirNames,
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
