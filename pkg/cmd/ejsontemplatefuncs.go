package cmd

import (
	"encoding/json"
	"os"

	"github.com/Shopify/ejson"
)

const (
	ejsonDefaultKeyDir = "/opt/ejson/keys"
	ejsonEnvKeyDir     = "EJSON_KEYDIR"
)

type ejsonConfig struct {
	KeyDir string `json:"keyDir" mapstructure:"keyDir" yaml:"keyDir"`
	Key    string `json:"key" mapstructure:"key" yaml:"key"`
	cache  map[string]any
}

func (c *Config) ejsonDecryptWithKeyTemplateFunc(filePath, key string) any {
	if data, ok := c.Ejson.cache[filePath]; ok {
		return data
	}

	if c.Ejson.cache == nil {
		c.Ejson.cache = make(map[string]any)
	}

	/* We accept here that an empty string is considered as if
	   the value was not provided; in reality, someone could
	   provide an empty value, but ejson would then look into
	   the root of the filesystem; this means that this is not
	   a real limitation, since people could simply set '/'
	   if wanting to use the filesystem root (or equivalent) */
	keyDir := c.Ejson.KeyDir
	if keyDir == "" {
		keyDir = os.Getenv(ejsonEnvKeyDir)
	}
	if keyDir == "" {
		keyDir = ejsonDefaultKeyDir
	}

	decrypted, err := ejson.DecryptFile(filePath, keyDir, key)
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
