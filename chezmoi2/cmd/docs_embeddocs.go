// +build !nodocs
// +build !noembeddocs

package cmd

import (
	"strings"
)

// DocsDir is unused when chezmoi is built with embedded docs.
var DocsDir = ""

var docsPrefix = "docs/"

func doc(filename string) ([]byte, error) {
	return asset(docsPrefix + filename)
}

func docsFilenames() ([]string, error) {
	var docsFilenames []string
	for name := range assets {
		if strings.HasPrefix(name, docsPrefix) {
			docsFilenames = append(docsFilenames, strings.TrimPrefix(name, docsPrefix))
		}
	}
	return docsFilenames, nil
}
