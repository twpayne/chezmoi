//go:build darwin

package chezmoi

import "testing"

func TestFindExecutable(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		paths   []string
		want    string
		wantErr bool
	}{
		{
			name:    "Finds first",
			file:    "sh",
			paths:   []string{"/usr/bin", "/bin"},
			want:    "/bin/sh",
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
			if got != tt.want {
				t.Errorf("FindExecutable() got = %v, want %v", got, tt.want)
			}
		})
	}
}
