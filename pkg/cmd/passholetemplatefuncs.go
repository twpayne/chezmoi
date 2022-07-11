package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type passholeConfig struct {
	Command      string
	Args         []string
	Config       chezmoi.AbsPath
	Database     chezmoi.AbsPath
	Keyfile      chezmoi.AbsPath
	NoPassword   bool
	NoCache      bool
	CacheTimeout int
	cache        map[string]string
	password     string
}

func raise(err error) {
	if err != nil {
		panic(err)
	}
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *Config) passholeAssembleArgs(args []string) []string {
	var result []string
	if c.Passhole.Config.Empty() && c.Passhole.Database.Empty() {
		panic(errors.New("passhole.database or passhole.config not set"))
	}
	if !c.Passhole.Config.Empty() {
		result = append([]string{"--config", c.Passhole.Config.String()}, args...)
		return result
	}
	switch {
	case !c.Passhole.Database.Empty():
		result = append([]string{"--database", c.Passhole.Database.String()}, args...)
	case !c.Passhole.Keyfile.Empty():
		result = append([]string{
			"--database",
			c.Passhole.Database.String(), "--keyfile", c.Passhole.Keyfile.String(),
		}, args...)
	default:
		panic(errors.New("passhole.database not set"))
	}
	if c.Passhole.NoCache {
		result = append([]string{"--no-cache"}, result...)
	}
	if c.Passhole.NoPassword {
		result = append([]string{"--no-password"}, result...)
	}
	return result
}

func (c *Config) passholeTemplateFunction(record string) map[string]string {
	home, err := homedir.Dir()
	raise(err)
	defaultConfigPath := home + "/.config/passhole.ini"
	configExists := PathExists(defaultConfigPath)
	var nameArgs []string
	var passArgs []string
	if configExists {
		nameArgs = []string{"show", "--field", "username", record}
		passArgs = []string{"show", "--field", "password", record}
	} else {
		nameArgs = c.passholeAssembleArgs([]string{"show", "--field", "username", record})
		passArgs = c.passholeAssembleArgs([]string{"show", "--field", "password", record})
	}
	username, err := c.passholeOutput(nameArgs)
	raise(err)
	password, err := c.passholeOutput(passArgs)
	raise(err)
	result := make(map[string]string)
	result["UserName"] = username
	result["Password"] = password
	print(result)
	return result
}

func (c *Config) passholeOutput(args []string) (string, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.Passhole.cache[key]; ok {
		return data, nil
	}

	if c.Passhole.password == "" && !c.Passhole.NoPassword {
		print("NoPassword: ")
		print(c.Passhole.NoPassword)
		password, err := c.readPassword("Insert password to unlock Keepass database:")
		raise(err)
		c.Passhole.password = password
	}

	name := c.Passhole.Command
	args = append(c.Passhole.Args, args...)
	cmd := exec.Command(name, args...)
	if c.Passhole.NoPassword {
		cmd.Stdin = os.Stdin
	} else {
		cmd.Stdin = bytes.NewBufferString(c.Passhole.password + "\n")
	}
	cmd.Stderr = os.Stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return "", newCmdOutputError(cmd, output, err)
	}
	if c.Passhole.cache == nil {
		c.Passhole.cache = make(map[string]string)
	}
	result := strings.TrimSuffix(string(output), "\n")
	c.Passhole.cache[key] = result
	return result, nil
}
