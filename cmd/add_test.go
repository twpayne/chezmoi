package cmd

import (
	"testing"

	"github.com/twpayne/go-vfs/vfst"
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
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi", vfst.TestIsDir, vfst.TestModePerm(0700)),
				vfst.TestPath("/home/jenkins/.chezmoi/dot_bashrc", vfst.TestModeIsRegular, vfst.TestContentsString("foo")),
			},
		},
		{
			name: "add_template",
			args: []string{"/home/jenkins/.gitconfig"},
			addCommandConfig: AddCommandConfig{
				Template: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":            &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":   &vfst.Dir{Perm: 0700},
				"/home/jenkins/.gitconfig": "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/dot_gitconfig.tmpl", vfst.TestModeIsRegular, vfst.TestContentsString("[user]\n\tname = {{ .name }}\n\temail = {{ .email }}\n")),
			},
		},
		{
			name: "add_recursive",
			args: []string{"/home/jenkins/.config"},
			addCommandConfig: AddCommandConfig{
				Recursive: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":                             &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":                    &vfst.Dir{Perm: 0700},
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/dot_config/micro/settings.json", vfst.TestModeIsRegular, vfst.TestContentsString("{}")),
			},
		},
		{
			name: "add_nested_directory",
			args: []string{"/home/jenkins/.config/micro/settings.json"},
			root: map[string]interface{}{
				"/home/jenkins":                             &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":                    &vfst.Dir{Perm: 0700},
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/dot_config/micro/settings.json", vfst.TestModeIsRegular, vfst.TestContentsString("{}")),
			},
		},
		{
			name: "add_empty_file",
			args: []string{"/home/jenkins/empty"},
			addCommandConfig: AddCommandConfig{
				Empty: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":          &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/jenkins/empty":    "",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/empty_empty", vfst.TestModeIsRegular, vfst.TestContents(nil)),
			},
		},
		{
			name: "add_symlink",
			args: []string{"/home/jenkins/foo"},
			root: map[string]interface{}{
				"/home/jenkins":          &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/jenkins/foo":      &vfst.Symlink{Target: "bar"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/symlink_foo", vfst.TestModeIsRegular, vfst.TestContentsString("bar")),
			},
		},
		{
			name: "add_symlink_in_dir_recursive",
			args: []string{"/home/jenkins/foo"},
			addCommandConfig: AddCommandConfig{
				Recursive: true,
			},
			root: map[string]interface{}{
				"/home/jenkins":          &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi": &vfst.Dir{Perm: 0700},
				"/home/jenkins/foo/bar":  &vfst.Symlink{Target: "baz"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/foo", vfst.TestIsDir),
				vfst.TestPath("/home/jenkins/.chezmoi/foo/symlink_bar", vfst.TestModeIsRegular, vfst.TestContentsString("baz")),
			},
		},
		{
			name: "add_symlink_with_parent_dir",
			args: []string{"/home/jenkins/foo/bar/baz"},
			root: map[string]interface{}{
				"/home/jenkins":             &vfst.Dir{Perm: 0755},
				"/home/jenkins/.chezmoi":    &vfst.Dir{Perm: 0700},
				"/home/jenkins/foo/bar/baz": &vfst.Symlink{Target: "qux"},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/jenkins/.chezmoi/foo", vfst.TestIsDir),
				vfst.TestPath("/home/jenkins/.chezmoi/foo/bar", vfst.TestIsDir),
				vfst.TestPath("/home/jenkins/.chezmoi/foo/bar/symlink_baz", vfst.TestModeIsRegular, vfst.TestContentsString("qux")),
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
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			if err := c.runAddCommandE(fs, nil, tc.args); err != nil {
				t.Errorf("c.runAddCommandE(fs, nil, %+v) == %v, want <nil>", tc.args, err)
			}
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}
