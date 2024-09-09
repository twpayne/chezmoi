package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type onepasswordMode string

const (
	onepasswordModeAccount onepasswordMode = "account"
	onepasswordModeConnect onepasswordMode = "connect"
	onepasswordModeService onepasswordMode = "service"
)

type withSessionTokenType bool

const (
	withSessionToken    withSessionTokenType = true
	withoutSessionToken withSessionTokenType = false
)

var (
	onepasswordVersionRx  = regexp.MustCompile(`^(\d+\.\d+\.\d+\S*)`)
	onepasswordMinVersion = semver.Version{Major: 2}
)

type onepasswordAccount struct {
	URL         string `json:"url"`
	Email       string `json:"email"`
	UserUUID    string `json:"user_uuid"`    //nolint:tagliatelle
	AccountUUID string `json:"account_uuid"` //nolint:tagliatelle
	Shorthand   string `json:"shorthand"`
}

type onepasswordConfig struct {
	Command       string          `json:"command" mapstructure:"command" yaml:"command"`
	Prompt        bool            `json:"prompt"  mapstructure:"prompt"  yaml:"prompt"`
	Mode          onepasswordMode `json:"mode"    mapstructure:"mode"    yaml:"mode"`
	outputCache   map[string][]byte
	sessionTokens map[string]string
	accountMap    map[string]string
	accountMapErr error
	modeChecked   bool
}

type onepasswordArgs struct {
	item    string
	vault   string
	account string
	args    []string
}

type onepasswordItem struct {
	Fields []map[string]any `json:"fields"`
}

func (c *Config) onepasswordTemplateFunc(userArgs ...string) map[string]any {
	if err := c.onepasswordCheckMode(); err != nil {
		panic(err)
	}

	args, err := c.newOnepasswordArgs([]string{"item", "get", "--format", "json"}, userArgs)
	if err != nil {
		panic(err)
	}

	output, err := c.onepasswordOutput(args, withSessionToken)
	if err != nil {
		panic(err)
	}

	var data map[string]any
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.Onepassword.Command, args.args, output, err))
	}
	return data
}

func (c *Config) onepasswordDetailsFieldsTemplateFunc(userArgs ...string) map[string]any {
	item, err := c.onepasswordItem(userArgs)
	if err != nil {
		panic(err)
	}

	result := make(map[string]any)
	for _, field := range item.Fields {
		if _, ok := field["section"]; ok {
			continue
		}
		if id, ok := field["id"].(string); ok && id != "" {
			result[id] = field
			continue
		}
		if label, ok := field["label"].(string); ok && label != "" {
			result[label] = field
			continue
		}
	}
	return result
}

func (c *Config) onepasswordDocumentTemplateFunc(userArgs ...string) string {
	if err := c.onepasswordCheckMode(); err != nil {
		panic(err)
	}

	if c.Onepassword.Mode == onepasswordModeConnect {
		panic(fmt.Errorf("onepasswordDocument cannot be used in %s mode", onepasswordModeConnect))
	}

	args, err := c.newOnepasswordArgs([]string{"document", "get"}, userArgs)
	if err != nil {
		panic(err)
	}

	output, err := c.onepasswordOutput(args, withSessionToken)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) onepasswordItemFieldsTemplateFunc(userArgs ...string) map[string]any {
	item, err := c.onepasswordItem(userArgs)
	if err != nil {
		panic(err)
	}

	result := make(map[string]any)
	for _, field := range item.Fields {
		if _, ok := field["section"]; !ok {
			continue
		}
		if label, ok := field["label"].(string); ok {
			result[label] = field
		}
	}
	return result
}

// onepasswordGetOrRefreshSessionToken will return the current session token if
// the token within the environment is still valid. Otherwise it will ask the
// user to sign in and get the new token.
func (c *Config) onepasswordGetOrRefreshSessionToken(args *onepasswordArgs) (string, error) {
	if !c.Onepassword.Prompt {
		return "", nil
	}

	// Check if there's already a valid session token cached in this run for
	// this account.
	sessionToken, ok := c.Onepassword.sessionTokens[args.account]
	if ok {
		return sessionToken, nil
	}

	// If no account has been given then look for any session tokens in the
	// environment.
	if args.account == "" {
		sessionToken = onepasswordUniqueSessionToken(os.Environ())
		if sessionToken != "" {
			return sessionToken, nil
		}
	}

	commandArgs := []string{"signin"}
	if args.account != "" {
		commandArgs = append(commandArgs, "--account", args.account)
	}
	commandArgs = append(commandArgs, "--raw")
	if session := os.Getenv("OP_SESSION_" + args.account); session != "" {
		commandArgs = append(commandArgs, "--session", session)
	}

	cmd := exec.Command(c.Onepassword.Command, commandArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return "", newCmdOutputError(cmd, output, err)
	}
	sessionToken = strings.TrimSpace(string(output))

	// Cache the session token in memory, so we don't try to refresh it again
	// for this run for this account.
	if c.Onepassword.sessionTokens == nil {
		c.Onepassword.sessionTokens = make(map[string]string)
	}
	c.Onepassword.sessionTokens[args.account] = sessionToken

	return sessionToken, nil
}

