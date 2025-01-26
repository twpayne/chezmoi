package chezmoi

import (
	"encoding/hex"
	"strconv"
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

// MarshalYAML implements github.com/goccy/go-yaml.BytesMarshaler.MarshalYAML.
func (h HexBytes) MarshalYAML() ([]byte, error) {
	data := make([]byte, 2+2*len(h))
	data[0] = '"'
	hex.Encode(data[1:len(data)-1], []byte(h))
	data[len(data)-1] = '"'
	return data, nil
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

// UnmarshalYAML implements github.com/goccy/go-yaml.BytesUnmarshaler.UnmarshalYAML.
func (h *HexBytes) UnmarshalYAML(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	if s == "" {
		*h = nil
		return nil
	}
	hexBytes, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*h = hexBytes
	return nil
}

func (h HexBytes) String() string {
	return hex.EncodeToString(h)
}
