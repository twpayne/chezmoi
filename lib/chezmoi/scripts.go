package chezmoi

import (
	"fmt"
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

// Script is a script supposed to run
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

// Apply executes the script if necessary to reach target state
func (s *Script) Apply() error {
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
	return s.execute()
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

func (s *Script) execute() error {
	c := exec.Command(s.executor)
	c.Dir = path.Dir(s.sourcePath)
	in, err := c.StdinPipe()
	if err != nil {
		return err
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	go func() {
		defer in.Close()
		if _, err := in.Write(s.contents); err != nil {
			panic(err)
		}
	}()

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
