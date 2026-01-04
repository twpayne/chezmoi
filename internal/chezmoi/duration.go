package chezmoi

import "time"

// A Duration is a [time.Duration] that implements [encoding.TextUnmarshaler].
type Duration time.Duration

func (d *Duration) UnmarshalText(data []byte) error {
	timeDuration, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(timeDuration)
	return nil
}
