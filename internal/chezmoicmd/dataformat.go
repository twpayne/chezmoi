package chezmoicmd

import (
	"errors"
	"strings"
)

// A dataFormat is either JSON or YAML and implements the
// github.com/spf13/pflag.Value interface.
type dataFormat string

const (
	dataFormatJSON dataFormat = "json"
	dataFormatYAML dataFormat = "yaml"

	defaultDataFormat = dataFormatJSON
)

func (f *dataFormat) Set(s string) error {
	switch strings.ToLower(s) {
	case "json":
		*f = dataFormatJSON
	case "yaml":
		*f = dataFormatYAML
	default:
		return errors.New("invalid data format")
	}
	return nil
}

func (f dataFormat) String() string {
	return string(f)
}

func (f dataFormat) Type() string {
	return "json|yaml"
}
