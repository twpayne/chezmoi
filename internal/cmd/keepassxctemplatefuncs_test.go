package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoitest"
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

type keepassEntry struct {
	database         string
	databasePassword string
	groupName        string
	entryName        string
	entryUsername    string
	entryPassword    string
	attachmentName   string
	attachmentData   string
}

func TestKeepassxcTemplateFuncs(t *testing.T) {
	// Find the path to keepassxc-cli command.
	command, err := exec.LookPath("keepassxc-cli")
	if err != nil {
		t.Skip("keepassxc-cli not found in $PATH")
	}
	assert.NoError(t, err)

	tempDir := t.TempDir()

	// The following test data includes spaces and slashes to test quoting.
	database := filepath.Join(tempDir, "KeePassXC Passwords.kdbx")
	databasePassword := "test / database / password"
	groupName := "test group"
	entryName := groupName + "/test entry"
	entryUsername := "test / username"
	entryPassword := "test / password"
	attachmentName := "test / attachment name"
	attachmentData := "test / attachment data"

	nestedGroupName := "some/nested/group"
	nestedEntryName := nestedGroupName + "/nested entry"
	nestedEntryUsername := "nested / username"
	nestedEntryPassword := "nested / password"
	nestedAttachmentName := "nested / attachment name"
	nestedAttachmentData := "nested / attachment data"

	// Create a KeePassXC database.
	dbCreateCmd := exec.Command(command, "db-create", "--set-password", database)
	dbCreateCmd.Stdin = strings.NewReader(chezmoitest.JoinLines(
		databasePassword,
		databasePassword,
	))
	dbCreateCmd.Stdout = os.Stdout
	dbCreateCmd.Stderr = os.Stderr
	assert.NoError(t, dbCreateCmd.Run())

	createKeepassEntry(t, command, tempDir, keepassEntry{
		database:         database,
		databasePassword: databasePassword,
		groupName:        groupName,
		entryName:        entryName,
		entryUsername:    entryUsername,
		entryPassword:    entryPassword,
		attachmentName:   attachmentName,
		attachmentData:   attachmentData,
	})
	createKeepassEntry(t, command, tempDir, keepassEntry{
		database:         database,
		databasePassword: databasePassword,
		groupName:        nestedGroupName,
		entryName:        nestedEntryName,
		entryUsername:    nestedEntryUsername,
		entryPassword:    nestedEntryPassword,
		attachmentName:   nestedAttachmentName,
		attachmentData:   nestedAttachmentData,
	})

	for _, mode := range []keepassxcMode{
		keepassxcModeBuiltin,
		keepassxcModeCachePassword,
		keepassxcModeOpen,
	} {
		t.Run(string(mode), func(t *testing.T) {
			t.Run("correct_password", func(t *testing.T) {
				config := newTestConfig(t, vfs.OSFS)
				defer config.keepassxcClose()
				config.Keepassxc.Database = chezmoi.NewAbsPath(database)
				config.Keepassxc.Mode = mode
				config.Keepassxc.Prompt = true
				config.Keepassxc.password = databasePassword
				assert.Equal(t, entryPassword, config.keepassxcTemplateFunc(entryName)["Password"])
				assert.Equal(t, entryUsername, config.keepassxcAttributeTemplateFunc(entryName, "UserName"))
				assert.Equal(t, attachmentData, config.keepassxcAttachmentTemplateFunc(entryName, attachmentName))

				assert.Equal(t, nestedEntryPassword, config.keepassxcTemplateFunc(nestedEntryName)["Password"])
				assert.Equal(t, nestedEntryUsername, config.keepassxcAttributeTemplateFunc(nestedEntryName, "UserName"))
				assert.Equal(
					t,
					nestedAttachmentData,
					config.keepassxcAttachmentTemplateFunc(nestedEntryName, nestedAttachmentName),
				)
			})

			t.Run("incorrect_password", func(t *testing.T) {
				config := newTestConfig(t, vfs.OSFS)
				defer config.keepassxcClose()
				config.Keepassxc.Database = chezmoi.NewAbsPath(database)
				config.Keepassxc.Mode = mode
				config.Keepassxc.Prompt = true
				config.Keepassxc.password = "incorrect-" + databasePassword
				assert.Panics(t, func() {
					config.keepassxcTemplateFunc(entryName)
				})
				assert.Panics(t, func() {
					config.keepassxcAttributeTemplateFunc(entryName, "UserName")
				})
				assert.Panics(t, func() {
					config.keepassxcAttachmentTemplateFunc(entryName, attachmentName)
				})
			})

			t.Run("incorrect_database", func(t *testing.T) {
				config := newTestConfig(t, vfs.OSFS)
				defer config.keepassxcClose()
				config.Keepassxc.Database = chezmoi.NewAbsPath(filepath.Join(tempDir, "Non-existent database.kdbx"))
				config.Keepassxc.Mode = mode
				config.Keepassxc.Prompt = true
				config.Keepassxc.password = databasePassword
				assert.Panics(t, func() {
					config.keepassxcTemplateFunc(entryName)
				})
				assert.Panics(t, func() {
					config.keepassxcAttributeTemplateFunc(entryName, "UserName")
				})
				assert.Panics(t, func() {
					config.keepassxcAttachmentTemplateFunc(entryName, attachmentName)
				})
			})
		})
	}
}

func createKeepassEntry(t *testing.T, command, tempDir string, kpe keepassEntry) {
	t.Helper()
	// Create nested groups in the database.
	groupPath := strings.Split(kpe.groupName, "/")
	for i := range groupPath {
		name := strings.Join(groupPath[0:i+1], "/")
		mkdirCmd := exec.Command(command, "mkdir", kpe.database, name)
		mkdirCmd.Stdin = strings.NewReader(kpe.databasePassword + "\n")
		mkdirCmd.Stdout = os.Stdout
		mkdirCmd.Stderr = os.Stderr
		assert.NoError(t, mkdirCmd.Run())
	}
	// Create an entry in the database.
	addCmd := exec.Command(command, "add", kpe.database, kpe.entryName, "--username", kpe.entryUsername, "--password-prompt")
	addCmd.Stdin = strings.NewReader(chezmoitest.JoinLines(
		kpe.databasePassword,
		kpe.entryPassword,
	))
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	assert.NoError(t, addCmd.Run())

	// Import an attachment to the entry in the database.
	importFile := filepath.Join(tempDir, "import file")
	assert.NoError(t, os.WriteFile(importFile, []byte(kpe.attachmentData), 0o666))
	attachmentImportCmd := exec.Command(
		command,
		"attachment-import",
		kpe.database,
		kpe.entryName,
		kpe.attachmentName,
		importFile,
	)
	attachmentImportCmd.Stdin = strings.NewReader(kpe.databasePassword + "\n")
	attachmentImportCmd.Stdout = os.Stdout
	attachmentImportCmd.Stderr = os.Stderr
	assert.NoError(t, attachmentImportCmd.Run())
}
