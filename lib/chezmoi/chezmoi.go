package chezmoi

import (
	"bytes"
	"os"
	"regexp"
	"text/template"

	"github.com/pkg/errors"
)

var (
	nameRegexp        = regexp.MustCompile(`\A(?P<private>private_)?(?P<executable>executable_)?(?P<dot>dot_)?(?P<name>.*?)(?P<template>\.tmpl)?\z`)
	nameSubexpIndexes = makeSubexpIndexes(nameRegexp)
)

type Target struct {
	Name     string
	Mode     os.FileMode
	Contents []byte
}

func ParseTarget(sourceName string, contents []byte, data interface{}) (*Target, error) {
	m := nameRegexp.FindStringSubmatch(sourceName)
	if m == nil {
		return nil, errors.Errorf("invalid source name %q", sourceName)
	}
	name := m[nameSubexpIndexes["name"]]
	if m[nameSubexpIndexes["dot"]] != "" {
		name = "." + name
	}
	mode := os.FileMode(0666)
	if m[nameSubexpIndexes["executable"]] != "" {
		mode |= 0111
	}
	if m[nameSubexpIndexes["private"]] != "" {
		mode &= 0700
	}
	if m[nameSubexpIndexes["template"]] != "" {
		tmpl, err := template.New(sourceName).Parse(string(contents))
		if err != nil {
			return nil, errors.Wrap(err, sourceName)
		}
		output := &bytes.Buffer{}
		if err := tmpl.Execute(output, data); err != nil {
			return nil, errors.Wrap(err, sourceName)
		}
		contents = output.Bytes()
	}
	return &Target{
		Name:     name,
		Mode:     mode,
		Contents: contents,
	}, nil
}

func makeSubexpIndexes(re *regexp.Regexp) map[string]int {
	result := make(map[string]int)
	for index, name := range re.SubexpNames() {
		result[name] = index
	}
	return result
}
