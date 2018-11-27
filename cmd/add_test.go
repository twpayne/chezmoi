package cmd

import (
	"testing"

	"github.com/twpayne/go-vfs/vfstest"
)

func TestAddCommand(t *testing.T) {
	for _, tc := range []struct {
		name             string
		args             []string
		addCommandConfig AddCommandConfig
		root             interface{}
		tests            interface{}
	}{
		{
			name: "add_first_file",
			args: []string{"/home/jenkins/.bashrc"},
			root: map[string]interface{}{
				"/home/jenkins/.bashrc": "foo",
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi", vfstest.TestIsDir, vfstest.TestModePerm(0700)),
				vfstest.TestPath("/home/jenkins/.chezmoi/dot_bashrc", vfstest.TestModeIsRegular, vfstest.TestContentsString("foo")),
			},
		},
		{
			name: "add_template",
			args: []string{"/home/jenkins/.gitconfig"},
			addCommandConfig: AddCommandConfig{
				Template: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":            &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":   &vfstest.Dir{Perm: 0700},
				"/home/jenkins/.gitconfig": "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/dot_gitconfig.tmpl", vfstest.TestModeIsRegular, vfstest.TestContentsString("[user]\n\tname = {{ .name }}\n\temail = {{ .email }}\n")),
			},
		},
		{
			name: "add_recursive",
			args: []string{"/home/jenkins/.config"},
			addCommandConfig: AddCommandConfig{
				Recursive: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":                             &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":                    &vfstest.Dir{Perm: 0700},
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/dot_config/micro/settings.json", vfstest.TestModeIsRegular, vfstest.TestContentsString("{}")),
			},
		},
		{
			name: "add_nested_directory",
			args: []string{"/home/jenkins/.config/micro/settings.json"},
			root: map[string]interface{}{
				"/home/jenkins":                             &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":                    &vfstest.Dir{Perm: 0700},
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/dot_config/micro/settings.json", vfstest.TestModeIsRegular, vfstest.TestContentsString("{}")),
			},
		},
		{
			name: "add_empty_file",
			args: []string{"/home/jenkins/empty"},
			addCommandConfig: AddCommandConfig{
				Empty: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":          &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi": &vfstest.Dir{Perm: 0700},
				"/home/jenkins/empty":    "",
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/empty_empty", vfstest.TestModeIsRegular, vfstest.TestContents(nil)),
			},
		},
		{
			name: "add_symlink",
			args: []string{"/home/jenkins/foo"},
			root: map[string]interface{}{
				"/home/jenkins":          &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi": &vfstest.Dir{Perm: 0700},
				"/home/jenkins/foo":      &vfstest.Symlink{Target: "bar"},
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/symlink_foo", vfstest.TestModeIsRegular, vfstest.TestContentsString("bar")),
			},
		},
		{
			name: "add_symlink_in_dir_recursive",
			args: []string{"/home/jenkins/foo"},
			addCommandConfig: AddCommandConfig{
				Recursive: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":          &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi": &vfstest.Dir{Perm: 0700},
				"/home/jenkins/foo/bar":  &vfstest.Symlink{Target: "baz"},
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/foo", vfstest.TestIsDir),
				vfstest.TestPath("/home/jenkins/.chezmoi/foo/symlink_bar", vfstest.TestModeIsRegular, vfstest.TestContentsString("baz")),
			},
		},
		{
			name: "add_symlink_with_parent_dir",
			args: []string{"/home/jenkins/foo/bar/baz"},
			root: map[string]interface{}{
				"/home/jenkins":             &vfstest.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":    &vfstest.Dir{Perm: 0700},
				"/home/jenkins/foo/bar/baz": &vfstest.Symlink{Target: "qux"},
			},
			tests: []vfstest.Test{
				vfstest.TestPath("/home/jenkins/.chezmoi/foo", vfstest.TestIsDir),
				vfstest.TestPath("/home/jenkins/.chezmoi/foo/bar", vfstest.TestIsDir),
				vfstest.TestPath("/home/jenkins/.chezmoi/foo/bar/symlink_baz", vfstest.TestModeIsRegular, vfstest.TestContentsString("qux")),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{
				SourceDir:        "/home/jenkins/.chezmoi",
				TargetDir:        "/home/jenkins",
				Umask:            022,
				DryRun:           false,
				Verbose:          true,
				SourceVCSCommand: "git",
				Data: map[string]interface{}{
					"name":  "John Smith",
					"email": "john.smith@company.com",
				},
				Add: tc.addCommandConfig,
			}
			fs, cleanup, err := vfstest.NewTempFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfstest.NewTempFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			if err := c.runAddCommandE(fs, nil, tc.args); err != nil {
				t.Errorf("c.runAddCommandE(fs, nil, %+v) == %v, want <nil>", tc.args, err)
			}
			vfstest.RunTests(t, fs, "", tc.tests)
		})
	}
}
