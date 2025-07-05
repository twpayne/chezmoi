// lint-commit-messages lints commit messages.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var commitRx = regexp.MustCompile(`\A([0-9a-f]{40}) (chore(?:\([\w\-]+\))?|docs|feat|fix): ([A-Z]|bump)`)

func run() error {
	args := append([]string{"log", "--format=oneline"}, os.Args[1:]...)
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git: %w", err)
	}

	var invalidCommitMessages []string
	for commitMessage := range bytes.Lines(output) {
		if !commitRx.Match(commitMessage) {
			invalidCommitMessages = append(invalidCommitMessages, string(commitMessage))
		}
	}

	if len(invalidCommitMessages) != 0 {
		return fmt.Errorf("invalid commit messages:\n%s", strings.Join(invalidCommitMessages, "\n"))
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
