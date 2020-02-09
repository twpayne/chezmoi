package configparser

import (
	"io"

	"gopkg.in/yaml.v3"
)

func init() {
	parsers[".yaml"] = parseYAML
	parsers[".yml"] = parseYAML
}

func parseYAML(r io.Reader, value interface{}) error {
	return yaml.NewDecoder(r).Decode(value)
}
