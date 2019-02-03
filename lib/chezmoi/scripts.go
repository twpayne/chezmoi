package chezmoi

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

var shebangRegex *regexp.Regexp
var onceRegex *regexp.Regexp

func init() {
	shebangRegex = regexp.MustCompile(`(?m)^#!([^;\n]+)$`)
	onceRegex = regexp.MustCompile(`(?m)^#\s*chezmoi:\s*once$`)
}

// Script is a script supposed to run
type Script struct {
	name             string
	sourcePath       string
	executor         string
	once             bool
	alreadyExecuted  bool
	contents         []byte
	contentsErr      error
	evaluateContents func() ([]byte, error)
}

func toScriptName(name string) string {
	i := strings.LastIndex(name, TemplateSuffix)
	if i > 0 {
		return name[0:i]
	}
	return name
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
	fmt.Printf("chezmoi: Running script %s\n", s.name)
	return s.execute()
}

func (s *Script) parse() error {
	if s.contentsErr != nil {
		return s.contentsErr
	}
	reg := shebangRegex.FindStringSubmatch(string(s.contents))
	if len(reg) < 2 {
		return fmt.Errorf("Shebang missing in script \"%s\"", s.name)
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
