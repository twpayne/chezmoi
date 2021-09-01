package chezmoi

import (
	"encoding/hex"
)

// A HexBytes is a []byte which is marshaled as a hex string.
type HexBytes []byte

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (h HexBytes) MarshalText() ([]byte, error) {
	if len(h) == 0 {
		return nil, nil
	}
	result := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(result, h)
	return result, nil
}

// UnmarshalText implements encoding.TextMarshaler.UnmarshalText.
func (h *HexBytes) UnmarshalText(text []byte) error {
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

func (h HexBytes) String() string {
	return hex.EncodeToString(h)
}
