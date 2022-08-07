package cmd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type autoBool struct {
	auto  bool
	value bool
}

// autoBoolFlagCompletionFunc is a function that completes the value of autoBool
// flags.
var autoBoolFlagCompletionFunc = chezmoi.FlagCompletionFunc([]string{
	"1", "t", "T", "true", "TRUE", "True",
	"0", "f", "F", "false", "FALSE", "False",
	"auto", "AUTO", "Auto",
})

// Set implements github.com/spf13/pflag.Value.Set.
func (b *autoBool) Set(s string) error {
	if strings.ToLower(s) == "auto" {
		b.auto = true
		return nil
	}
	b.auto = false
	var err error
	b.value, err = parseBool(s)
	if err != nil {
		return err
	}
	return nil
}

func (b *autoBool) String() string {
	if b.auto {
		return "auto"
	}
	return strconv.FormatBool(b.value)
}

// Type implements github.com/spf13/pflag.Value.Type.
func (b *autoBool) Type() string {
	return "bool|auto"
}

// Value returns b's value, calling b's autoFunc if needed.
func (b *autoBool) Value(autoFunc func() bool) bool {
	if b.auto {
		b.value = autoFunc()
		b.auto = false
	}
	return b.value
}

// StringOrBoolToAutoBoolHookFunc is a
// github.com/mitchellh/mapstructure.DecodeHookFunc that parses an autoBool from
// a bool or string.
func StringOrBoolToAutoBoolHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeOf(autoBool{}) {
			return data, nil
		}
		var b autoBool
		switch data := data.(type) {
		case bool:
			b.auto = false
			b.value = data
		case string:
			if err := b.Set(data); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("expected a bool or string, got a %T", data)
		}
		return b, nil
	}
}
