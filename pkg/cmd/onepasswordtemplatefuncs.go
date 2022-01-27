package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type onepasswordConfig struct {
	sync.Mutex
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

func (c *Config) onepasswordItem(args ...string) *onePasswordItem {
	sessionToken := c.onepasswordGetOrRefreshSession(args)
	onepasswordArgs := getOnepasswordArgs([]string{"get", "item"}, args)
	output := c.onepasswordOutput(onepasswordArgs, sessionToken)
	var onepasswordItem onePasswordItem
	if err := json.Unmarshal(output, &onepasswordItem); err != nil {
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(c.Onepassword.Command, onepasswordArgs), err, output))
		return nil
	}
	return &onepasswordItem
}

func (c *Config) onepasswordDetailsFieldsTemplateFunc(args ...string) map[string]interface{} {
	c.Onepassword.Lock()
	defer c.Onepassword.Unlock()

	onepasswordItem := c.onepasswordItem(args...)
	result := make(map[string]interface{})
	for _, field := range onepasswordItem.Details.Fields {
		if designation, ok := field["designation"].(string); ok {
			result[designation] = field
		}
	}
	return result
}

func (c *Config) onepasswordItemFieldsTemplateFunc(args ...string) map[string]interface{} {
	c.Onepassword.Lock()
	defer c.Onepassword.Unlock()

	onepasswordItem := c.onepasswordItem(args...)
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

func (c *Config) onepasswordDocumentTemplateFunc(args ...string) string {
	c.Onepassword.Lock()
	defer c.Onepassword.Unlock()

	sessionToken := c.onepasswordGetOrRefreshSession(args)
	onepasswordArgs := getOnepasswordArgs([]string{"get", "document"}, args)
	output := c.onepasswordOutput(onepasswordArgs, sessionToken)
	return string(output)
}

func (c *Config) onepasswordOutput(args []string, sessionToken string) []byte {
	key := strings.Join(args, "\x00")
	if output, ok := c.Onepassword.outputCache[key]; ok {
		return output
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
		returnTemplateError(fmt.Errorf("%s: %w: %s", shellQuoteCommand(name, args), err, bytes.TrimSpace(stderr.Bytes())))
		return nil
	}

	if c.Onepassword.outputCache == nil {
		c.Onepassword.outputCache = make(map[string][]byte)
	}
	c.Onepassword.outputCache[key] = output
	return output
}

func (c *Config) onepasswordTemplateFunc(args ...string) map[string]interface{} {
	c.Onepassword.Lock()
	defer c.Onepassword.Unlock()

	sessionToken := c.onepasswordGetOrRefreshSession(args)
	onepasswordArgs := getOnepasswordArgs([]string{"get", "item"}, args)
	output := c.onepasswordOutput(onepasswordArgs, sessionToken)
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(c.Onepassword.Command, onepasswordArgs), err, output))
		return nil
	}
	return data
}

func getOnepasswordArgs(baseArgs, args []string) []string {
	if len(args) < 1 || len(args) > 3 {
		returnTemplateError(fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(args)))
		return nil
	}
	baseArgs = append(baseArgs, args[0])
	if len(args) > 1 {
		baseArgs = append(baseArgs, "--vault", args[1])
	}
	if len(args) > 2 {
		baseArgs = append(baseArgs, "--account", args[2])
	}

	return baseArgs
}

// refreshSession will return the current session token if the token within the
// environment is still valid. Otherwise it will ask the user to sign in and get
// the new token. If `sessioncheck` is disabled, it returns an empty string.
func (c *Config) onepasswordGetOrRefreshSession(callerArgs []string) string {
	if !c.Onepassword.Prompt {
		return ""
	}

	var account string
	if len(callerArgs) > 2 {
		account = callerArgs[2]
	}

	// Check if there's already a valid token cached in this run for this
	// account.
	token, ok := c.Onepassword.sessionTokens[account]
	if ok {
		return token
	}

	var args []string
	if account == "" {
		// If no account has been given then look for any session tokens in the
		// environment.
		token = onepasswordInferSessionToken()
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
		returnTemplateError(fmt.Errorf("%s: %w: %s",
			shellQuoteCommand(c.Onepassword.Command, args), err, bytes.TrimSpace(stderr.Bytes())))
		return ""
	}
	token = strings.TrimSpace(string(output))

	// Cache the session token in memory, so we don't try to refresh it again
	// for this run for this account.
	if c.Onepassword.sessionTokens == nil {
		c.Onepassword.sessionTokens = make(map[string]string)
	}
	c.Onepassword.sessionTokens[account] = token

	return token
}

// onepasswordInferSessionToken will look for any session tokens in the
// environment and if it finds exactly one then it will return it.
func onepasswordInferSessionToken() string {
	var token string
	for _, env := range os.Environ() {
		key, value, found := chezmoi.CutString(env, "=")
		if found && strings.HasPrefix(key, "OP_SESSION_") {
			if token != "" {
				// This is the second session we find. Let's bail.
				return ""
			}
			token = value
		}
	}
	return token
}
