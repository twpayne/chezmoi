package cmd

import (
	"encoding/json"

	"github.com/Shopify/ejson"
)

type ejsonConfig struct {
	cache  map[string]any
	KeyDir string `json:"keyDir" mapstructure:"keyDir" yaml:"keyDir"`
	Key    string `json:"key"    mapstructure:"key"    yaml:"key"`
}

func (c *Config) ejsonDecryptWithKeyTemplateFunc(filePath, key string) any {
	if data, ok := c.Ejson.cache[filePath]; ok {
		return data
	}

	if c.Ejson.cache == nil {
		c.Ejson.cache = make(map[string]any)
	}

	decrypted, err := ejson.DecryptFile(filePath, c.Ejson.KeyDir, key)
	if err != nil {
		panic(err)
	}

	var data any
	if err := json.Unmarshal(decrypted, &data); err != nil {
		panic(err)
	}

	c.Ejson.cache[filePath] = data
	return data
}

func (c *Config) ejsonDecryptTemplateFunc(filePath string) any {
	return c.ejsonDecryptWithKeyTemplateFunc(filePath, c.Ejson.Key)
}
