package chezmoi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/tailscale/hujson"
	"gopkg.in/yaml.v3"
)

// Formats.
var (
	FormatJSON  Format = formatJSON{}
	FormatJSONC Format = formatJSONC{}
	FormatTOML  Format = formatTOML{}
	FormatYAML  Format = formatYAML{}
)

var errExpectedEOF = errors.New("expected EOF")

// A Format is a serialization format.
type Format interface {
	Marshal(value any) ([]byte, error)
	Name() string
	Unmarshal(data []byte, value any) error
}

// A formatJSON implements the JSON serialization format.
type formatJSON struct{}

// A formatJSONC implements the JSONC serialization format.
type formatJSONC struct{}

// A formatTOML implements the TOML serialization format.
type formatTOML struct{}

// A formatYAML implements the YAML serialization format.
type formatYAML struct{}

var (
	// FormatsByName is a map of all FormatsByName by name.
	FormatsByName = map[string]Format{
		"jsonc": FormatJSONC,
		"json":  FormatJSON,
		"toml":  FormatTOML,
		"yaml":  FormatYAML,
	}

	// FormatsByExtension is a map of all Formats by extension.
	FormatsByExtension = map[string]Format{
		"jsonc": FormatJSONC,
		"json":  FormatJSON,
		"toml":  FormatTOML,
		"yaml":  FormatYAML,
		"yml":   FormatYAML,
	}
	FormatExtensions = slices.Sorted(maps.Keys(FormatsByExtension))
)

// Marshal implements Format.Marshal.
func (formatJSONC) Marshal(value any) ([]byte, error) {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return hujson.Format([]byte(builder.String()))
}

// Name implements Format.Name.
func (formatJSONC) Name() string {
	return "jsonc"
}

// Unmarshal implements Format.Unmarshal.
func (formatJSONC) Unmarshal(data []byte, value any) error {
	data, err := hujson.Standardize(data)
	if err != nil {
		return err
	}
	return FormatJSON.Unmarshal(data, value)
}

// Marshal implements Format.Marshal.
func (formatJSON) Marshal(value any) ([]byte, error) {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// Name implements Format.Name.
func (formatJSON) Name() string {
	return "json"
}

// Unmarshal implements Format.Unmarshal.
func (formatJSON) Unmarshal(data []byte, value any) error {
	switch value := value.(type) {
	case *any, *[]any, *map[string]any:
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		if err := decoder.Decode(value); err != nil {
			return err
		}
		if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
			return errExpectedEOF
		}
		switch value := value.(type) {
		case *any:
			*value = replaceJSONNumbersWithNumericValues(*value)
		case *[]any:
			*value = replaceJSONNumbersWithNumericValuesSlice(*value)
		case *map[string]any:
			*value = replaceJSONNumbersWithNumericValuesMap(*value)
		}
		return nil
	default:
		return json.Unmarshal(data, value)
	}
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

func isPrefixDotFormatDotTmpl(name, prefix string) bool {
	for extension := range FormatsByExtension {
		if name == prefix+"."+extension+TemplateSuffix {
			return true
		}
	}
	return false
}

// replaceJSONNumbersWithNumericValues replaces any json.Numbers in value with
// int64s or float64s if possible and returns the new value. If value is a slice
// or a map then it is mutated in place.
func replaceJSONNumbersWithNumericValues(value any) any {
	switch value := value.(type) {
	case json.Number:
		if int64Value, err := value.Int64(); err == nil {
			return int64Value
		}
		if float64Value, err := value.Float64(); err == nil {
			return float64Value
		}
		// If value cannot be represented as an int64 or a float64 then return
		// it as a string to preserve its value. Such values are valid JSON but
		// are unlikely to occur in practice. See
		// https://www.rfc-editor.org/rfc/rfc7159#section-6.
		return value.String()
	case []any:
		return replaceJSONNumbersWithNumericValuesSlice(value)
	case map[string]any:
		return replaceJSONNumbersWithNumericValuesMap(value)
	default:
		return value
	}
}

func replaceJSONNumbersWithNumericValuesMap(value map[string]any) map[string]any {
	for k, v := range value {
		value[k] = replaceJSONNumbersWithNumericValues(v)
	}
	return value
}

func replaceJSONNumbersWithNumericValuesSlice(value []any) []any {
	for i, e := range value {
		value[i] = replaceJSONNumbersWithNumericValues(e)
	}
	return value
}
