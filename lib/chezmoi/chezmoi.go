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

type FileState struct {
	Name     string
	Mode     os.FileMode
	Contents []byte
}

type State []*FileState

func ParseFileState(filename string, contents []byte, data interface{}) (*FileState, error) {
	m := nameRegexp.FindStringSubmatch(filename)
	if m == nil {
		return nil, errors.Errorf("invalid source name %q", filename)
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
		tmpl, err := template.New(filename).Parse(string(contents))
		if err != nil {
			return nil, errors.Wrap(err, filename)
		}
		output := &bytes.Buffer{}
		if err := tmpl.Execute(output, data); err != nil {
			return nil, errors.Wrap(err, filename)
		}
		contents = output.Bytes()
	}
	return &FileState{
		Name:     name,
		Mode:     mode,
		Contents: contents,
	}, nil
}

type byName State

func (bn byName) Len() int           { return len(bn) }
func (bn byName) Less(i, j int) bool { return bn[i].Name < bn[j].Name }
func (bn byName) Swap(i, j int)      { bn[i], bn[j] = bn[j], bn[i] }

func makeSubexpIndexes(re *regexp.Regexp) map[string]int {
	result := make(map[string]int)
	for index, name := range re.SubexpNames() {
		result[name] = index
	}
	return result
}
