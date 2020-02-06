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
func (s *Symlink) Apply(fs vfs.FS, mutator Mutator, follow bool, applyOptions *ApplyOptions) error {
	if applyOptions.Ignore(s.targetName) {
		return nil
	}
	target, err := s.Linkname()
	if err != nil {
		return err
	}
	targetPath := filepath.Join(applyOptions.DestDir, s.targetName)
	var info os.FileInfo
	if follow {
		info, err = fs.Stat(targetPath)
	} else {
		info, err = fs.Lstat(targetPath)
	}
	switch {
	case err == nil && target == "":
		return mutator.RemoveAll(targetPath)
	case os.IsNotExist(err) && target == "":
		return nil
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
func (s *Symlink) ConcreteValue(destDir string, ignore func(string) bool, sourceDir string, umask os.FileMode, recursive bool) (interface{}, error) {
	if ignore(s.targetName) {
		return nil, nil
	}
	linkname, err := s.Linkname()
	if err != nil {
		return nil, err
	}
	return &symlinkConcreteValue{
		Type:       "symlink",
		SourcePath: filepath.Join(sourceDir, s.SourceName()),
		TargetPath: filepath.Join(destDir, s.TargetName()),
		Template:   s.Template,
		Linkname:   linkname,
	}, nil
}

// Evaluate evaluates s's target.
func (s *Symlink) Evaluate(ignore func(string) bool) error {
	if ignore(s.targetName) {
		return nil
	}
	_, err := s.Linkname()
	return err
}

// Linkname returns s's link name.
func (s *Symlink) Linkname() (string, error) {
	if s.evaluateLinkname != nil {
		s.linkname, s.linknameErr = s.evaluateLinkname()
		s.evaluateLinkname = nil
	}
	return s.linkname, s.linknameErr
}

// SourceName implements Entry.SourceName.
func (s *Symlink) SourceName() string {
	return s.sourceName
}

// TargetName implements Entry.TargetName.
func (s *Symlink) TargetName() string {
	return s.targetName
}

// archive writes s to w.
func (s *Symlink) archive(w *tar.Writer, ignore func(string) bool, headerTemplate *tar.Header, umask os.FileMode) error {
	if ignore(s.targetName) {
		return nil
	}
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
