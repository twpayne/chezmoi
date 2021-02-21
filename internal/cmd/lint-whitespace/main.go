package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.uber.org/multierr"
)

var (
	ignoreRxs = []*regexp.Regexp{
		regexp.MustCompile(`\.svg\z`),
		regexp.MustCompile(`\A\.devcontainer/library-scripts\z`),
		regexp.MustCompile(`\A\.git\z`),
		regexp.MustCompile(`\Aassets/scripts/install\.ps1\z`),
		regexp.MustCompile(`\Acompletions/chezmoi\.ps1\z`),
		regexp.MustCompile(`\Achezmoi\.io/public\z`),
		regexp.MustCompile(`\Achezmoi\.io/resources\z`),
		regexp.MustCompile(`\Achezmoi\.io/themes/book\z`),
	}
	crlfLineEndingRx     = regexp.MustCompile(`\r\z`)
	trailingWhitespaceRx = regexp.MustCompile(`\s+\z`)
)

func lintFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(http.DetectContentType(data), "text/") {
		return nil
	}

	lines := bytes.Split(data, []byte{'\n'})

	for i, line := range lines {
		switch {
		case crlfLineEndingRx.Match(line):
			err = multierr.Append(err, fmt.Errorf("%s:%d: CRLF line ending", filename, i+1))
		case trailingWhitespaceRx.Match(line):
			err = multierr.Append(err, fmt.Errorf("%s:%d: trailing whitespace", filename, i+1))
		}
	}

	if len(data) > 0 && len(lines[len(lines)-1]) != 0 {
		err = multierr.Append(err, fmt.Errorf("%s: no newline at end of file", filename))
	}

	return err
}

func run() error {
	filenames := make(map[string]struct{})
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, rx := range ignoreRxs {
			if rx.MatchString(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.Mode().IsRegular() {
			filenames[path] = struct{}{}
		}
		return nil
	}); err != nil {
		return err
	}

	sortedFilenames := make([]string, 0, len(filenames))
	for path := range filenames {
		sortedFilenames = append(sortedFilenames, path)
	}
	sort.Strings(sortedFilenames)

	var err error
	for _, filename := range sortedFilenames {
		err = multierr.Append(err, lintFile(filename))
	}
	return err
}

func main() {
	if err := run(); err != nil {
		for _, e := range multierr.Errors(err) {
			fmt.Println(e)
		}
		os.Exit(1)
	}
}
