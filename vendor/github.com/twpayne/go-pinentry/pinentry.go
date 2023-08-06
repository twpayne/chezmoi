// Package pinentry provides a client to GnuPG's pinentry.
//
// See info pinentry.
// See https://www.gnupg.org/related_software/pinentry/index.html.
// See https://www.gnupg.org/documentation/manuals/assuan.pdf.
package pinentry

// FIXME add secure logging mode to avoid logging PIN

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/multierr"
)

// Options.
const (
	OptionAllowExternalPasswordCache = "allow-external-password-cache"
	OptionDefaultOK                  = "default-ok"
	OptionDefaultCancel              = "default-cancel"
	OptionDefaultPrompt              = "default-prompt"
	OptionTTYName                    = "ttyname"
	OptionTTYType                    = "ttytype"
	OptionLCCType                    = "lc-ctype"
)

// Error codes.
const (
	AssuanErrorCodeCancelled = 83886179
)

// An AssuanError is returned when an error is sent over the Assuan protocol.
type AssuanError struct {
	Code        int
	Description string
}

func (e *AssuanError) Error() string {
	return e.Description
}

// An UnexpectedResponseError is returned when an unexpected response is
// received.
type UnexpectedResponseError struct {
	Line string
}

func newUnexpectedResponseError(line []byte) UnexpectedResponseError {
	return UnexpectedResponseError{
		Line: string(line),
	}
}

func (e UnexpectedResponseError) Error() string {
	return fmt.Sprintf("pinentry: unexpected response: %q", e.Line)
}

var errorRx = regexp.MustCompile(`\AERR (\d+) (.*)\z`)

// A QualityFunc evaluates the quality of a password. It should return a value
// between -100 and 100. The absolute value of the return value is used as the
// quality. Negative values turn the quality bar red. The boolean return value
// indicates whether the quality is valid.
type QualityFunc func(string) (int, bool)

// A Client is a pinentry client.
type Client struct {
	binaryName  string
	args        []string
	commands    []string
	process     Process
	qualityFunc QualityFunc
	logger      *zerolog.Logger
}

// A ClientOption sets an option on a Client.
type ClientOption func(*Client)

// WithArgs appends extra arguments to the pinentry command.
func WithArgs(args []string) ClientOption {
	return func(c *Client) {
		c.args = append(c.args, args...)
	}
}

// WithBinaryName sets the name of the pinentry binary name. The default is
// pinentry.
func WithBinaryName(binaryName string) ClientOption {
	return func(c *Client) {
		c.binaryName = binaryName
	}
}

// WithCancel sets the cancel button text.
func WithCancel(cancel string) ClientOption {
	return WithCommandf("SETCANCEL %s", escape(cancel))
}

// WithCommand appends an Assuan command that is sent when the connection is
// established.
func WithCommand(command string) ClientOption {
	return func(c *Client) {
		c.commands = append(c.commands, command)
	}
}

// WithCommandf appends an Assuan command that is sent when the connection is
// established, using fmt.Sprintf to format the command.
func WithCommandf(format string, args ...interface{}) ClientOption {
	command := fmt.Sprintf(format, args...)
	return WithCommand(command)
}

// WithDebug tells the pinentry command to print debug messages.
func WithDebug() ClientOption {
	return func(c *Client) {
		c.args = append(c.args, "--debug")
	}
}

// WithDesc sets the description text.
func WithDesc(desc string) ClientOption {
	return WithCommandf("SETDESC %s", escape(desc))
}

// WithError sets the error text.
func WithError(err string) ClientOption {
	return WithCommandf("SETERROR %s", escape(err))
}

// WithKeyInfo sets a stable key identifier for use with password caching.
func WithKeyInfo(keyInfo string) ClientOption {
	return WithCommandf("SETKEYINFO %s", escape(keyInfo))
}

