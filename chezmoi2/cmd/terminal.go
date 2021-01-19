package cmd

import (
	"bufio"
	"io"
	"strings"
)

// A terminal implements a subset of the methods implemented by
// golang.org/x/term.*Terminal.
type terminal interface {
	ReadLine() (string, error)
	ReadPassword(prompt string) (string, error)
	SetPrompt(prompt string)
}

// A dumbTerminal reads and writes without any terminal processing.
type dumbTerminal struct {
	reader *bufio.Reader
	writer io.Writer
	prompt []byte
}

// newDumbTerminal returns a new dumbTerminal.
func newDumbTerminal(reader io.Reader, writer io.Writer, prompt string) *dumbTerminal {
	return &dumbTerminal{
		reader: bufio.NewReader(reader),
		writer: writer,
		prompt: []byte(prompt),
	}
}

// ReadLine implements terminal.ReadLine.
func (t *dumbTerminal) ReadLine() (string, error) {
	_, err := t.writer.Write(t.prompt)
	if err != nil {
		return "", err
	}
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

// ReadPassword implements terminal.ReadPassword.
func (t *dumbTerminal) ReadPassword(prompt string) (string, error) {
	_, err := t.writer.Write([]byte(prompt))
	if err != nil {
		return "", err
	}
	password, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(password, "\n"), nil
}

func (t *dumbTerminal) SetPrompt(prompt string) {
	t.prompt = []byte(prompt)
}
