// +build !windows

package cmd

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

func getWriterSupportsColor(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}

func getWriterWidth(w io.Writer) int {
	if f, ok := w.(*os.File); ok && isatty.IsTerminal(f.Fd()) {
		if width, _, err := terminal.GetSize(int(f.Fd())); err == nil {
			return width
		}
	}
	return 80
}
