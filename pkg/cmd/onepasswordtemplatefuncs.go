package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type withSessionTokenType bool

const (
	withSessionToken    withSessionTokenType = true
	withoutSessionToken withSessionTokenType = false
)

var onepasswordVersionRx = regexp.MustCompile(`^(\d+\.\d+\.\d+\S*)`)

type onepasswordAccount struct {
	URL         string `json:"url"`
	Email       string `json:"email"`
	UserUUID    string `json:"user_uuid"`    //nolint:tagliatelle
	AccountUUID string `json:"account_uuid"` //nolint:tagliatelle
	Shorthand   string `json:"shorthand"`
}

type onepasswordConfig struct {
	Command       string
	Prompt        bool
	version       *semver.Version
	versionErr    error
	environFunc   func() []string
	outputCache   map[string][]byte
	sessionTokens map[string]string
	accountMap    map[string]string
	accountMapErr error
}

type onepasswordArgs struct {
	item    string
	vault   string
	account string
	args    []string
}

type onepasswordItemV1 struct {
	Details struct {
		Fields   []map[string]any `json:"fields"`
		Sections []struct {
			Fields []map[string]any `json:"fields,omitempty"`
		} `json:"sections"`
	} `json:"details"`
}

type onepasswordItemV2 struct {
	Fields []map[string]any `json:"fields"`
}

func (c *Config) onepasswordTemplateFunc(userArgs ...string) map[string]any {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	var baseArgs []string
	switch {
	case version.Major == 1:
		baseArgs = []string{"get", "item"}
	case version.Major >= 2:
		baseArgs = []string{"item", "get", "--format", "json"}
	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}

	args, err := c.newOnepasswordArgs(baseArgs, userArgs)
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
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	switch {
	case version.Major == 1:
		item, err := c.onepasswordItemV1(userArgs)
		if err != nil {
			panic(err)
		}

		result := make(map[string]any)
		for _, field := range item.Details.Fields {
			if designation, ok := field["designation"].(string); ok {
				result[designation] = field
			}
		}
		return result

	case version.Major >= 2:
		item, err := c.onepasswordItemV2(userArgs)
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

	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}
}

func (c *Config) onepasswordDocumentTemplateFunc(userArgs ...string) string {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	var baseArgs []string
	switch {
	case version.Major == 1:
		baseArgs = []string{"get", "document"}
	case version.Major >= 2:
		baseArgs = []string{"document", "get"}
	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}

	args, err := c.newOnepasswordArgs(baseArgs, userArgs)
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
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	switch {
	case version.Major == 1:
		item, err := c.onepasswordItemV1(userArgs)
		if err != nil {
			panic(err)
		}

		result := make(map[string]any)
		for _, section := range item.Details.Sections {
			for _, field := range section.Fields {
				if t, ok := field["t"].(string); ok {
					result[t] = field
				}
			}
		}
		return result

	case version.Major >= 2:
		item, err := c.onepasswordItemV2(userArgs)
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

	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}
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
		var environ []string
		if c.Onepassword.environFunc != nil {
			environ = c.Onepassword.environFunc()
		} else {
			environ = os.Environ()
		}
		sessionToken = onepasswordUniqueSessionToken(environ)
		if sessionToken != "" {
			return sessionToken, nil
		}
	}

	var commandArgs []string

	if args.account == "" {
		commandArgs = []string{"signin", "--raw"}
	} else {
		sessionToken = os.Getenv("OP_SESSION_" + args.account)

		switch {
		case c.Onepassword.version.Major == 1:
			commandArgs = []string{"signin", args.account, "--raw"}
		case c.Onepassword.version.Major >= 2:
			commandArgs = []string{"signin", "--account", args.account, "--raw"}
		default:
			panic(&unsupportedVersionError{
				version: c.Onepassword.version,
			})
		}
	}

	if sessionToken != "" {
		commandArgs = append([]string{"--session", sessionToken}, commandArgs...)
	}

	cmd := exec.Command(c.Onepassword.Command, commandArgs...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
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

func (c *Config) onepasswordItemV1(userArgs []string) (*onepasswordItemV1, error) {
	args, err := c.newOnepasswordArgs([]string{"get", "item"}, userArgs)
	if err != nil {
		return nil, err
	}

	output, err := c.onepasswordOutput(args, withSessionToken)
	if err != nil {
		return nil, err
	}

	var item onepasswordItemV1
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, newParseCmdOutputError(c.Onepassword.Command, args.args, output, err)
	}
	return &item, nil
}

