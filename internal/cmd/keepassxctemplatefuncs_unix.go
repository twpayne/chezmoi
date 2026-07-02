//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/pete-woods/go-expect"

	"chezmoi.io/chezmoi/v2/internal/chezmoilog"
)

var (
	keepassxcAnyResponseRx                               = regexp.MustCompile(`(?m)\A.*\r\n`)
	keepassxcEnterPasswordToUnlockDatabaseRx             = regexp.MustCompile(`^Enter password to unlock .*: `)
	keepassxcPleasePresentOrTouchYourYubiKeyToContinueRx = regexp.MustCompile(
		"^Please present or touch your \\S+ to continue\\.\r\n",
	)
	keepassxcPromptRx = regexp.MustCompile(`^.*> `)
)

// keepassxcOutputOpen returns the output of command and args using an
// interactive connection via keepassxc-cli open command's interactive console.
func (c *Config) keepassxcOutputOpen(command string, args ...string) ([]byte, error) {
	// Create the console if it is not already created.
	if c.Keepassxc.console == nil {
		// Create the console.
		console, err := expect.NewConsole()
		if err != nil {
			return nil, err
		}

		// Start the keepassxc-cli open command.
		cmdArgs := []string{"open"}
		cmdArgs = append(cmdArgs, c.Keepassxc.Args...)
		cmdArgs = append(cmdArgs, c.Keepassxc.Database.String())
		cmd := exec.Command(c.Keepassxc.Command, cmdArgs...)
		env := os.Environ()
		// Ensure prompt is in English.
		env = append(env, "LANGUAGE=en")
		// Reduce injection of terminal control characters.
		env = slices.DeleteFunc(env, func(s string) bool {
			return strings.HasPrefix(s, "TERM=")
		})
		cmd.Env = env
		cmd.Stdin = console.Tty()
		cmd.Stdout = console.Tty()
		cmd.Stderr = console.Tty()
		if err := chezmoilog.LogCmdStart(c.logger, cmd); err != nil {
			return nil, err
		}

		if c.Keepassxc.Prompt {
			// Expect the password or YubiKey response.
			response, err := console.Expect(
				expect.Regexp(keepassxcEnterPasswordToUnlockDatabaseRx),
				expect.Regexp(keepassxcPleasePresentOrTouchYourYubiKeyToContinueRx),
				expect.Regexp(keepassxcAnyResponseRx),
			)
			if err != nil {
				return nil, err
			}

			switch {
			case keepassxcEnterPasswordToUnlockDatabaseRx.MatchString(response):
				// Read the password from the user, if necessary.
				var password string
				if c.Keepassxc.password != "" {
					password = c.Keepassxc.password
				} else {
					password, err = c.readPassword(response, "password")
					if err != nil {
						return nil, err
					}
				}

				// Send the password.
				if _, err := console.SendLine(password); err != nil {
					return nil, err
				}

				// Wait for the end of the password prompt.
				if _, err := console.ExpectString("\r\n"); err != nil {
					return nil, err
				}
			case keepassxcPleasePresentOrTouchYourYubiKeyToContinueRx.MatchString(response):
				if _, err := console.ExpectString("\r\n"); err != nil {
					return nil, err
				}
				if _, err := c.stderr.Write([]byte(strings.TrimSpace(response) + "\n")); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("%q: unexpected response", response)
			}
		}

		// Read the response, e.g "Passwords> ", so we can expect it later.
		response, err := console.Expect(
			expect.Regexp(keepassxcPromptRx),
			expect.Regexp(keepassxcAnyResponseRx),
		)
		if err != nil {
			return nil, err
		}
		if !keepassxcPromptRx.MatchString(response) {
			return nil, fmt.Errorf("%q: unexpected response", response)
		}

		c.Keepassxc.cmd = cmd
		c.Keepassxc.console = console
		c.Keepassxc.promptStr = keepassxcPromptRx.FindString(response)
	}

	// Build the command line. Strings with spaces and other non-word characters
	// need to be quoted.
	quotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, maybeQuote(arg))
	}
	line := strings.Join(append([]string{command}, quotedArgs...), " ")

	// Send the line.
	if _, err := c.Keepassxc.console.SendLine(line); err != nil {
		return nil, err
	}

	// Read everything up to and including the prompt.
	output, err := c.Keepassxc.console.ExpectString(c.Keepassxc.promptStr)
	if err != nil {
		return nil, err
	}
	outputLines := strings.Split(output, "\r\n")

	// Trim the echoed command from the output, which is the first line.
	if len(outputLines) > 0 {
		outputLines = outputLines[1:]
	}

	// Trim the prompt from the output, which is the last line.
	if len(outputLines) > 0 {
		outputLines = outputLines[:len(outputLines)-1]
	}

	return []byte(strings.Join(outputLines, "\r\n")), nil
}
