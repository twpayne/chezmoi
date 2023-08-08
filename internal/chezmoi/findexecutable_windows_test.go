//go:build windows

package chezmoi

import (
	"strings"
	"testing"
)

func TestFindExecutable(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		paths   []string
		want    string
		wantErr bool
	}{
		{
			name:    "Finds with extension",
			file:    "powershell.exe",
			paths:   []string{"c:\\windows\\system32", "c:\\windows\\system64", "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0"},
			want:    "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
			wantErr: false,
		},
		{
			name:    "Finds without extension",
			file:    "powershell",
			paths:   []string{"c:\\windows\\system32", "c:\\windows\\system64", "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0"},
			want:    "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
			wantErr: false,
		},
		{
			name:    "Fails to find with extension",
			file:    "weakshell.exe",
			paths:   []string{"c:\\windows\\system32", "c:\\windows\\system64", "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0"},
			want:    "",
			wantErr: false,
		},
		{
			name:    "Fails to find without extension",
			file:    "weakshell",
			paths:   []string{"c:\\windows\\system32", "c:\\windows\\system64", "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0"},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindExecutable(tt.file, tt.paths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindExecutable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.EqualFold(got, tt.want) {
				t.Errorf("FindExecutable() got = %v, want %v", got, tt.want)
			}
		})
	}
}
