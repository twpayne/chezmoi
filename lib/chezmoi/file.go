package chezmoi

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

// A FileAttributes holds attributes passed from a source file name.
type FileAttributes struct {
	Name     string
	Mode     os.FileMode
	Empty    bool
	Template bool
}

// A File represents the target state of a file.
type File struct {
	sourceName       string
	targetName       string
	Empty            bool
	Perm             os.FileMode
	Template         bool
	contents         []byte
	contentsErr      error
	evaluateContents func() ([]byte, error)
}

type fileConcreteValue struct {
	Type       string `json:"type" yaml:"type"`
	SourcePath string `json:"sourcePath" yaml:"sourcePath"`
	TargetPath string `json:"targetPath" yaml:"targetPath"`
	Empty      bool   `json:"empty" yaml:"empty"`
	Perm       int    `json:"perm" yaml:"perm"`
	Template   bool   `json:"template" yaml:"template"`
	Contents   string `json:"contents" yaml:"contents"`
}

// ParseFileAttributes parses a source file name.
func ParseFileAttributes(sourceName string) FileAttributes {
	name := sourceName
	mode := os.FileMode(0666)
	empty := false
	template := false
	if strings.HasPrefix(name, symlinkPrefix) {
		name = strings.TrimPrefix(name, symlinkPrefix)
		mode |= os.ModeSymlink
	} else {
		private := false
		if strings.HasPrefix(name, privatePrefix) {
			name = strings.TrimPrefix(name, privatePrefix)
			private = true
		}
		if strings.HasPrefix(name, emptyPrefix) {
			name = strings.TrimPrefix(name, emptyPrefix)
			empty = true
		}
		if strings.HasPrefix(name, executablePrefix) {
			name = strings.TrimPrefix(name, executablePrefix)
			mode |= 0111
		}
		if private {
			mode &= 0700
		}
	}
	if strings.HasPrefix(name, dotPrefix) {
		name = "." + strings.TrimPrefix(name, dotPrefix)
	}
	if strings.HasSuffix(name, templateSuffix) {
		name = strings.TrimSuffix(name, templateSuffix)
		template = true
	}
	return FileAttributes{
		Name:     name,
		Mode:     mode,
		Empty:    empty,
		Template: template,
	}
}

// SourceName returns fa's source name.
func (fa FileAttributes) SourceName() string {
	sourceName := ""
	switch fa.Mode & os.ModeType {
	case 0:
		if fa.Mode.Perm()&os.FileMode(077) == os.FileMode(0) {
			sourceName = privatePrefix
		}
		if fa.Empty {
			sourceName += emptyPrefix
		}
		if fa.Mode.Perm()&os.FileMode(0111) != os.FileMode(0) {
			sourceName += executablePrefix
		}
	case os.ModeSymlink:
		sourceName = symlinkPrefix
	default:
		panic(fmt.Sprintf("%+v: unsupported type", fa)) // FIXME return error instead of panicing
	}
	if strings.HasPrefix(fa.Name, ".") {
		sourceName += dotPrefix + strings.TrimPrefix(fa.Name, ".")
	} else {
		sourceName += fa.Name
	}
	if fa.Template {
		sourceName += templateSuffix
	}
	return sourceName
}

// Apply ensures that the state of targetPath in fs matches f.
func (f *File) Apply(fs vfs.FS, targetDir string, umask os.FileMode, mutator Mutator) error {
	contents, err := f.Contents()
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, f.targetName)
	info, err := fs.Lstat(targetPath)
	var currData []byte
	switch {
	case err == nil && info.Mode().IsRegular():
		if len(contents) == 0 && !f.Empty {
			return mutator.RemoveAll(targetPath)
		}
		currData, err = fs.ReadFile(targetPath)
		if err != nil {
			return err
		}
		if !bytes.Equal(currData, contents) {
			break
		}
		if info.Mode().Perm() != f.Perm&^umask {
			if err := mutator.Chmod(targetPath, f.Perm&^umask); err != nil {
				return err
			}
		}
		return nil
	case err == nil:
		if err := mutator.RemoveAll(targetPath); err != nil {
			return err
		}
	case os.IsNotExist(err):
	default:
		return err
	}
	if len(contents) == 0 && !f.Empty {
		return nil
	}
	return mutator.WriteFile(targetPath, contents, f.Perm&^umask, currData)
}

// ConcreteValue implements Entry.ConcreteValue.
func (f *File) ConcreteValue(targetDir, sourceDir string, recursive bool) (interface{}, error) {
	contents, err := f.Contents()
	if err != nil {
		return nil, err
	}
	return &fileConcreteValue{
		Type:       "file",
		SourcePath: filepath.Join(sourceDir, f.SourceName()),
		TargetPath: filepath.Join(targetDir, f.TargetName()),
		Empty:      f.Empty,
		Perm:       int(f.Perm),
		Template:   f.Template,
		Contents:   string(contents),
	}, nil
}

// Evaluate evaluates f's contents.
func (f *File) Evaluate() error {
	_, err := f.Contents()
	return err
}

// Contents returns f's contents.
func (f *File) Contents() ([]byte, error) {
	if f.evaluateContents != nil {
		f.contents, f.contentsErr = f.evaluateContents()
		f.evaluateContents = nil
	}
	return f.contents, f.contentsErr
}

// Executable returns true is f is executable.
func (f *File) Executable() bool {
	return f.Perm&0111 != 0
}

// Private returns true if f is private.
func (f *File) Private() bool {
	return f.Perm&077 == 0
}

// SourceName implements Entry.SourceName.
func (f *File) SourceName() string {
	return f.sourceName
}

// TargetName implements Entry.TargetName.
func (f *File) TargetName() string {
	return f.targetName
}

// archive writes f to w.
func (f *File) archive(w *tar.Writer, headerTemplate *tar.Header, umask os.FileMode) error {
	contents, err := f.Contents()
	if err != nil {
		return err
	}
	if len(contents) == 0 && !f.Empty {
		return nil
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = f.targetName
	header.Size = int64(len(contents))
	header.Mode = int64(f.Perm &^ umask)
	if err := w.WriteHeader(&header); err != nil {
		return nil
	}
	_, err = w.Write(contents)
	return err
}
