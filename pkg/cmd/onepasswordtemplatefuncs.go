package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type onepasswordConfig struct {
	Command       string
	Prompt        bool
	outputCache   map[string][]byte
	sessionTokens map[string]string
}

type onePasswordItem struct {
	Details struct {
		Fields   []map[string]interface{} `json:"fields"`
		Sections []struct {
			Fields []map[string]interface{} `json:"fields,omitempty"`
		} `json:"sections"`
	} `json:"details"`
}

func (c *Config) onepasswordTemplateFunc(args ...string) map[string]interface{} {
	sessionToken, err := c.onepasswordGetOrRefreshSessionToken(args)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	onepasswordArgs, err := onepasswordArgs([]string{"get", "item"}, args)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	output, err := c.onepasswordOutput(onepasswordArgs, sessionToken)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(c.Onepassword.Command, onepasswordArgs), err, output))
		return nil
	}
	return data
}

func (c *Config) onepasswordDetailsFieldsTemplateFunc(args ...string) map[string]interface{} {
	onepasswordItem, err := c.onepasswordItem(args...)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	result := make(map[string]interface{})
	for _, field := range onepasswordItem.Details.Fields {
		if designation, ok := field["designation"].(string); ok {
			result[designation] = field
		}
	}
	return result
}

func (c *Config) onepasswordDocumentTemplateFunc(args ...string) string {
	sessionToken, err := c.onepasswordGetOrRefreshSessionToken(args)
	if err != nil {
		returnTemplateError(err)
		return ""
	}

	onepasswordArgs, err := onepasswordArgs([]string{"get", "document"}, args)
	if err != nil {
		returnTemplateError(err)
		return ""
	}

	output, err := c.onepasswordOutput(onepasswordArgs, sessionToken)
	if err != nil {
		returnTemplateError(err)
		return ""
	}
	return string(output)
}

func (c *Config) onepasswordItemFieldsTemplateFunc(args ...string) map[string]interface{} {
	onepasswordItem, err := c.onepasswordItem(args...)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	result := make(map[string]interface{})
	for _, section := range onepasswordItem.Details.Sections {
		for _, field := range section.Fields {
			if t, ok := field["t"].(string); ok {
				result[t] = field
			}
		}
	}
	return result
}

// onepasswordGetOrRefreshSessionToken will return the current session token if
// the token within the environment is still valid. Otherwise it will ask the
// user to sign in and get the new token.
func (c *Config) onepasswordGetOrRefreshSessionToken(callerArgs []string) (string, error) {
	if !c.Onepassword.Prompt {
		return "", nil
	}

	var account string
	if len(callerArgs) > 2 {
		account = callerArgs[2]
	}

	// Check if there's already a valid token cached in this run for this
	// account.
	token, ok := c.Onepassword.sessionTokens[account]
	if ok {
		return token, nil
	}

	var args []string
	if account == "" {
		// If no account has been given then look for any session tokens in the
		// environment.
		token = onepasswordSessionToken()
		args = []string{"signin", "--raw"}
	} else {
		token = os.Getenv("OP_SESSION_" + account)
		args = []string{"signin", account, "--raw"}
	}

	// Do not specify an empty session string if no session tokens were found.
	var secretArgs []string
	if token != "" {
		secretArgs = []string{"--session", token}
	}

	name := c.Onepassword.Command
	// Append the session token here, so it is not logged by accident.
	cmd := exec.Command(name, append(secretArgs, args...)...)
	cmd.Stdin = c.stdin
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	output, err := cmd.Output()
	if err != nil {
		commandStr := shellQuoteCommand(c.Onepassword.Command, args)
		return "", fmt.Errorf("%s: %w: %s", commandStr, err, bytes.TrimSpace(stderr.Bytes()))
	}
	token = strings.TrimSpace(string(output))

	// Cache the session token in memory, so we don't try to refresh it again
	// for this run for this account.
	if c.Onepassword.sessionTokens == nil {
		c.Onepassword.sessionTokens = make(map[string]string)
	}
	c.Onepassword.sessionTokens[account] = token

	return token, nil
}

func (c *Config) onepasswordItem(args ...string) (*onePasswordItem, error) {
	sessionToken, err := c.onepasswordGetOrRefreshSessionToken(args)
	if err != nil {
		return nil, err
	}

	onepasswordArgs, err := onepasswordArgs([]string{"get", "item"}, args)
	if err != nil {
		return nil, err
	}

	output, err := c.onepasswordOutput(onepasswordArgs, sessionToken)
	if err != nil {
		return nil, err
	}

	var onepasswordItem onePasswordItem
	if err := json.Unmarshal(output, &onepasswordItem); err != nil {
		return nil, fmt.Errorf("%s: %w\n%s", shellQuoteCommand(c.Onepassword.Command, onepasswordArgs), err, output)
	}
	return &onepasswordItem, nil
}

func (c *Config) onepasswordOutput(args []string, sessionToken string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if output, ok := c.Onepassword.outputCache[key]; ok {
		return output, nil
	}

	var secretArgs []string
	if sessionToken != "" {
		secretArgs = []string{"--session", sessionToken}
	}

	name := c.Onepassword.Command
	// Append the session token here, so it is not logged by accident.
	cmd := exec.Command(name, append(secretArgs, args...)...)
	cmd.Stdin = c.stdin
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", shellQuoteCommand(name, args), err, bytes.TrimSpace(stderr.Bytes()))
	}

	if c.Onepassword.outputCache == nil {
		c.Onepassword.outputCache = make(map[string][]byte)
	}
	c.Onepassword.outputCache[key] = output

	return output, nil
}

func onepasswordArgs(baseArgs, args []string) ([]string, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(args))
	}
	baseArgs = append(baseArgs, args[0])
	if len(args) > 1 {
		baseArgs = append(baseArgs, "--vault", args[1])
	}
	if len(args) > 2 {
		baseArgs = append(baseArgs, "--account", args[2])
	}
	return baseArgs, nil
}

// onepasswordSessionToken will look for any session tokens in the environment.
// If it finds exactly one then it will return it.
func onepasswordSessionToken() string {
	var token string
	for _, env := range os.Environ() {
		key, value, found := chezmoi.CutString(env, "=")
		if found && strings.HasPrefix(key, "OP_SESSION_") {
			if token != "" {
				return ""
			}
			token = value
		}
	}
	return token
}
