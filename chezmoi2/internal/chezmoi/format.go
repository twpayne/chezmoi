package chezmoi

import (
	"encoding/json"
	"strings"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// A Format is a serialization format.
type Format interface {
	Marshal(value interface{}) ([]byte, error)
	Name() string
	Unmarshal(data []byte, value interface{}) error
}

// A jsonFormat implements the JSON serialization format.
type jsonFormat struct{}

// A tomlFormat implements the TOML serialization format.
type tomlFormat struct{}

// A yamlFormat implements the YAML serialization format.
type yamlFormat struct{}

// Formats is a map of all Formats by name.
var Formats = map[string]Format{
	"json": jsonFormat{},
	"toml": tomlFormat{},
	"yaml": yamlFormat{},
}

// Marshal implements Format.Marshal.
func (jsonFormat) Marshal(value interface{}) ([]byte, error) {
	sb := strings.Builder{}
	e := json.NewEncoder(&sb)
	e.SetIndent("", "  ")
	if err := e.Encode(value); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// Name implements Format.Name.
func (jsonFormat) Name() string {
	return "json"
}

// Unmarshal implements Format.Unmarshal.
func (jsonFormat) Unmarshal(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (tomlFormat) Marshal(value interface{}) ([]byte, error) {
	return toml.Marshal(value)
}

// Name implements Format.Name.
func (yamlFormat) Name() string {
	return "yaml"
}

// Unmarshal implements Format.Unmarshal.
func (tomlFormat) Unmarshal(data []byte, value interface{}) error {
	return toml.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (yamlFormat) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(value)
}

// Name implements Format.Name.
func (tomlFormat) Name() string {
	return "toml"
}

// Unmarshal implements Format.Unmarshal.
func (yamlFormat) Unmarshal(data []byte, value interface{}) error {
	return yaml.Unmarshal(data, value)
}
