//go:build !noupgrade && !windows

package cmd

import (
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v4"
)

func TestConfigGetPackageFilename(t *testing.T) {
	for _, tc := range []struct {
		packageType string
		arch        string
		expected    string
	}{
		{
			packageType: packageTypeAPK,
			arch:        "amd64",
			expected:    "chezmoi_2.0.0_linux_amd64.apk",
		},
		{
			packageType: packageTypeDEB,
			arch:        "386",
			expected:    "chezmoi_2.0.0_linux_i386.deb",
		},
		{
			packageType: packageTypeDEB,
			arch:        "amd64",
			expected:    "chezmoi_2.0.0_linux_amd64.deb",
		},
		{
			packageType: packageTypeDEB,
			arch:        "arm",
			expected:    "chezmoi_2.0.0_linux_armel.deb",
		},
		{
			packageType: packageTypeDEB,
			arch:        "arm64",
			expected:    "chezmoi_2.0.0_linux_arm64.deb",
		},
		{
			packageType: packageTypeRPM,
			arch:        "386",
			expected:    "chezmoi-2.0.0-i686.rpm",
		},
		{
			packageType: packageTypeRPM,
			arch:        "amd64",
			expected:    "chezmoi-2.0.0-x86_64.rpm",
		},
		{
			packageType: packageTypeRPM,
			arch:        "arm",
			expected:    "chezmoi-2.0.0-armhfp.rpm",
		},
		{
			packageType: packageTypeRPM,
			arch:        "arm64",
			expected:    "chezmoi-2.0.0-aarch64.rpm",
		},
		{
			packageType: packageTypeRPM,
			arch:        "ppc64",
			expected:    "chezmoi-2.0.0-ppc64.rpm",
		},
		{
			packageType: packageTypeRPM,
			arch:        "ppc64le",
			expected:    "chezmoi-2.0.0-ppc64le.rpm",
		},
	} {
		t.Run(tc.expected, func(t *testing.T) {
			c := newTestConfig(t, vfs.EmptyFS{})
			version := semver.Must(semver.NewVersion("2.0.0"))
			actual, err := c.getPackageFilename(tc.packageType, version, "linux", tc.arch)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
