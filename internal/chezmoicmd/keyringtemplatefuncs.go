package chezmoicmd

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

type keyringKey struct {
	service string
	user    string
}

type keyringData struct {
	cache map[keyringKey]string
}

func (c *Config) keyringTemplateFunc(service, user string) string {
	key := keyringKey{
		service: service,
		user:    user,
	}
	if password, ok := c.keyring.cache[key]; ok {
		return password
	}
	password, err := keyring.Get(service, user)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w", service, user, err))
		return ""
	}

	if c.keyring.cache == nil {
		c.keyring.cache = make(map[keyringKey]string)
	}

	c.keyring.cache[key] = password
	return password
}
