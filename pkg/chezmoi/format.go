package chezmoi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// Formats.
var (
	FormatJSON Format = formatJSON{}
	FormatTOML Format = formatTOML{}
	FormatYAML Format = formatYAML{}
)

// A Format is a serialization format.
type Format interface {
	Marshal(value interface{}) ([]byte, error)
	Name() string
	Unmarshal(data []byte, value interface{}) error
}

// A formatGzippedJSON implements the gzipped JSON serialization format.
type formatGzippedJSON struct{}

// A formatJSON implements the JSON serialization format.
type formatJSON struct{}

// A formatTOML implements the TOML serialization format.
type formatTOML struct{}

// A formatYAML implements the YAML serialization format.
type formatYAML struct{}

// Formats is a map of all Formats by name.
var Formats = map[string]Format{
	"json": FormatJSON,
	"toml": FormatTOML,
	"yaml": FormatYAML,
}

// Marshal implements Format.Marshal.
func (formatGzippedJSON) Marshal(value interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	builder := &strings.Builder{}
	builder.Grow(len(jsonData))
	gzipWriter := gzip.NewWriter(builder)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// Name implements Format.Name.
func (formatGzippedJSON) Name() string {
	return "json.gz"
}

// Unmask implements Format.Unmarshal.
func (formatGzippedJSON) Unmarshal(data []byte, value interface{}) (err error) {
	var r *gzip.Reader
	if r, err = gzip.NewReader(bytes.NewReader(data)); err != nil {
		return
	}
	defer multierr.AppendInvoke(&err, multierr.Close(r))
	jsonData, err := io.ReadAll(r)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonData, value)
	return
}

// Marshal implements Format.Marshal.
func (formatJSON) Marshal(value interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// Name implements Format.Name.
func (formatJSON) Name() string {
	return "json"
}

// Unmarshal implements Format.Unmarshal.
func (formatJSON) Unmarshal(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (formatTOML) Marshal(value interface{}) ([]byte, error) {
	return toml.Marshal(value)
}

// Name implements Format.Name.
func (formatYAML) Name() string {
	return "yaml"
}

// Unmarshal implements Format.Unmarshal.
func (formatTOML) Unmarshal(data []byte, value interface{}) error {
	return toml.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (formatYAML) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(value)
}

// Name implements Format.Name.
func (formatTOML) Name() string {
	return "toml"
}

// Unmarshal implements Format.Unmarshal.
func (formatYAML) Unmarshal(data []byte, value interface{}) error {
	return yaml.Unmarshal(data, value)
}

func isPrefixDotFormat(name, prefix string) bool {
	for _, format := range Formats {
		if name == prefix+"."+format.Name() {
			return true
		}
	}
	return false
}
