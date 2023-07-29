package chezmoi

import (
	"encoding/hex"
)

// A HexBytes is a []byte which is marshaled as a hex string.
type HexBytes []byte

// Bytes returns h as a []byte.
func (h HexBytes) Bytes() []byte {
	return []byte(h)
}

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (h HexBytes) MarshalText() ([]byte, error) {
	if len(h) == 0 {
		return nil, nil
	}
	result := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(result, h)
	return result, nil
}

// UnmarshalText implements encoding.TextUnmarshaler.UnmarshalText.
func (h *HexBytes) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*h = nil
		return nil
	}
	result := make([]byte, hex.DecodedLen(len(text)))
	if _, err := hex.Decode(result, text); err != nil {
		return err
	}
	*h = result
	return nil
}

func (h HexBytes) String() string {
	return hex.EncodeToString(h)
}
