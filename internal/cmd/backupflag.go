package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-viper/mapstructure/v2"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

// A backupFlag represents the value of the --backup flag which specifies the
// path to a tar archive for backed up files. An empty value disables backups.
type backupFlag struct{ path chezmoi.AbsPath }

// MarshalJSON implements json.Marshaler.
func (b backupFlag) MarshalJSON() ([]byte, error) { return json.Marshal(b.path.String()) }

// MarshalText implements encoding.TextMarshaler.
func (b backupFlag) MarshalText() ([]byte, error) { return []byte(b.path.String()), nil }

// Set implements pflag.Value.Set.
func (b *backupFlag) Set(s string) error {
	if s == "" {
		b.path = chezmoi.EmptyAbsPath
		return nil
	}
	homeDirAbsPath, err := chezmoi.HomeDirAbsPath()
	if err != nil {
		return err
	}
	absPath, err := chezmoi.NewAbsPathFromExtPath(s, homeDirAbsPath)
	if err != nil {
		return err
	}
	b.path = absPath
	return nil
}

func (b backupFlag) String() string { return b.path.String() }

func (b backupFlag) Type() string { return "path" }

// UnmarshalJSON implements json.Unmarshaler.
func (b *backupFlag) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || (len(data) == 2 && string(data) == "\"\"") {
		b.path = chezmoi.EmptyAbsPath
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return b.Set(s)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *backupFlag) UnmarshalText(text []byte) error { return b.Set(string(text)) }

// StringToBackupFlagHookFunc returns a DecodeHookFunc that parses a backupFlag
// from a string.
func StringToBackupFlagHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeOf(backupFlag{}) {
			return data, nil
		}
		var bf backupFlag
		if v, ok := data.(string); ok {
			if v == "" {
				return bf, nil
			}
			if err := bf.Set(v); err != nil {
				return nil, err
			}
			return bf, nil
		}
		return nil, fmt.Errorf("expected a string, got a %T", data)
	}
}
