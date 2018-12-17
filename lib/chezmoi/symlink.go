package chezmoi

import (
	"archive/tar"
	"os"
	"path/filepath"

	vfs "github.com/twpayne/go-vfs"
)

// A Symlink represents the target state of a symlink.
type Symlink struct {
	sourceName       string
	targetName       string
	Template         bool
	linkname         string
	linknameErr      error
	evaluateLinkname func() (string, error)
}

type symlinkConcreteValue struct {
	Type       string `json:"type" yaml:"type"`
	SourcePath string `json:"sourcePath" yaml:"sourcePath"`
	TargetPath string `json:"targetPath" yaml:"targetPath"`
	Template   bool   `json:"template" yaml:"template"`
	Linkname   string `json:"linkname" yaml:"linkname"`
}

// Apply ensures that the state of s's target in fs matches s.
func (s *Symlink) Apply(fs vfs.FS, targetDir string, umask os.FileMode, mutator Mutator) error {
	target, err := s.Linkname()
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, s.targetName)
	info, err := fs.Lstat(targetPath)
	switch {
	case err == nil && info.Mode()&os.ModeType == os.ModeSymlink:
		currentTarget, err := fs.Readlink(targetPath)
		if err != nil {
			return err
		}
		if currentTarget == target {
			return nil
		}
	case err == nil:
	case os.IsNotExist(err):
	default:
		return err
	}
	return mutator.WriteSymlink(target, targetPath)
}

// ConcreteValue implements Entry.ConcreteValue.
func (s *Symlink) ConcreteValue(targetDir, sourceDir string, recursive bool) (interface{}, error) {
	linkname, err := s.Linkname()
	if err != nil {
		return nil, err
	}
	return &symlinkConcreteValue{
		Type:       "symlink",
		SourcePath: filepath.Join(sourceDir, s.SourceName()),
		TargetPath: filepath.Join(targetDir, s.TargetName()),
		Template:   s.Template,
		Linkname:   linkname,
	}, nil
}

// Evaluate evaluates s's target.
func (s *Symlink) Evaluate() error {
	_, err := s.Linkname()
	return err
}

// SourceName implements Entry.SourceName.
func (s *Symlink) SourceName() string {
	return s.sourceName
}

// Linkname returns s's link name.
func (s *Symlink) Linkname() (string, error) {
	if s.evaluateLinkname != nil {
		s.linkname, s.linknameErr = s.evaluateLinkname()
		s.evaluateLinkname = nil
	}
	return s.linkname, s.linknameErr
}

// TargetName implements Entry.TargetName.
func (s *Symlink) TargetName() string {
	return s.targetName
}

// archive writes s to w.
func (s *Symlink) archive(w *tar.Writer, headerTemplate *tar.Header, umask os.FileMode) error {
	linkname, err := s.Linkname()
	if err != nil {
		return err
	}
	header := *headerTemplate
	header.Name = s.targetName
	header.Typeflag = tar.TypeSymlink
	header.Linkname = linkname
	return w.WriteHeader(&header)
}
