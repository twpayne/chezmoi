package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
)

var (
	ignoreRxs = []*regexp.Regexp{
		regexp.MustCompile(`\.svg\z`),
		regexp.MustCompile(`\A\.git\z`),
		regexp.MustCompile(`\A\.idea\z`),
		regexp.MustCompile(`\A\.ruff_cache\z`),
		regexp.MustCompile(`\A\.vagrant\z`),
		regexp.MustCompile(`\b\.?venv\b`),
		regexp.MustCompile(`\A\.vscode\z`),
		regexp.MustCompile(`\ACOMMIT\z`),
		regexp.MustCompile(`\Aassets/chezmoi\.io/site\z`),
		regexp.MustCompile(`\Aassets/scripts/install\.ps1\z`),
		regexp.MustCompile(`\Acompletions/chezmoi\.ps1\z`),
		regexp.MustCompile(`\Adist\z`),
	}
	crlfLineEndingRx     = regexp.MustCompile(`\r\z`)
	trailingWhitespaceRx = regexp.MustCompile(`\s+\z`)
)

func lintData(filename string, data []byte) error {
	if !strings.HasPrefix(http.DetectContentType(data), "text/") {
		return nil
	}

	lines := bytes.Split(data, []byte{'\n'})

	var errs []error

	for i, line := range lines {
		switch {
		case crlfLineEndingRx.Match(line):
			errs = append(errs, fmt.Errorf("::error file=%s,line=%d::CRLF line ending", filename, i+1))
		case trailingWhitespaceRx.Match(line):
			errs = append(errs, fmt.Errorf("::error file=%s,line=%d::trailing whitespace", filename, i+1))
		}
	}

	if len(data) > 0 && len(lines[len(lines)-1]) != 0 {
		errs = append(errs, fmt.Errorf("::error file=%s,line=%d::no newline at end of file", filename, len(lines)+1))
	}

	return chezmoierrors.Combine(errs...)
}

func lintFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return lintData(filename, data)
}

func run() error {
	var lintErrs []error
	if err := fs.WalkDir(os.DirFS("."), ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		for _, rx := range ignoreRxs {
			if rx.MatchString(path) {
				if dirEntry.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		if dirEntry.Type().IsRegular() {
			lintErrs = append(lintErrs, lintFile(path))
		}
		return nil
	}); err != nil {
		return err
	}
	return chezmoierrors.Combine(lintErrs...)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
