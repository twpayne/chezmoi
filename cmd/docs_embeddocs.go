// +build !nodocs
// +build !noembeddocs

package cmd

import (
	"strings"
)

// DocsDir is unused when chezmoi is built with embedded docs.
var DocsDir = ""

var docsPrefix = "docs/"

func getDocsFilenames() ([]string, error) {
	var docsFilenames []string
	for name := range assets {
		if strings.HasPrefix(name, docsPrefix) {
			docsFilenames = append(docsFilenames, strings.TrimPrefix(name, docsPrefix))
		}
	}
	return docsFilenames, nil
}

func getDoc(filename string) ([]byte, error) {
	return getAsset(docsPrefix + filename)
}