// WithLogger sets the logger.
func WithLogger(logger *zerolog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithNoGlobalGrab instructs pinentry to only grab the password when the window
// is focused.
func WithNoGlobalGrab() ClientOption {
	return func(c *Client) {
		c.args = append(c.args, "--no-global-grab")
	}
}

// WithNotOK sets the text of the non-affirmative response button.
func WithNotOK(notOK string) ClientOption {
	return WithCommandf("SETNOTOK %s", escape(notOK))
}

// WithOK sets the text of the OK button.
func WithOK(ok string) ClientOption {
	return WithCommandf("SETOK %s", escape(ok))
}

// WithOption sets an option.
func WithOption(option string) ClientOption {
	return WithCommandf("OPTION %s", escape(option))
}

// WithOptions sets multiple options.
func WithOptions(options []string) ClientOption {
	return func(c *Client) {
		for _, option := range options {
			command := fmt.Sprintf("OPTION %s", escape(option))
			c.commands = append(c.commands, command)
		}
	}
}

// WithProcess sets the process.
func WithProcess(process Process) ClientOption {
	return func(c *Client) {
		c.process = process
	}
}

// WithPrompt sets the prompt.
func WithPrompt(prompt string) ClientOption {
	return WithCommandf("SETPROMPT %s", escape(prompt))
}

// WithQualityBar enables the quality bar.
func WithQualityBar(qualityFunc QualityFunc) ClientOption {
	return func(c *Client) {
		c.commands = append(c.commands, "SETQUALITYBAR")
		c.qualityFunc = qualityFunc
	}
}

// WithQualityBarToolTip sets the quality bar tool tip.
func WithQualityBarToolTip(qualityBarTT string) ClientOption {
	return WithCommandf("SETQUALITYBAR_TT %s", escape(qualityBarTT))
}

// WithTimeout sets the timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return WithCommandf("SETTIMEOUT %d", timeout/time.Second)
}

// WithTitle sets the title.
func WithTitle(title string) ClientOption {
	return WithCommandf("SETTITLE %s", escape(title))
}

// NewClient returns a new Client with the given options.
func NewClient(options ...ClientOption) (c *Client, err error) {
	c = &Client{
		binaryName:  "pinentry",
		process:     &execProcess{},
		qualityFunc: func(string) (int, bool) { return 0, false },
	}

	for _, option := range options {
		if option != nil {
			option(c)
		}
	}

	err = c.process.Start(c.binaryName, c.args)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = multierr.Append(err, c.Close())
		}
	}()

	var line []byte
	line, err = c.readLine()
	if err != nil {
		return
	}
	if !isOK(line) {
		err = newUnexpectedResponseError(line)
		return
	}

	for _, command := range c.commands {
		if err = c.command(command); err != nil {
			return
		}
	}

	return c, nil
}

// Close closes the connection to the pinentry process.
func (c *Client) Close() (err error) {
	defer func() {
		err = multierr.Append(err, c.process.Close())
	}()
	if err = c.writeLine("BYE"); err != nil {
		return
	}
	err = c.readOK()
	return
}

// Confirm asks the user for confirmation.
func (c *Client) Confirm(option string) (bool, error) {
	command := "CONFIRM"
	if option != "" {
		command += " " + option
	}
	if err := c.writeLine(command); err != nil {
		return false, err
	}
	switch line, err := c.readLine(); {
	case err != nil:
		return false, err
	case isOK(line):
		return true, nil
	case bytes.Equal(line, []byte("ASSUAN_Not_Confirmed")):
		return false, nil
	default:
		return false, newUnexpectedResponseError(line)
	}
}

// GetPIN gets a PIN from the user. If the user cancels, an error is returned
// which can be tested with IsCancelled.
func (c *Client) GetPIN() (pin string, fromCache bool, err error) {
	if err = c.writeLine("GETPIN"); err != nil {
		return "", false, err
	}
	for {
		var line []byte
		switch line, err = c.readLine(); {
		case err != nil:
			return
		case isOK(line):
			return
		case isData(line):
			pin = getPIN(line[2:])
		case bytes.Equal(line, []byte("S PASSWORD_FROM_CACHE")):
			fromCache = true
		case bytes.HasPrefix(line, []byte("INQUIRE QUALITY ")):
			pin = getPIN(line[16:])
			if quality, ok := c.qualityFunc(pin); ok {
				if quality < -100 {
					quality = -100
				} else if quality > 100 {
					quality = 100
				}
				if err = c.writeLine(fmt.Sprintf("D %d", quality)); err != nil {
					return
				}
				if err = c.writeLine("END"); err != nil {
					return
				}
			} else {
				if err = c.writeLine("CAN"); err != nil {
					return
				}
			}
		default:
			err = newUnexpectedResponseError(line)
			return
		}
	}
}

