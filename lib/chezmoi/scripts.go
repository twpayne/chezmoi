package chezmoi

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
)

var shebangRegex *regexp.Regexp
var onceRegex *regexp.Regexp

func init() {
	shebangRegex = regexp.MustCompile(`(?m)^#!([^;\n]+)$`)
	onceRegex = regexp.MustCompile(`(?m)^#\s*chezmoi:\s*once$`)
}

// Script is a script supposed to run.
type Script struct {
	Name             string
	sourcePath       string
	executor         string
	once             bool
	alreadyExecuted  bool
	contents         []byte
	contentsErr      error
	evaluateContents func() ([]byte, error)
}

// Apply executes the script if necessary to reach target state.
func (s *Script) Apply(destDir string, dryRun bool) error {
	if s.once && s.alreadyExecuted {
		return nil
	}

	s.Contents()
	if s.contentsErr != nil {
		return s.contentsErr
	}

	if err := s.parse(); err != nil {
		return err
	}
	fmt.Printf("chezmoi: Running script %s\n", s.Name)
	if dryRun {
		println(string(s.contents), "\n")
		return nil
	}
	return s.execute(destDir)
}

func (s *Script) parse() error {
	if s.contentsErr != nil {
		return s.contentsErr
	}
	reg := shebangRegex.FindStringSubmatch(string(s.contents))
	if len(reg) < 2 {
		return fmt.Errorf("Shebang missing in script \"%s\"", s.Name)
	}
	s.executor = string(reg[1])
	s.once = onceRegex.Match(s.contents)
	return nil
}

func (s *Script) execute(destDir string) error {
	f, err := ioutil.TempFile(os.TempDir(), s.Name)
	if err != nil {
		return err
	}
	// We can ignore the error, the file is in a temporary dir anyways.
	defer os.Remove(f.Name())
	if _, err := f.Write(s.contents); err != nil {
		return err
	}
	if err := f.Chmod(os.FileMode(0700)); err != nil {
		return err
	}
	c := exec.Command(f.Name())
	c.Dir = destDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	return c.Run()
}

// Contents returns f's contents.
func (s *Script) Contents() ([]byte, error) {
	if s.evaluateContents != nil {
		s.contents, s.contentsErr = s.evaluateContents()
		s.evaluateContents = nil
	}
	return s.contents, s.contentsErr
}

// Evaluate evaluates s's script.
func (s *Script) Evaluate() error {
	if _, err := s.Contents(); err != nil {
		return err
	}
	return s.parse()
}
