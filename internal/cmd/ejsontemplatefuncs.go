package cmd

import (
	"encoding/json"

	"github.com/Shopify/ejson"
)

type ejsonConfig struct {
	KeyDir string `json:"keyDir" mapstructure:"keyDir" yaml:"keyDir"`
	Key    string `json:"key"    mapstructure:"key"    yaml:"key"`
	cache  map[string]any
}

func (c *Config) ejsonDecryptWithKeyTemplateFunc(filePath, key string) any {
	if data, ok := c.Ejson.cache[filePath]; ok {
		return data
	}

	if c.Ejson.cache == nil {
		c.Ejson.cache = make(map[string]any)
	}

	decrypted := mustValue(ejson.DecryptFile(filePath, c.Ejson.KeyDir, key))

	var data any
	must(json.Unmarshal(decrypted, &data))

	c.Ejson.cache[filePath] = data

	return data
}

func (c *Config) ejsonDecryptTemplateFunc(filePath string) any {
	return c.ejsonDecryptWithKeyTemplateFunc(filePath, c.Ejson.Key)
}
