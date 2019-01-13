package cmd

import (
	"reflect"
	"testing"

	"github.com/twpayne/go-vfs/vfst"
)

func TestChattrCommand(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		root  interface{}
		tests interface{}
	}{
		{
			name: "dir_add_exact",
			args: []string{"+exact", "/home/user/dir"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"dir": &vfst.Dir{Perm: 0755},
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/dir",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/exact_dir",
					vfst.TestIsDir,
				),
			},
		},
		{
			name: "dir_remove_exact",
			args: []string{"-exact", "/home/user/dir"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"exact_dir": &vfst.Dir{Perm: 0755},
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/exact_dir",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/dir",
					vfst.TestIsDir,
				),
			},
		},
		{
			name: "dir_add_private",
			args: []string{"+private", "/home/user/dir"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"dir": &vfst.Dir{Perm: 0755},
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/dir",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/private_dir",
					vfst.TestIsDir,
				),
			},
		},
		{
			name: "dir_remove_private",
			args: []string{"-private", "/home/user/dir"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"private_dir": &vfst.Dir{Perm: 0755},
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/private_dir",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/dir",
					vfst.TestIsDir,
				),
			},
		},
		{
			name: "add_empty",
			args: []string{"+empty", "/home/user/foo"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"foo": "# contents of ~/foo\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/foo",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/empty_foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/foo\n"),
				),
			},
		},
		{
			name: "remove_empty",
			args: []string{"-empty", "/home/user/foo"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"empty_foo": "# contents of ~/foo\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/foo\n"),
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/empty_foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "add_private",
			args: []string{"+private", "/home/user/foo"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"foo": "# contents of ~/foo\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/foo",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/private_foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/foo\n"),
				),
			},
		},
		{
			name: "remove_private",
			args: []string{"-private", "/home/user/foo"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"private_foo": "# contents of ~/foo\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/foo\n"),
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/private_foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "add_template",
			args: []string{"+template", "/home/user/foo"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"foo": "# contents of ~/foo\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/foo",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/foo.tmpl",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/foo\n"),
				),
			},
		},
		{
			name: "remove_template",
			args: []string{"-template", "/home/user/foo"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"foo.tmpl": "# contents of ~/foo\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/foo\n"),
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/foo.tmpl",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "add_template_in_private_dir",
			args: []string{"+template", "/home/user/.ssh/authorized_keys"},
			root: map[string]interface{}{
				"/home/user/.config/share/chezmoi": map[string]interface{}{
					"private_dot_ssh": map[string]interface{}{
						"authorized_keys": "# contents of ~/.ssh/authorized_keys\n",
					},
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.config/share/chezmoi/private_dot_ssh/authorized_keys",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.config/share/chezmoi/private_dot_ssh/authorized_keys.tmpl",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of ~/.ssh/authorized_keys\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{
				SourceDir: "/home/user/.config/share/chezmoi",
				TargetDir: "/home/user",
				Umask:     022,
				Verbose:   true,
			}
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			if err != nil {
				t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			defer cleanup()
			if err := c.runChattrCommand(fs, nil, tc.args); err != nil {
				t.Errorf("c.runChattrCommand(fs, nil, %+v) == %v, want <nil>", tc.args, err)
			}
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}

func TestParseAttributeModifiers(t *testing.T) {
	for _, tc := range []struct {
		s       string
		want    *attributeModifiers
		wantErr bool
	}{
		{s: "empty", want: &attributeModifiers{empty: 1}},
		{s: "+empty", want: &attributeModifiers{empty: 1}},
		{s: "-empty", want: &attributeModifiers{empty: -1}},
		{s: "e", want: &attributeModifiers{empty: 1}},
		{s: "+e", want: &attributeModifiers{empty: 1}},
		{s: "-e", want: &attributeModifiers{empty: -1}},
		{s: "executable", want: &attributeModifiers{executable: 1}},
		{s: "+executable", want: &attributeModifiers{executable: 1}},
		{s: "-executable", want: &attributeModifiers{executable: -1}},
		{s: "x", want: &attributeModifiers{executable: 1}},
		{s: "+x", want: &attributeModifiers{executable: 1}},
		{s: "-x", want: &attributeModifiers{executable: -1}},
		{s: "private", want: &attributeModifiers{private: 1}},
		{s: "+private", want: &attributeModifiers{private: 1}},
		{s: "-private", want: &attributeModifiers{private: -1}},
		{s: "p", want: &attributeModifiers{private: 1}},
		{s: "+p", want: &attributeModifiers{private: 1}},
		{s: "-p", want: &attributeModifiers{private: -1}},
		{s: "template", want: &attributeModifiers{template: 1}},
		{s: "+template", want: &attributeModifiers{template: 1}},
		{s: "-template", want: &attributeModifiers{template: -1}},
		{s: "t", want: &attributeModifiers{template: 1}},
		{s: "+t", want: &attributeModifiers{template: 1}},
		{s: "-t", want: &attributeModifiers{template: -1}},
		{s: "empty,executable,private,template", want: &attributeModifiers{empty: 1, executable: 1, private: 1, template: 1}},
		{s: "+empty,+executable,+private,+template", want: &attributeModifiers{empty: 1, executable: 1, private: 1, template: 1}},
		{s: "-empty,-executable,-private,-template", want: &attributeModifiers{empty: -1, executable: -1, private: -1, template: -1}},
		{s: "foo", wantErr: true},
		{s: "empty,foo", wantErr: true},
		{s: "empty,foo", wantErr: true},
		{s: " empty , -private ", want: &attributeModifiers{empty: 1, private: -1}},
		{s: "empty,,-private", want: &attributeModifiers{empty: 1, private: -1}},
	} {
		if got, gotErr := parseAttributeModifiers(tc.s); (gotErr != nil && !tc.wantErr) || (gotErr == nil && tc.wantErr) || !reflect.DeepEqual(got, tc.want) {
			wantErrStr := "<nil>"
			if tc.wantErr {
				wantErrStr = "!" + wantErrStr
			}
			t.Errorf("parseAttributeModifiers(%q) == %+v, %v, want %+v, %s", tc.s, got, gotErr, tc.want, wantErrStr)
		}
	}
}
