package pinentry

import (
	"bufio"
	"io"
	"os/exec"

	"go.uber.org/multierr"
)

// A Process abstracts the interface to a pinentry Process.
type Process interface {
	io.WriteCloser
	ReadLine() ([]byte, bool, error)
	Start(string, []string) error
}

// A execProcess executes a pinentry process.
type execProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

func (p *execProcess) Close() (err error) {
	defer func() {
		err = multierr.Append(err, p.cmd.Wait())
	}()
	err = p.stdin.Close()
	return
}

func (p *execProcess) ReadLine() ([]byte, bool, error) {
	return p.stdout.ReadLine()
}

func (p *execProcess) Start(name string, args []string) (err error) {
	p.cmd = exec.Command(name, args...)
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return
	}
	var stdoutPipe io.ReadCloser
	stdoutPipe, err = p.cmd.StdoutPipe()
	if err != nil {
		return
	}
	p.stdout = bufio.NewReader(stdoutPipe)
	err = p.cmd.Start()
	return
}

func (p *execProcess) Write(data []byte) (int, error) {
	return p.stdin.Write(data)
}