func (c *Config) onepasswordItemV2(userArgs []string) (*onepasswordItemV2, error) {
	args, err := c.newOnepasswordArgs([]string{"item", "get", "--format", "json"}, userArgs)
	if err != nil {
		return nil, err
	}

	output, err := c.onepasswordOutput(args, withSessionToken)
	if err != nil {
		return nil, err
	}

	var item onepasswordItemV2
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
	if withSessionToken {
		sessionToken, err := c.onepasswordGetOrRefreshSessionToken(args)
		if err != nil {
			return nil, err
		}
		if sessionToken != "" {
			commandArgs = append([]string{"--session", sessionToken}, commandArgs...)
		}
	}

	cmd := exec.Command(c.Onepassword.Command, commandArgs...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
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
	onepasswordArgs := &onepasswordArgs{
		args: []string{"read", url},
	}

	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		onepasswordArgs.args = append(onepasswordArgs.args, "--account", c.onepasswordAccount(args[0]))
	default:
		panic(fmt.Errorf("expected 1 or 2 arguments, got %d", len(args)))
	}

	output, err := c.onepasswordOutput(onepasswordArgs, withSessionToken)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) onepasswordAccount(key string) string {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	// Account listing does not affect version 1
	if version.Major == 1 {
		return key
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

// Shorthand names have been removed from 1Password CLI v2 when biometric
// authentication is used. Mostly, this does not matter. However, this function
// builds a better set of aliases that can be used in the `"account" field`.
// The following values returned from `op account list` will always be mapped
// to the AccountUUID during the actual call.
//
// Given the following values:
//
// ```json
// [
//
//	{
//	  "url": "account1.1password.ca",
//	  "email": "my@email.com",
//	  "user_uuid": "some-user-uuid",
//	  "account_uuid": "some-account-uuid"
//	}
//
// ]
// ```
//
// The following values can be used in the `account` parameter and the value
// `some-account-uuid` will be passed as the `--account` parameter to `op`.
//
// - `some-account-uuid`
// - `some-user-uuid`
// - `account1.1password.ca`
// - `account1`
// - `my@email.com`
// - `my`
// - `my@account1.1password.ca`
// - `my@account1`
//
// If there are multiple accounts and *any* value exists more than once, that
// value will be removed from the account mapping. That is, if you are signed
// into `my@email.com` and `your@email.com` for `account1.1password.ca`, then
// `account1.1password.ca` will not be a valid lookup value, but `my@account1`,
// `my@account1.1password.ca`, `your@account1`, and
// `your@account1.1password.ca` would all be valid lookups.
func (c *Config) onepasswordAccounts() (map[string]string, error) {
	if c.Onepassword.accountMap != nil || c.Onepassword.accountMapErr != nil {
		return c.Onepassword.accountMap, c.Onepassword.accountMapErr
	}

	if version, err := c.onepasswordVersion(); err != nil {
		if version.Major == 1 {
			return make(map[string]string), nil
		}
	}

	args := &onepasswordArgs{
		args: []string{"account", "list", "--format=json"},
	}

	output, err := c.onepasswordOutput(args, withoutSessionToken)
	if err != nil {
		c.Onepassword.accountMapErr = err
		return nil, c.Onepassword.accountMapErr
	}

	var data []onepasswordAccount

	if err := json.Unmarshal(output, &data); err != nil {
		c.Onepassword.accountMapErr = err
		return nil, c.Onepassword.accountMapErr
	}

	collisions := make(map[string]bool)
	result := make(map[string]string)

	for _, account := range data {
		result[account.UserUUID] = account.AccountUUID
		result[account.AccountUUID] = account.AccountUUID

		if _, exists := result[account.URL]; exists {
			collisions[account.URL] = true
		} else {
			result[account.URL] = account.AccountUUID
		}

		parts := strings.SplitN(account.URL, ".", 2)
		accountName := parts[0]

		parts = strings.SplitN(account.Email, "@", 2)
		emailName := parts[0]

		userAccountName := emailName + "@" + accountName
		userAccountURL := emailName + "@" + account.URL

		if _, exists := result[accountName]; exists {
			collisions[accountName] = true
		} else {
			result[accountName] = account.AccountUUID
		}

		if _, exists := result[account.Email]; exists {
			collisions[account.Email] = true
		} else {
			result[account.Email] = account.AccountUUID
		}

		if _, exists := result[emailName]; exists {
			collisions[emailName] = true
		} else {
			result[emailName] = account.AccountUUID
		}

		if _, exists := result[userAccountName]; exists {
			collisions[userAccountName] = true
		} else {
			result[userAccountName] = account.AccountUUID
		}

		if _, exists := result[userAccountURL]; exists {
			collisions[userAccountURL] = true
		} else {
			result[userAccountURL] = account.AccountUUID
		}

		if account.Shorthand != "" {
			if _, exists := result[account.Shorthand]; exists {
				collisions[account.Shorthand] = true
			} else {
				result[account.Shorthand] = account.AccountUUID
			}
		}
	}

	for k := range collisions {
		delete(result, k)
	}

	c.Onepassword.accountMap = result
	return c.Onepassword.accountMap, c.Onepassword.accountMapErr
}

func (c *Config) onepasswordVersion() (*semver.Version, error) {
	if c.Onepassword.version != nil || c.Onepassword.versionErr != nil {
		return c.Onepassword.version, c.Onepassword.versionErr
	}

	args := &onepasswordArgs{
		args: []string{"--version"},
	}
	output, err := c.onepasswordOutput(args, withoutSessionToken)
	if err != nil {
		c.Onepassword.versionErr = err
		return nil, c.Onepassword.versionErr
	}

	m := onepasswordVersionRx.FindSubmatch(output)
	if m == nil {
		c.Onepassword.versionErr = &extractVersionError{
			output: output,
		}
		return nil, c.Onepassword.versionErr
	}

	version, err := semver.NewVersion(string(m[1]))
	if err != nil {
		c.Onepassword.versionErr = &parseVersionError{
			output: m[1],
			err:    err,
		}
		return nil, c.Onepassword.versionErr
	}

	c.Onepassword.version = version
	return c.Onepassword.version, c.Onepassword.versionErr
}

func (c *Config) newOnepasswordArgs(baseArgs, userArgs []string) (*onepasswordArgs, error) {
	if len(userArgs) < 1 || 3 < len(userArgs) {
		return nil, fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(userArgs))
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
