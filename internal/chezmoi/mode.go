package chezmoi

// A Mode is a mode of operation.
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

func (m Mode) String() string {
	return string(m)
}

func (m Mode) Type() string {
	return "mode"
}
