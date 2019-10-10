package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
	"golang.org/x/crypto/ssh/terminal"
)

var keePassXCCmd = &cobra.Command{
	Use:     "keepassxc [args...]",
	Short:   "Execute the KeePassXC CLI (keepassxc-cli)",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runKeePassXCCmd),
}

type keePassXCCmdConfig struct {
	Command  string
	Database string
	Args     []string
}

type keePassXCAttributeCacheKey struct {
	entry     string
	attribute string
}

var (
	keePassXCCache          = make(map[string]map[string]string)
	keePassXCAttributeCache = make(map[keePassXCAttributeCacheKey]string)
	keePassXCPairRegexp     = regexp.MustCompile(`^([^:]+): (.*)$`)
	keePassXCPassword       string
)

func init() {
	config.KeePassXC.Command = "keepassxc-cli"
	config.addTemplateFunc("keepassxc", config.keePassXCFunc)
	config.addTemplateFunc("keepassxcAttribute", config.keePassXCAttributeFunc)

	secretCmd.AddCommand(keePassXCCmd)
}

func (c *Config) runKeePassXCCmd(fs vfs.FS, args []string) error {
	return c.exec(fs, append([]string{c.KeePassXC.Command}, args...))
}

func (c *Config) keePassXCFunc(entry string) map[string]string {
	if data, ok := keePassXCCache[entry]; ok {
		return data
	}
	if c.KeePassXC.Database == "" {
		panic(errors.New("keepassxc: keepassxc.database not set"))
	}
	name := c.KeePassXC.Command
	args := []string{"show"}
	args = append(args, c.KeePassXC.Args...)
	args = append(args, c.KeePassXC.Database, entry)
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := c.runKeePassXCCLICommand(name, args)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, strings.Join(args, " "), err))
	}
	data, err := parseKeyPassXCOutput(output)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, strings.Join(args, " "), err))
	}
	keePassXCCache[entry] = data
	return data
}

func (c *Config) keePassXCAttributeFunc(entry, attribute string) string {
	key := keePassXCAttributeCacheKey{
		entry:     entry,
		attribute: attribute,
	}
	if data, ok := keePassXCAttributeCache[key]; ok {
		return data
	}
	if c.KeePassXC.Database == "" {
		panic(errors.New("keepassxc: keepassxc.database not set"))
	}
	name := c.KeePassXC.Command
	args := []string{"show", "--attributes", attribute, "--quiet"}
	args = append(args, c.KeePassXC.Args...)
	args = append(args, c.KeePassXC.Database, entry)
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := c.runKeePassXCCLICommand(name, args)
	if err != nil {
		panic(fmt.Errorf("keepassxc: %s %s: %w", name, strings.Join(args, " "), err))
	}
	outputStr := strings.TrimSpace(string(output))
	keePassXCAttributeCache[key] = outputStr
	return outputStr
}

func (c *Config) runKeePassXCCLICommand(name string, args []string) ([]byte, error) {
	if keePassXCPassword == "" {
		fmt.Printf("Insert password to unlock %s: ", c.KeePassXC.Database)
		password, err := terminal.ReadPassword(int(os.Stdout.Fd()))
		fmt.Println()
		if err != nil {
			return nil, err
		}
		keePassXCPassword = string(password)
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewBufferString(keePassXCPassword + "\n")
	cmd.Stderr = c.Stderr()
	return cmd.Output()
}

func parseKeyPassXCOutput(output []byte) (map[string]string, error) {
	data := make(map[string]string)
	s := bufio.NewScanner(bytes.NewReader(output))
	for i := 0; s.Scan(); i++ {
		if i == 0 {
			continue
		}
		match := keePassXCPairRegexp.FindStringSubmatch(s.Text())
		if match == nil {
			return nil, fmt.Errorf("cannot parse %q", s.Text())
		}
		data[match[1]] = match[2]
	}
	return data, s.Err()
}