func (c *Config) onepasswordItem(userArgs []string) (*onepasswordItem, error) {
	args, err := c.newOnepasswordArgs([]string{"item", "get", "--format", "json"}, userArgs)
	if err != nil {
		return nil, err
	}

	output, err := c.onepasswordOutput(args, withSessionToken)
	if err != nil {
		return nil, err
	}

	var item onepasswordItem
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, newParseCmdOutputError(c.Onepassword.Command, args.args, output, err)
	}
	return &item, nil
}

func (c *Config) onepasswordOutput(args *onepasswordArgs, withSessionToken withSessionTokenType) ([]byte, error) {
	key := strings.Join(args.args, "\x00")
	if output, ok := c.Onepassword.outputCache[key]; ok {
		return output, nil
	}

	commandArgs := args.args
	if c.Onepassword.Mode == onepasswordModeAccount && withSessionToken {
		sessionToken, err := c.onepasswordGetOrRefreshSessionToken(args)
		if err != nil {
			return nil, err
		}
		if sessionToken != "" {
			commandArgs = append([]string{"--session", sessionToken}, commandArgs...)
		}
	}

	cmd := exec.Command(c.Onepassword.Command, commandArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Onepassword.outputCache == nil {
		c.Onepassword.outputCache = make(map[string][]byte)
	}
	c.Onepassword.outputCache[key] = output

	return output, nil
}

func (c *Config) onepasswordReadTemplateFunc(url string, args ...string) string {
	if err := c.onepasswordCheckMode(); err != nil {
		panic(err)
	}

	onepasswordArgs := &onepasswordArgs{
		args: []string{"read", "--no-newline", url},
	}

	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		if err := onepasswordCheckInvalidAccountParameters(c.Onepassword.Mode); err != nil {
			panic(err)
		}

		onepasswordArgs.account = c.onepasswordAccount(args[0])
		onepasswordArgs.args = append(onepasswordArgs.args, "--account", onepasswordArgs.account)
	default:
		panic(fmt.Errorf("expected 1..2 arguments, got %d", len(args)+1))
	}

	output, err := c.onepasswordOutput(onepasswordArgs, withSessionToken)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) onepasswordAccount(key string) string {
	// This should not happen, but better to be safe
	if err := onepasswordCheckInvalidAccountParameters(c.Onepassword.Mode); err != nil {
		panic(err)
	}

	accounts, err := c.onepasswordAccounts()
	if err != nil {
		panic(err)
	}

	if account, exists := accounts[key]; exists {
		return account
	}

	panic(fmt.Errorf("no 1Password account found matching %s", key))
}

// onepasswordAccounts returns a map of keys to unique account UUIDs.
func (c *Config) onepasswordAccounts() (map[string]string, error) {
	// This should not happen, but better to be safe
	if err := onepasswordCheckInvalidAccountParameters(c.Onepassword.Mode); err != nil {
		return nil, err
	}

	if c.Onepassword.accountMap != nil || c.Onepassword.accountMapErr != nil {
		return c.Onepassword.accountMap, c.Onepassword.accountMapErr
	}

	args := &onepasswordArgs{
		args: []string{"account", "list", "--format=json"},
	}

	output, err := c.onepasswordOutput(args, withoutSessionToken)
	if err != nil {
		c.Onepassword.accountMapErr = err
		return nil, c.Onepassword.accountMapErr
	}

	var accounts []onepasswordAccount
	if err := json.Unmarshal(output, &accounts); err != nil {
		c.Onepassword.accountMapErr = err
		return nil, c.Onepassword.accountMapErr
	}

	c.Onepassword.accountMap = onepasswordAccountMap(accounts)
	return c.Onepassword.accountMap, c.Onepassword.accountMapErr
}

