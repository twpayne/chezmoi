//go:build !noupgrade && unix

package cmd

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/coreos/go-semver/semver"
	"github.com/twpayne/go-vfs/v5"
)

func TestGetInstallPackageArgs(t *testing.T) {
	for _, tc := range []struct {
		name           string
		packageType    string
		packageAbsPath string
		useSudo        bool
		expected       []string
	}{
		{
			name:           "apk",
			packageType:    packageTypeAPK,
			packageAbsPath: "/tmp/chezmoi.apk",
			expected:       []string{"apk", "add", "--allow-untrusted", "/tmp/chezmoi.apk"},
		},
		{
			name:           "apk with sudo",
			packageType:    packageTypeAPK,
			packageAbsPath: "/tmp/chezmoi.apk",
			useSudo:        true,
			expected:       []string{"sudo", "apk", "add", "--allow-untrusted", "/tmp/chezmoi.apk"},
		},
		{
			name:           "deb",
			packageType:    packageTypeDEB,
			packageAbsPath: "/tmp/chezmoi.deb",
			expected:       []string{"dpkg", "-i", "/tmp/chezmoi.deb"},
		},
		{
			name:           "rpm",
			packageType:    packageTypeRPM,
			packageAbsPath: "/tmp/chezmoi.rpm",
			expected:       []string{"rpm", "-U", "/tmp/chezmoi.rpm"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := getInstallPackageArgs(tc.packageType, tc.packageAbsPath, tc.useSudo)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

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
			expected:    "chezmoi_2.0.0_linux_armhf.deb",
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
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
