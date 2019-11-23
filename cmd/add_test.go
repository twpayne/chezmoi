package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/go-vfs/vfst"
)

func TestAddAfterModification(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user":          &vfst.Dir{Perm: 0755},
		"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
		"/home/user/.bashrc": &vfst.File{
			Contents: []byte("# contents of .bashrc\n"),
			Perm:     0600,
		},
	})
	require.NoError(t, err)
	defer cleanup()
	c := &Config{
		fs:        fs,
		mutator:   chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false),
		SourceDir: "/home/user/.chezmoi",
		DestDir:   "/home/user",
		Umask:     022,
	}
	args := []string{"/home/user/.bashrc"}
	assert.NoError(t, c.runAddCmd(nil, args))
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.chezmoi/private_dot_bashrc",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("# contents of .bashrc\n"),
		),
	)
	assert.NoError(t, fs.WriteFile("/home/user/.bashrc", []byte("# new contents of .bashrc\n"), 0600))
	assert.NoError(t, c.runAddCmd(nil, args))
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.chezmoi/private_dot_bashrc",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("# new contents of .bashrc\n"),
		),
	)
}

func TestAddCommand(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		add    addCmdConfig
		follow bool
		root   interface{}
		tests  interface{}
	}{
		{
			name: "add_first_file",
			args: []string{"/home/user/.bashrc"},
			root: map[string]interface{}{
				"/home/user/.bashrc": "foo",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi",
					vfst.TestIsDir,
					vfst.TestModePerm(0700),
				),
				vfst.TestPath("/home/user/.chezmoi/dot_bashrc",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo"),
				),
			},
		},
		{
			name: "add_template",
			args: []string{"/home/user/.gitconfig"},
			add: addCmdConfig{
				options: chezmoi.AddOptions{
					Template:     true,
					AutoTemplate: true,
				},
			},
			root: map[string]interface{}{
				"/home/user":            &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":   &vfst.Dir{Perm: 0700},
				"/home/user/.gitconfig": "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/dot_gitconfig.tmpl",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("[user]\n\tname = {{ .name }}\n\temail = {{ .email }}\n"),
				),
			},
		},
		{
			// Test for PR #393
			// Ensure that auto template generating is disabled by default
			name: "add_autotemplate_off_by_default",
			args: []string{"/home/user/.gitconfig"},
			add: addCmdConfig{
				options: chezmoi.AddOptions{
					Template: true,
				},
			},
			root: map[string]interface{}{
				"/home/user":            &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":   &vfst.Dir{Perm: 0700},
				"/home/user/.gitconfig": "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/dot_gitconfig.tmpl",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("[user]\n\tname = John Smith\n\temail = john.smith@company.com\n"),
				),
			},
		},
		{
			name: "add_recursive",
			args: []string{"/home/user/.config"},
			add: addCmdConfig{
				recursive: true,
			},
			root: map[string]interface{}{
				"/home/user":                             &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":                    &vfst.Dir{Perm: 0700},
				"/home/user/.config/micro/settings.json": "{}",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/dot_config/micro/settings.json",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("{}"),
				),
			},
		},
		{
			name: "add_nested_directory",
			args: []string{"/home/user/.config/micro/settings.json"},
			root: map[string]interface{}{
				"/home/user":                             &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":                    &vfst.Dir{Perm: 0700},
				"/home/user/.config/micro/settings.json": "{}",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/dot_config/micro/settings.json",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("{}"),
				),
			},
		},
		{
			name: "add_exact_dir",
			args: []string{"/home/user/dir"},
			add: addCmdConfig{
				options: chezmoi.AddOptions{
					Exact: true,
				},
			},
			root: map[string]interface{}{
				"/home/user":          &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/user/dir":      &vfst.Dir{Perm: 0755},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/exact_dir",
					vfst.TestIsDir,
				),
			},
		},
		{
			name: "add_exact_dir_recursive",
			args: []string{"/home/user/dir"},
			add: addCmdConfig{
				recursive: true,
				options: chezmoi.AddOptions{
					Exact: true,
				},
			},
			root: map[string]interface{}{
				"/home/user":          &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/user/dir": map[string]interface{}{
					"foo": "bar",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/exact_dir",
					vfst.TestIsDir,
				),
				vfst.TestPath("/home/user/.chezmoi/exact_dir/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bar"),
				),
			},
		},
		{
			name: "add_empty_file",
			args: []string{"/home/user/empty"},
			add: addCmdConfig{
				options: chezmoi.AddOptions{
					Empty: true,
				},
			},
			root: map[string]interface{}{
				"/home/user":          &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/user/empty":    "",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/empty_empty",
					vfst.TestModeIsRegular,
					vfst.TestContents(nil),
				),
			},
		},
		{
			name: "add_symlink",
			args: []string{"/home/user/foo"},
			root: map[string]interface{}{
				"/home/user":          &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/user/foo":      &vfst.Symlink{Target: "bar"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/symlink_foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bar"),
				),
			},
		},
		{
			name:   "add_followed_symlink",
			args:   []string{"/home/user/foo"},
			follow: true,
			root: map[string]interface{}{
				"/home/user":          &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/user/foo":      &vfst.Symlink{Target: "bar"},
				"/home/user/bar":      "bux",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bux"),
				),
			},
		},
		{
			name:   "add_symlink_follow",
			args:   []string{"/home/user/foo"},
			follow: true,
			root: map[string]interface{}{
				"/home/user":               &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":      &vfst.Dir{Perm: 0700},
				"/home/user/.dotfiles/foo": "bar",
				"/home/user/foo":           &vfst.Symlink{Target: ".dotfiles/foo"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bar"),
				),
			},
		},
		{
			name: "add_symlink_no_follow",
			args: []string{"/home/user/foo"},
			root: map[string]interface{}{
				"/home/user":               &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":      &vfst.Dir{Perm: 0700},
				"/home/user/.dotfiles/foo": "bar",
				"/home/user/foo":           &vfst.Symlink{Target: ".dotfiles/foo"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/symlink_foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString(filepath.FromSlash(".dotfiles/foo")),
				),
			},
		},
		{
			name:   "add_symlink_follow_double",
			args:   []string{"/home/user/foo"},
			follow: true,
			root: map[string]interface{}{
				"/home/user":               &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":      &vfst.Dir{Perm: 0700},
				"/home/user/.dotfiles/baz": "qux",
				"/home/user/foo":           &vfst.Symlink{Target: "bar"},
				"/home/user/bar":           &vfst.Symlink{Target: ".dotfiles/baz"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("qux"),
				),
			},
		},
		{
			name: "add_symlink_in_dir_recursive",
			args: []string{"/home/user/foo"},
			add: addCmdConfig{
				recursive: true,
			},
			root: map[string]interface{}{
				"/home/user":          &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/user/foo/bar":  &vfst.Symlink{Target: "baz"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestIsDir,
				),
				vfst.TestPath("/home/user/.chezmoi/foo/symlink_bar",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("baz"),
				),
			},
		},
		{
			name: "add_symlink_with_parent_dir",
			args: []string{"/home/user/foo/bar/baz"},
			root: map[string]interface{}{
				"/home/user":             &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi":    &vfst.Dir{Perm: 0700},
				"/home/user/foo/bar/baz": &vfst.Symlink{Target: "qux"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestIsDir,
				),
				vfst.TestPath("/home/user/.chezmoi/foo/bar",
					vfst.TestIsDir,
				),
				vfst.TestPath("/home/user/.chezmoi/foo/bar/symlink_baz",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("qux"),
				),
			},
		},
		{
			name: "dont_add_ignored_file",
			args: []string{"/home/user/foo"},
			root: map[string]interface{}{
				"/home/user": &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{
					Perm: 0700,
					Entries: map[string]interface{}{
						".chezmoiignore": "foo\n",
					},
				},
				"/home/user/foo": "bar",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dont_add_ignored_file_recursive",
			args: []string{"/home/user/foo"},
			add: addCmdConfig{
				recursive: true,
			},
			root: map[string]interface{}{
				"/home/user": &vfst.Dir{Perm: 0755},
				"/home/user/.chezmoi": &vfst.Dir{
					Perm: 0700,
					Entries: map[string]interface{}{
						"exact_foo/.chezmoiignore": "bar/qux\n",
					},
				},
				"/home/user/foo/bar": map[string]interface{}{
					"baz": "baz",
					"qux": "quz",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi/exact_foo/bar/baz",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("baz"),
				),
				vfst.TestPath("/home/user/.chezmoi/exact_foo/bar/qux",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dest_dir_is_symlink",
			args: []string{"/home/user/foo"},
			root: []interface{}{
				map[string]interface{}{
					"/local/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
					"/local/home/user/foo":      "bar",
				},
				map[string]interface{}{
					"/home/user": &vfst.Symlink{Target: "../local/home/user"},
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.chezmoi",
					vfst.TestIsDir,
				),
				vfst.TestPath("/home/user/.chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bar"),
				),
				vfst.TestPath("/local/home/user/.chezmoi",
					vfst.TestIsDir,
				),
				vfst.TestPath("/local/home/user/.chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bar"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			c := &Config{
				fs:        fs,
				mutator:   chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false),
				SourceDir: "/home/user/.chezmoi",
				DestDir:   "/home/user",
				Follow:    tc.follow,
				Umask:     022,
				SourceVCS: sourceVCSConfig{
					Command: "git",
				},
				Data: map[string]interface{}{
					"name":  "John Smith",
					"email": "john.smith@company.com",
				},
				add: tc.add,
			}
			assert.NoError(t, c.runAddCmd(nil, tc.args))
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}

func TestIssue192(t *testing.T) {
	root := []interface{}{
		map[string]interface{}{
			"/local/home/offbyone": &vfst.Dir{
				Perm: 0750,
				Entries: map[string]interface{}{
					".local/share/chezmoi": &vfst.Dir{Perm: 0700},
					"snoop/.list":          "# contents of .list\n",
				},
			},
		},
		map[string]interface{}{
			"/home/offbyone": &vfst.Symlink{Target: "/local/home/offbyone/"},
		},
	}
	fs, cleanup, err := vfst.NewTestFS(root)
	require.NoError(t, err)
	defer cleanup()
	c := &Config{
		fs:        fs,
		mutator:   chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false),
		SourceDir: "/home/offbyone/.local/share/chezmoi",
		DestDir:   "/home/offbyone",
		Umask:     022,
	}
	assert.NoError(t, c.runAddCmd(nil, []string{"/home/offbyone/snoop/.list"}))
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/local/home/offbyone/.local/share/chezmoi/snoop/dot_list",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("# contents of .list\n"),
		),
	)
}
