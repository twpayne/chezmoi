//go:build !windows && !darwin

package chezmoi

import (
	"os"
	"testing"
)

func TestLookPathIn(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		paths   string
		want    string
		wantErr bool
	}{
		{
			name:    "Finds first",
			file:    "sh",
			paths:   "/usr/bin:/bin",
			want:    "/usr/bin/sh",
			wantErr: false,
		},
		{
			name:    "Finds first 2",
			file:    "sh",
			paths:   "/bin:/usr/bin",
			want:    "/bin/sh",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != "" {
				if _, err := os.Stat(tt.want); err != nil {
					t.Skip("Alpine doesn't have a symlink for sh")
				}
			}
			got, err := LookPathIn(tt.file, tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookPathIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LookPathIn() got = %v, want %v", got, tt.want)
			}
		})
	}
}
