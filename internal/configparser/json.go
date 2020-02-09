package configparser

import (
	"encoding/json"
	"io"
)

func init() {
	parsers[".json"] = parseJSON
}

func parseJSON(r io.Reader, value interface{}) error {
	return json.NewDecoder(r).Decode(value)
}
