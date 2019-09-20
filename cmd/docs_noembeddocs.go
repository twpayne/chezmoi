// +build !nodocs
// +build noembeddocs

package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// DocsDir is the directory containing docs when chezmoi is built without
// embedded docs. It should be an absolute path.
var DocsDir = "docs"

func getDocsFilenames() ([]string, error) {
	f, err := os.Open(DocsDir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdirnames(-1)
}

func getDoc(filename string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(DocsDir, filename))
}
