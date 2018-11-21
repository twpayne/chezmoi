package cmd

import (
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/twpayne/chezmoi/internal/absfstesting"
)

func TestAddCommand(t *testing.T) {
	for _, tc := range []struct {
		name             string
		args             []string
		addCommandConfig AddCommandConfig
		mapFs            map[string]string
		wantMapFs        map[string]string
	}{
		{
			name: "add_template",
			args: []string{"/home/jenkins/.gitconfig"},
			addCommandConfig: AddCommandConfig{
				Template: true,
			},
			mapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep": "",
				"/home/jenkins/.gitconfig":     "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
			wantMapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep":              "",
				"/home/jenkins/.chezmoi/dot_gitconfig.tmpl": "[user]\n\tname = {{ .name }}\n\temail = {{ .email }}\n",
				"/home/jenkins/.gitconfig":                  "[user]\n\tname = John Smith\n\temail = john.smith@company.com\n",
			},
		},
		{
			name: "add_recursive",
			args: []string{"/home/jenkins/.config"},
			addCommandConfig: AddCommandConfig{
				Recursive: true,
			},
			mapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep":              "",
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			wantMapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep":                          "",
				"/home/jenkins/.chezmoi/dot_config/micro/settings.json": "{}",
				"/home/jenkins/.config/micro/settings.json":             "{}",
			},
		},
		{
			name: "add_nested_directory",
			args: []string{"/home/jenkins/.config/micro/settings.json"},
			mapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep":              "",
				"/home/jenkins/.config/micro/settings.json": "{}",
			},
			wantMapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep":                          "",
				"/home/jenkins/.chezmoi/dot_config/micro/settings.json": "{}",
				"/home/jenkins/.config/micro/settings.json":             "{}",
			},
		},
		{
			name: "add_empty_file",
			args: []string{"/home/jenkins/empty"},
			addCommandConfig: AddCommandConfig{
				Empty: true,
			},
			mapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep": "",
				"/home/jenkins/empty":          "",
			},
			wantMapFs: map[string]string{
				"/home/jenkins/.chezmoi/.keep":       "",
				"/home/jenkins/.chezmoi/empty_empty": "",
				"/home/jenkins/empty":                "",
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
			fs, err := absfstesting.MakeMemMapFs(tc.mapFs)
			if err != nil {
				t.Fatalf("absfstesting.MakeMemMapFs(%+v) == %v, %v, want _, !<nil>", tc.mapFs, fs, err)
			}
			if err := c.runAddCommandE(fs, nil, tc.args); err != nil {
				t.Errorf("c.runAddCommandE(fs, nil, %+v) == %v, want <nil>", tc.args, err)
			}
			mapFs, err := absfstesting.MakeMapFs(fs)
			if err != nil {
				t.Fatalf("absfstesting.MakeMapFs(%+v) == %v, %v, want _, !<nil>", fs, mapFs, err)
			}
			if diff, equal := messagediff.PrettyDiff(tc.wantMapFs, mapFs); !equal {
				t.Errorf("%s\n", diff)
			}
		})
	}
}
