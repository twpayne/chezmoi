package configparser

import (
	"io"

	"github.com/pelletier/go-toml"
)

func init() {
	parsers[".toml"] = parseTOML
}

func parseTOML(r io.Reader, value interface{}) error {
	return toml.NewDecoder(r).Decode(value)
}
