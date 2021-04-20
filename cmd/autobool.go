package cmd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type autoBool struct {
	auto     bool
	autoFunc func() (bool, error)
	value    bool
	valueErr error
}

// newAutoBool returns a new autoBool.
func newAutoBool(autoFunc func() (bool, error)) *autoBool {
	return &autoBool{
		auto:     true,
		autoFunc: autoFunc,
	}
}

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
	return "bool or auto"
}

// Value returns b's value, calling b's autoFunc if needed.
func (b *autoBool) Value() (bool, error) {
	if b.auto {
		b.value, b.valueErr = b.autoFunc()
		b.auto = false
	}
	return b.value, b.valueErr
}

// StringOrBoolToAutoBoolHookFunc is a
// github.com/mitchellh/mapstructure.DecodeHookFunc that parses an autoBool from
// a bool or string.
func StringOrBoolToAutoBoolHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
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
