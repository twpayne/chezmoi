package cmd

import (
	"fmt"
	"os/exec"

	"howett.net/plist"
)

type ioregData struct {
	value map[string]interface{}
}

func init() {
	config.addTemplateFunc("ioreg", config.ioregFunc)
}

func (c *Config) ioregFunc() map[string]interface{} {
	if c.ioregData.value != nil {
		return c.ioregData.value
	}

	cmd := exec.Command("ioreg", "-a", "-l")
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("ioreg: %w", err))
	}

	var value map[string]interface{}
	if _, err := plist.Unmarshal(output, &value); err != nil {
		panic(fmt.Errorf("ioreg: %w", err))
	}
	c.ioregData.value = value
	return value
}
