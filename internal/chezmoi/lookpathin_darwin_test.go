//go:build darwin

package chezmoi

import "testing"

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
			want:    "/bin/sh",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