// command writes a command and reads an OK response.
func (c *Client) command(command string) error {
	if err := c.writeLine(command); err != nil {
		return err
	}
	return c.readOK()
}

// readLine reads a line, ignoring blank lines and comments.
func (c *Client) readLine() ([]byte, error) {
	for {
		line, _, err := c.process.ReadLine()
		if err != nil {
			return nil, err
		}
		if c.logger != nil {
			c.logger.Err(err).Bytes("line", line).Msg("readLine")
		}
		switch {
		case isBlank(line):
		case isComment(line):
		case isError(line):
			return nil, newError(line)
		default:
			return line, err
		}
	}
}

// readOK reads an OK response.
func (c *Client) readOK() error {
	switch line, err := c.readLine(); {
	case err != nil:
		return err
	case isOK(line):
		return nil
	default:
		return newUnexpectedResponseError(line)
	}
}

// writeLine writes a single line.
func (c *Client) writeLine(line string) error {
	_, err := c.process.Write([]byte(line + "\n"))
	if c.logger != nil {
		c.logger.Err(err).Str("line", line).Msg("write")
	}
	return err
}

// IsCancelled returns if the error is operation cancelled.
func IsCancelled(err error) bool {
	var assuanError *AssuanError
	if !errors.As(err, &assuanError) {
		return false
	}
	return assuanError.Code == AssuanErrorCodeCancelled
}

func escape(s string) string {
	bytes := []byte(s)
	escapedBytes := make([]byte, 0, len(bytes))
	for _, b := range bytes {
		switch b {
		case '\n':
			escapedBytes = append(escapedBytes, '%', '0', 'A')
		case '\r':
			escapedBytes = append(escapedBytes, '%', '0', 'D')
		case '%':
			escapedBytes = append(escapedBytes, '%', '2', '5')
		default:
			escapedBytes = append(escapedBytes, b)
		}
	}
	return string(escapedBytes)
}

// getPIN parses a PIN from suffix.
func getPIN(data []byte) string {
	return string(unescape(data))
}

// isBlank returns if line is blank.
func isBlank(line []byte) bool {
	return len(bytes.TrimSpace(line)) == 0
}

// isComment returns if line is a comment.
func isComment(line []byte) bool {
	return bytes.HasPrefix(line, []byte("#"))
}

// isData returns if line is a data line.
func isData(line []byte) bool {
	return bytes.HasPrefix(line, []byte("D "))
}

// isError returns if line is an error.
func isError(line []byte) bool {
	return bytes.HasPrefix(line, []byte("ERR "))
}

// isOK returns if the line is an OK response.
func isOK(line []byte) bool {
	return bytes.HasPrefix(line, []byte("OK"))
}

// isUppercaseHexDigit returns if c is an uppercase hexadecimal digit.
func isUppercaseHexDigit(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'A' <= c && c <= 'F':
		return true
	default:
		return false
	}
}

// newError returns an error parsed from line.
func newError(line []byte) error {
	match := errorRx.FindSubmatch(line)
	if match == nil {
		return newUnexpectedResponseError(line)
	}
	code, _ := strconv.Atoi(string(match[1]))
	return &AssuanError{
		Code:        code,
		Description: string(match[2]),
	}
}

// unescape unescapes data, interpreting invalid escape sequences literally
// rather than returning an error.
//
// This is to work around a bug in pinentry-mac 1.1.1 (and possibly earlier
// versions) which does not escape the PIN in INQUIRE QUALITY messages to the
// client.
func unescape(data []byte) []byte {
	unescapedData := make([]byte, 0, len(data))
	for i := 0; i < len(data); {
		if i < len(data)-2 && data[i] == '%' && isUppercaseHexDigit(data[i+1]) && isUppercaseHexDigit(data[i+2]) {
			c := (uppercaseHexDigitValue(data[i+1]) << 4) + uppercaseHexDigitValue(data[i+2])
			unescapedData = append(unescapedData, c)
			i += 3
		} else {
			unescapedData = append(unescapedData, data[i])
			i++
		}
	}
	return unescapedData
}

// uppercaseHexDigitValue returns the value of the uppercase hexadecimal digit
// c.
func uppercaseHexDigitValue(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'A' <= c && c <= 'F':
		return c - 'A' + 0xA
	default:
		return 0
	}
}
