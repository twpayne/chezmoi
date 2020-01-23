// +build !nodocs
// +build !noembeddocs

package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/markbates/pkger"
)

// DocsDir is unused when chezmoi is built with embedded docs.
var DocsDir = ""

func init() {
	pkger.Include("/docs")
}

func getDocsFilenames() ([]string, error) {
	var docFilenames []string
	if err := pkger.Walk("/docs", func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			docFilenames = append(docFilenames, strings.TrimPrefix(filename, "github.com/twpayne/chezmoi:/docs/"))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return docFilenames, nil
}

func getDoc(filename string) ([]byte, error) {
	return pkgerReadFile(path.Join("/docs", filename))
}
