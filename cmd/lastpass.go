package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type lastpassCommandConfig struct {
	lpass string
}

func init() {
	config.lastpass.lpass = "lpass"
	config.addFunc("lastpass", config.lastpassFunc)
}

func (c *Config) lastpassFunc(id string) interface{} {
	// FIXME is there a better way to return errors from template funcs?
	name := c.lastpass.lpass
	args := []string{"show", "-j", id}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return err
	}
	var data []map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		return err
	}
	return data
}
