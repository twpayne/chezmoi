package chezmoi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
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
	Marshal(value any) ([]byte, error)
	Name() string
	Unmarshal(data []byte, value any) error
}

// A formatGzippedJSON implements the gzipped JSON serialization format.
type formatGzippedJSON struct{}

// A formatJSON implements the JSON serialization format.
type formatJSON struct{}

// A formatTOML implements the TOML serialization format.
type formatTOML struct{}

// A formatYAML implements the YAML serialization format.
type formatYAML struct{}

var (
	// FormatsByName is a map of all FormatsByName by name.
	FormatsByName = map[string]Format{
		"json": FormatJSON,
		"toml": FormatTOML,
		"yaml": FormatYAML,
	}

	// Formats is a map of all Formats by extension.
	FormatsByExtension = map[string]Format{
		"json": FormatJSON,
		"toml": FormatTOML,
		"yaml": FormatYAML,
		"yml":  FormatYAML,
	}

	FormatExtensions = sortedKeys(FormatsByExtension)
)

// Marshal implements Format.Marshal.
func (formatGzippedJSON) Marshal(value any) ([]byte, error) {
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
func (formatGzippedJSON) Unmarshal(data []byte, value any) (err error) {
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
func (formatJSON) Marshal(value any) ([]byte, error) {
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
func (formatJSON) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (formatTOML) Marshal(value any) ([]byte, error) {
	return toml.Marshal(value)
}

// Name implements Format.Name.
func (formatYAML) Name() string {
	return "yaml"
}

// Unmarshal implements Format.Unmarshal.
func (formatTOML) Unmarshal(data []byte, value any) error {
	return toml.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (formatYAML) Marshal(value any) ([]byte, error) {
	return yaml.Marshal(value)
}

// Name implements Format.Name.
func (formatTOML) Name() string {
	return "toml"
}

// Unmarshal implements Format.Unmarshal.
func (formatYAML) Unmarshal(data []byte, value any) error {
	return yaml.Unmarshal(data, value)
}

// FormatFromAbsPath returns the expected format of absPath.
func FormatFromAbsPath(absPath AbsPath) (Format, error) {
	format, err := formatFromExtension(absPath.Ext())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", absPath, err)
	}
	return format, nil
}

// formatFromExtension returns the expected format of absPath.
func formatFromExtension(extension string) (Format, error) {
	format, ok := FormatsByExtension[strings.TrimPrefix(extension, ".")]
	if !ok {
		return nil, fmt.Errorf("%s: unknown format", extension)
	}
	return format, nil
}

func isPrefixDotFormat(name, prefix string) bool {
	for extension := range FormatsByExtension {
		if name == prefix+"."+extension {
			return true
		}
	}
	return false
}