func (c *Config) newOnepasswordArgs(baseArgs, userArgs []string) (*onepasswordArgs, error) {
	maxArgs := 3
	if c.Onepassword.Mode != onepasswordModeAccount {
		maxArgs = 2
	}

	// `session` and `connect` modes do not support the account parameter. Better
	// to error out early.
	if len(userArgs) < 1 || maxArgs < len(userArgs) {
		if err := onepasswordCheckInvalidAccountParameters(c.Onepassword.Mode); maxArgs < len(userArgs) && err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("expected 1..%d arguments in %s mode, got %d", maxArgs, c.Onepassword.Mode, len(userArgs))
	}

	a := &onepasswordArgs{
		args: baseArgs,
	}

	a.item = userArgs[0]
	a.args = append(a.args, a.item)

	if len(userArgs) > 1 && userArgs[1] != "" {
		a.vault = userArgs[1]
		a.args = append(a.args, "--vault", a.vault)
	}

	if len(userArgs) > 2 && userArgs[2] != "" {
		a.account = c.onepasswordAccount(userArgs[2])
		a.args = append(a.args, "--account", a.account)
	}
	return a, nil
}

// onepasswordAccountMap returns a map of unique IDs to account UUIDs.
func onepasswordAccountMap(accounts []onepasswordAccount) map[string]string {
	// Build a map of keys to account UUIDs.
	accountsMap := make(map[string][]string)
	for _, account := range accounts {
		keys := []string{
			account.URL,
			account.Email,
			account.UserUUID,
			account.AccountUUID,
			account.Shorthand,
		}

		accountName, _, accountNameOk := strings.Cut(account.URL, ".")
		if accountNameOk {
			keys = append(keys, accountName)
		}

		emailName, _, emailNameOk := strings.Cut(account.Email, "@")
		if emailNameOk {
			keys = append(keys, emailName, emailName+"@"+account.URL)
		}

		if accountNameOk && emailNameOk {
			keys = append(keys, emailName+"@"+accountName)
		}

		for _, key := range keys {
			accountsMap[key] = append(accountsMap[key], account.AccountUUID)
		}
	}

	// Select unique, non-empty keys.
	accountMap := make(map[string]string)
	for key, values := range accountsMap {
		if key != "" && len(values) == 1 {
			accountMap[key] = values[0]
		}
	}

	return accountMap
}

// onepasswordUniqueSessionToken will look for any session tokens in the
// environment. If it finds exactly one then it will return it.
func onepasswordUniqueSessionToken(environ []string) string {
	var token string
	for _, env := range environ {
		key, value, found := strings.Cut(env, "=")
		if found && strings.HasPrefix(key, "OP_SESSION_") {
			if token != "" {
				return ""
			}
			token = value
		}
	}
	return token
}

// onepasswordCheckMode verifies that things are set up correctly for the
// 1Password mode.
func (c *Config) onepasswordCheckMode() error {
	if c.Onepassword.modeChecked {
		return nil
	}

	c.Onepassword.modeChecked = true

	switch c.Onepassword.Mode {
	case onepasswordModeAccount:
		if os.Getenv("OP_SERVICE_ACCOUNT_TOKEN") != "" {
			return errors.New("onepassword.mode is account, but OP_SERVICE_ACCOUNT_TOKEN is set")
		}

		if os.Getenv("OP_CONNECT_HOST") != "" && os.Getenv("OP_CONNECT_TOKEN") != "" {
			return errors.New("onepassword.mode is account, but OP_CONNECT_HOST and OP_CONNECT_TOKEN are set")
		}

	case onepasswordModeConnect:
		if os.Getenv("OP_CONNECT_HOST") == "" {
			return errors.New("onepassword.mode is connect, but OP_CONNECT_HOST is not set")
		}

		if os.Getenv("OP_CONNECT_TOKEN") == "" {
			return errors.New("onepassword.mode is connect, but OP_CONNECT_TOKEN is not set")
		}

		if os.Getenv("OP_SERVICE_ACCOUNT_TOKEN") != "" {
			return errors.New("onepassword.mode is connect, but OP_SERVICE_ACCOUNT_TOKEN is set")
		}

	case onepasswordModeService:
		if os.Getenv("OP_SERVICE_ACCOUNT_TOKEN") == "" {
			return errors.New("onepassword.mode is service, but OP_SERVICE_ACCOUNT_TOKEN is not set")
		}

		if os.Getenv("OP_CONNECT_HOST") != "" && os.Getenv("OP_CONNECT_TOKEN") != "" {
			return errors.New("onepassword.mode is service, but OP_CONNECT_HOST and OP_CONNECT_TOKEN are set")
		}
	}

	return nil
}

// onepasswordCheckInvalidAccountParameters returns the error "1Password
// account parameters cannot be used in %s mode" if the provided mode is not
// account mode.
func onepasswordCheckInvalidAccountParameters(mode onepasswordMode) error {
	if mode != onepasswordModeAccount {
		return fmt.Errorf("1Password account parameters cannot be used in %s mode", mode)
	}
	return nil
}
