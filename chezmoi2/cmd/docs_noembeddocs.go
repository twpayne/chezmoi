// +build !nodocs
// +build noembeddocs

package cmd

import (
	"os"
	"path/filepath"
)

// DocsDir is the directory containing docs when chezmoi is built without
// embedded docs. It should be an absolute path.
var DocsDir = "docs"

func doc(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(DocsDir, filename))
}

func docsFilenames() ([]string, error) {
	f, err := os.Open(DocsDir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdirnames(-1)
}
