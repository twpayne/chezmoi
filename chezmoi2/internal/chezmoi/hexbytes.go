package chezmoi

import (
	"encoding/hex"
)

// A hexBytes is a []byte which is marhsaled as a hex string.
type hexBytes []byte

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (h hexBytes) MarshalText() ([]byte, error) {
	if len(h) == 0 {
		return nil, nil
	}
	result := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(result, h)
	return result, nil
}

// UnmarshalText implements encoding.TextMarshaler.UnmarshalText.
func (h *hexBytes) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*h = nil
		return nil
	}
	result := make([]byte, hex.DecodedLen(len(text)))
	_, err := hex.Decode(result, text)
	if err != nil {
		return err
	}
	*h = result
	return nil
}
