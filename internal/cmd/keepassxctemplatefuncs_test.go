package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestKeepassxcParseOutput(t *testing.T) {
	for i, tc := range []struct {
		output   []byte
		expected map[string]string
	}{
		{
			expected: map[string]string{},
		},
		{
			output: []byte(chezmoitest.JoinLines(
				"Title: test",
				"UserName: test",
				"Password: test",
				"URL:",
				"Notes: account: 123456789",
				"2021-11-27 [expires: 2023-02-25]",
				"main = false",
			)),
			expected: map[string]string{
				"Title":    "test",
				"UserName": "test",
				"Password": "test",
				"URL":      "",
				"Notes": strings.Join([]string{
					"account: 123456789",
					"2021-11-27 [expires: 2023-02-25]",
					"main = false",
				}, "\n"),
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, err := keepassxcParseOutput(tc.output)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestKeepassxcTemplateFuncs(t *testing.T) {
	// Find the path to keepassxc-cli command.
	command, err := exec.LookPath("keepassxc-cli")
	if err != nil {
		t.Skip("keepassxc-cli not found in $PATH")
	}
	assert.NoError(t, err)

	tempDir := t.TempDir()
	database := filepath.Join(tempDir, "Passwords.kdbx")
	databasePassword := "test-database-password"
	entryName := "test-entry"
	entryUsername := "test-username"
	entryPassword := "test-password"
	attachmentName := "test-attachment-name"
	attachmentData := "test-attachment-data"
	importFile := filepath.Join(tempDir, "import-file")
	assert.NoError(t, os.WriteFile(importFile, []byte(attachmentData), 0o666))

	// Create a KeePassXC database.
	dbCreateCmd := exec.Command(command, "db-create", "--set-password", database)
	dbCreateCmd.Stdin = strings.NewReader(chezmoitest.JoinLines(
		databasePassword,
		databasePassword,
	))
	dbCreateCmd.Stdout = os.Stdout
	dbCreateCmd.Stderr = os.Stderr
	assert.NoError(t, dbCreateCmd.Run())

	// Create an entry in the database.
	addCmd := exec.Command(command, "add", database, entryName, "--username", entryUsername, "--password-prompt")
	addCmd.Stdin = strings.NewReader(chezmoitest.JoinLines(
		databasePassword,
		entryPassword,
	))
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	assert.NoError(t, addCmd.Run())

	// Import an attachment to the entry in the database.
	attachmentImportCmd := exec.Command(command, "attachment-import", database, entryName, attachmentName, importFile)
	attachmentImportCmd.Stdin = strings.NewReader(chezmoitest.JoinLines(
		databasePassword,
	))
	attachmentImportCmd.Stdout = os.Stdout
	attachmentImportCmd.Stderr = os.Stderr
	assert.NoError(t, attachmentImportCmd.Run())

	for _, mode := range []keepassxcMode{
		keepassxcModeCachePassword,
		keepassxcModeOpen,
	} {
		t.Run(string(mode), func(t *testing.T) {
			config := newTestConfig(t, vfs.OSFS)
			config.Keepassxc.Database = chezmoi.NewAbsPath(database)
			config.Keepassxc.Mode = mode
			config.Keepassxc.Prompt = true
			config.Keepassxc.password = databasePassword

			assert.Equal(t, entryPassword, config.keepassxcTemplateFunc(entryName)["Password"])
			assert.Equal(t, entryUsername, config.keepassxcAttributeTemplateFunc(entryName, "UserName"))
			assert.Equal(t, attachmentData, config.keepassxcAttachmentTemplateFunc(entryName, attachmentName))
		})
	}
}
