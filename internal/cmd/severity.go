package cmd

import (
	"fmt"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type severity string

const (
	severityIgnore  severity = "ignore"
	severityWarning severity = "warning"
	severityError   severity = "error"
)

var severityFlagCompletionFunc = chezmoi.FlagCompletionFunc([]string{
	"i", "ignore",
	"w", "warning",
	"e", "error",
})

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (s severity) MarshalJSON() ([]byte, error) {
	switch s {
	case severityIgnore:
		return []byte(`"ignore"`), nil
	case severityWarning:
		return []byte(`"warning"`), nil
	case severityError:
		return []byte(`"error"`), nil
	default:
		return []byte(`"unknown"`), nil
	}
}

// MarshalYAML implements gopkg.in/yaml.v3.Marshaler.
func (s severity) MarshalYAML() (any, error) {
	return string(s), nil
}

// Set implements github.com/spf13/pflag.Value.Set.
func (s *severity) Set(str string) error {
	switch strings.ToLower(str) {
	case "i", "ignore":
		*s = severityIgnore
	case "w", "warning":
		*s = severityWarning
	case "e", "error":
		*s = severityError
	default:
		return fmt.Errorf("%s: unknown severity", str)
	}
	return nil
}

func (s *severity) String() string {
	return string(*s)
}

// Type implements github.com/spf13/pflag.Value.Type.
func (s *severity) Type() string {
	return "ignore|warning|error"
}
