package chezmoi

// A Mode is a mode of operation. It implements the github.com/spf13/flag.Value
// interface.
type Mode string

// Modes.
const (
	ModeFile    Mode = "file"
	ModeSymlink Mode = "symlink"
)

type invalidModeError string

func (e invalidModeError) Error() string {
	return "invalid mode: " + string(e)
}

// ModeFlagCompletionFunc is a function that completes the value of mode flags.
var ModeFlagCompletionFunc = FlagCompletionFunc([]string{
	string(ModeFile),
	string(ModeSymlink),
})

// Set implements github.com/spf13/flag.Value.Set.
func (m *Mode) Set(s string) error {
	switch Mode(s) {
	case ModeFile:
		*m = ModeFile
		return nil
	case ModeSymlink:
		*m = ModeSymlink
		return nil
	default:
		return invalidModeError(Mode(s))
	}
}

// String implements github.com/spf13/flag.Value.String.
func (m Mode) String() string {
	return string(m)
}

// Type implements github.com/spf13/flag.Value.Type.
func (m Mode) Type() string {
	return "mode"
}
