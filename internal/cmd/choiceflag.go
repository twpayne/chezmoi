package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
	"chezmoi.io/chezmoi/internal/chezmoiset"
)

// A choiceFlag is a flag which accepts a limited set of allowed values.
type choiceFlag struct {
	value               string
	allowedValues       chezmoiset.Set[string]
	uniqueAbbreviations map[string]string
}

// newChoiceFlag returns a new choiceFlag with the given value and allowed
// values. If value is not allowed then it panics.
//
// If allowedValues is empty then all values are allowed. This functionality,
// although counter-intuitive, is required because the allowed values are
// carried in the value, not in the type, so a serialization/deserialization
// round trip discards the allowed values. To allow deserialization to succeed,
// we must allow all values.
func newChoiceFlag(value string, allowedValues []string) *choiceFlag {
	allowedValuesSet := chezmoiset.New(allowedValues...)
	if !allowedValuesSet.IsEmpty() && !allowedValuesSet.Contains(value) {
		panic("value not allowed")
	}
	return &choiceFlag{
		value:               value,
		allowedValues:       allowedValuesSet,
		uniqueAbbreviations: chezmoi.UniqueAbbreviations(allowedValues),
	}
}

// FlagCompletionFunc returns f's flag completion function.
func (f *choiceFlag) FlagCompletionFunc() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return chezmoi.FlagCompletionFunc(slices.Sorted(maps.Keys(f.allowedValues)))
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (f *choiceFlag) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(f.value)), nil
}

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (f *choiceFlag) MarshalText() ([]byte, error) {
	return []byte(f.value), nil
}

// Set implements github.com/spf13/pflag.Value.Set.
func (f *choiceFlag) Set(s string) error {
	// If uniqueAbbreviations is nil then all values are allowed. This
	// functionality, although counter-intuitive, is required because the unique
	// abbreviations are carried in the value, not in the type, so a
	// serialization/deserialization round trip discards the unique
	// abbreviations. To allow deserialization to succeed, we must allow all
	// values.
	if f.uniqueAbbreviations == nil {
		f.value = s
		return nil
	}
	value, ok := f.uniqueAbbreviations[s]
	if !ok {
		return errors.New("invalid value")
	}
	f.value = value
	return nil
}

func (f *choiceFlag) String() string {
	return f.value
}

// Type implements github.com/spf13/pflag.Value.Type.
func (f *choiceFlag) Type() string {
	sortedKeys := slices.Sorted(maps.Keys(f.allowedValues))
	if len(sortedKeys) > 0 && sortedKeys[0] == "" {
		sortedKeys[0] = "<none>"
	}
	return strings.Join(sortedKeys, "|")
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (f *choiceFlag) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if !f.allowedValues.IsEmpty() && !f.allowedValues.Contains(value) {
		return fmt.Errorf("%s: invalid value", value)
	}
	f.value = value
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.UnmarshalText.
func (f *choiceFlag) UnmarshalText(text []byte) error {
	value := string(text)
	if !f.allowedValues.IsEmpty() && !f.allowedValues.Contains(value) {
		return fmt.Errorf("%s: invalid value", value)
	}
	f.value = value
	return nil
}

// StringToChoiceFlagHookFunc is a
// github.com/go-viper/mapstructure/v2.DecodeHookFunc that parses a choiceFlag
// from a string.
//
// Unfortunately, we only receive the type of the value that we're decoding
// into, not its value, so we do not have access to the set of allowed values.
// So, we have to return a choiceFlag that allows all values.
func StringToChoiceFlagHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[choiceFlag]() {
			return data, nil
		}
		var cf choiceFlag
		switch data := data.(type) {
		case string:
			cf.value = data
		default:
			return nil, fmt.Errorf("expected a string, got a %T", data)
		}
		return cf, nil
	}
}
