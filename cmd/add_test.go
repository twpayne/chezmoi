package cmd

import (
	"testing"

	"github.com/twpayne/aferot"
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
			tests: []aferot.Test{
				aferot.TestPath("/home/jenkins/.chezmoi", aferot.TestIsDir, aferot.TestModePerm(0700)),
				aferot.TestPath("/home/jenkins/.chezmoi/dot_bashrc", aferot.TestModeIsRegular, aferot.TestContentsString("foo")),
			},
		},
		{
			name: "add_template",
			args: []string{"/home/jenkins/.gitconfig"},
			addCommandConfig: AddCommandConfig{
				Template: true,
			},
			root: map[string]string{
				"/home/jenkins/.chezmoi/.keep": "",
				"/home/jenkins/.gitconfig":     "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
			tests: []aferot.Test{
				aferot.TestPath("/home/jenkins/.chezmoi/dot_gitconfig.tmpl", aferot.TestModeIsRegular, aferot.TestContentsString("[user]\n\tname = {{ .name }}\n\temail = {{ .email }}\n")),
			},
		},
		{
			name: "add_recursive",
			args: []string{"/home/jenkins/.config"},
			addCommandConfig: AddCommandConfig{
				Recursive: true,
			},
			root: map[string]string{
				"/home/jenkins/.chezmoi/.keep":              "",
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			tests: []aferot.Test{
				aferot.TestPath("/home/jenkins/.chezmoi/dot_config/micro/settings.json", aferot.TestModeIsRegular, aferot.TestContentsString("{}")),
			},
		},
		{
			name: "add_nested_directory",
			args: []string{"/home/jenkins/.config/micro/settings.json"},
			root: map[string]string{
				"/home/jenkins/.chezmoi/.keep":              "",
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			tests: []aferot.Test{
				aferot.TestPath("/home/jenkins/.chezmoi/dot_config/micro/settings.json", aferot.TestModeIsRegular, aferot.TestContentsString("{}")),
			},
		},
		{
			name: "add_empty_file",
			args: []string{"/home/jenkins/empty"},
			addCommandConfig: AddCommandConfig{
				Empty: true,
			},
			root: map[string]string{
				"/home/jenkins/.chezmoi/.keep": "",
				"/home/jenkins/empty":          "",
			},
			tests: []aferot.Test{
				aferot.TestPath("/home/jenkins/.chezmoi/empty_empty", aferot.TestModeIsRegular, aferot.TestContents(nil)),
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
			fs, err := aferot.NewMemMapFs(tc.root)
			if err != nil {
				t.Fatalf("aferot.NewMemMapFs(_) == %v, want !<nil>", err)
			}
			if err := c.runAddCommandE(fs, nil, tc.args); err != nil {
				t.Errorf("c.runAddCommandE(fs, nil, %+v) == %v, want <nil>", tc.args, err)
			}
			aferot.RunTest(t, fs, "", tc.tests)
		})
	}
}
