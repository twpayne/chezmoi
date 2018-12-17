package cmd

import (
	"testing"

	"github.com/twpayne/go-vfs/vfst"
)

func TestAddCommand(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		add   addCommandConfig
		root  interface{}
		tests interface{}
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
			add: addCommandConfig{
				template: true,
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
			add: addCommandConfig{
				recursive: true,
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
			add: addCommandConfig{
				empty: true,
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
			add: addCommandConfig{
				recursive: true,
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
				add: tc.add,
			}
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			if err := c.runAddCommand(fs, nil, tc.args); err != nil {
				t.Errorf("c.runAddCommand(fs, nil, %+v) == %v, want <nil>", tc.args, err)
			}
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}

func TestAddAfterModification(t *testing.T) {
	c := &Config{
		SourceDir: "/home/jenkins/.chezmoi",
		TargetDir: "/home/jenkins",
		Umask:     022,
		DryRun:    false,
		Verbose:   true,
	}
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/jenkins":          &vfst.Dir{Perm: 0755},
		"/home/jenkins/.chezmoi": &vfst.Dir{Perm: 0700},
		"/home/jenkins/.bashrc":  "# contents of .bashrc\n",
	})
	defer cleanup()
	if err != nil {
		t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
	}
	args := []string{"/home/jenkins/.bashrc"}
	if err := c.runAddCommand(fs, nil, args); err != nil {
		t.Errorf("c.runAddCommand(fs, nil, %+v) == %v, want <nil>", args, err)
	}
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/jenkins/.chezmoi/dot_bashrc",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("# contents of .bashrc\n"),
		),
	)
	if err := fs.WriteFile("/home/jenkins/.bashrc", []byte("# new contents of .bashrc\n"), 0644); err != nil {
		t.Errorf("fs.WriteFile(...) == %v, want <nil>", err)
	}
	if err := c.runAddCommand(fs, nil, args); err != nil {
		t.Errorf("c.runAddCommand(fs, nil, %+v) == %v, want <nil>", args, err)
	}
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/jenkins/.chezmoi/dot_bashrc",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("# new contents of .bashrc\n"),
		),
	)
}
