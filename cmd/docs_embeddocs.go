// +build !noembeddocs

package cmd

import packr "github.com/gobuffalo/packr/v2"

// DocsDir is unused when chezmoi is built with embedded docs.
var DocsDir = ""

var docsBox = packr.New("docs", "../docs")

func getDocsFilenames() ([]string, error) {
	return docsBox.List(), nil
}

func getDoc(filename string) ([]byte, error) {
	return docsBox.Find(filename)
}
