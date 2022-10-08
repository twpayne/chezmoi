package chezmoi

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-vfs/v4/vfst"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestKernel(t *testing.T) {
	for _, tc := range []struct {
		name           string
		root           any
		expectedKernel map[string]any
	}{
		{
			name: "windows_services_for_linux",
			root: map[string]any{
				"/proc/sys/kernel": map[string]any{
					"osrelease": "4.19.81-microsoft-standard\n",
					"ostype":    "Linux\n",
					"version":   "#1 SMP Debian 5.2.9-2 (2019-08-21)\n",
				},
			},
			expectedKernel: map[string]any{
				"osrelease": "4.19.81-microsoft-standard",
				"ostype":    "Linux",
				"version":   "#1 SMP Debian 5.2.9-2 (2019-08-21)",
			},
		},
		{
			name: "debian_version_only",
			root: map[string]any{
				"/proc/sys/kernel": map[string]any{
					"version": "#1 SMP Debian 5.2.9-2 (2019-08-21)\n",
				},
			},
			expectedKernel: map[string]any{
				"version": "#1 SMP Debian 5.2.9-2 (2019-08-21)",
			},
		},
		{
			name: "proc_sys_kernel_missing",
			root: map[string]any{
				"/proc/sys": &vfst.Dir{Perm: 0o755},
			},
			expectedKernel: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				actual, err := Kernel(fileSystem)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedKernel, actual)
			})
		})
	}
}

func TestOSRelease(t *testing.T) {
	for _, tc := range []struct {
		name     string
		root     map[string]any
		expected map[string]any
	}{
		{
			name: "archlinux",
			root: map[string]any{
				"/usr/lib/os-release": chezmoitest.JoinLines(
					`NAME="Arch Linux"`,
					`PRETTY_NAME="Arch Linux"`,
					`ID=arch`,
					`BUILD_ID=rolling`,
					`ANSI_COLOR="38;2;23;147;209"`,
					`HOME_URL="https://archlinux.org/"`,
					`DOCUMENTATION_URL="https://wiki.archlinux.org/"`,
					`SUPPORT_URL="https://bbs.archlinux.org/"`,
					`BUG_REPORT_URL="https://bugs.archlinux.org/"`,
					`LOGO=archlinux`,
				),
			},
			expected: map[string]any{
				"NAME":              "Arch Linux",
				"PRETTY_NAME":       "Arch Linux",
				"ID":                "arch",
				"BUILD_ID":          "rolling",
				"ANSI_COLOR":        "38;2;23;147;209",
				"HOME_URL":          "https://archlinux.org/",
				"DOCUMENTATION_URL": "https://wiki.archlinux.org/",
				"SUPPORT_URL":       "https://bbs.archlinux.org/",
				"BUG_REPORT_URL":    "https://bugs.archlinux.org/",
				"LOGO":              "archlinux",
			},
		},
		{
			name: "fedora",
			root: map[string]any{
				"/etc/os-release": chezmoitest.JoinLines(
					`NAME=Fedora`,
					`VERSION="17 (Beefy Miracle)"`,
					`ID=fedora`,
					`VERSION_ID=17`,
					`PRETTY_NAME="Fedora 17 (Beefy Miracle)"`,
					`ANSI_COLOR="0;34"`,
					`CPE_NAME="cpe:/o:fedoraproject:fedora:17"`,
					`HOME_URL="https://fedoraproject.org/"`,
					`BUG_REPORT_URL="https://bugzilla.redhat.com/"`,
				),
			},
			expected: map[string]any{
				"NAME":           "Fedora",
				"VERSION":        "17 (Beefy Miracle)",
				"ID":             "fedora",
				"VERSION_ID":     "17",
				"PRETTY_NAME":    "Fedora 17 (Beefy Miracle)",
				"ANSI_COLOR":     "0;34",
				"CPE_NAME":       "cpe:/o:fedoraproject:fedora:17",
				"HOME_URL":       "https://fedoraproject.org/",
				"BUG_REPORT_URL": "https://bugzilla.redhat.com/",
			},
		},
		{
			name: "ubuntu",
			root: map[string]any{
				"/usr/lib/os-release": chezmoitest.JoinLines(
					`NAME="Ubuntu"`,
					`VERSION="18.04.1 LTS (Bionic Beaver)"`,
					`ID=ubuntu`,
					`ID_LIKE=debian`,
					`PRETTY_NAME="Ubuntu 18.04.1 LTS"`,
					`VERSION_ID="18.04"`,
					`HOME_URL="https://www.ubuntu.com/"`,
					`SUPPORT_URL="https://help.ubuntu.com/"`,
					`BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"`,
					`PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"`,
					`VERSION_CODENAME=bionic`,
					`UBUNTU_CODENAME=bionic`,
				),
			},
			expected: map[string]any{
				"NAME":               "Ubuntu",
				"VERSION":            "18.04.1 LTS (Bionic Beaver)",
				"ID":                 "ubuntu",
				"ID_LIKE":            "debian",
				"PRETTY_NAME":        "Ubuntu 18.04.1 LTS",
				"VERSION_ID":         "18.04",
				"HOME_URL":           "https://www.ubuntu.com/",
				"SUPPORT_URL":        "https://help.ubuntu.com/",
				"BUG_REPORT_URL":     "https://bugs.launchpad.net/ubuntu/",
				"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
				"VERSION_CODENAME":   "bionic",
				"UBUNTU_CODENAME":    "bionic",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				actual, err := OSRelease(fileSystem)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
		})
	}
}

func TestParseOSRelease(t *testing.T) {
	for _, tc := range []struct {
		name     string
		s        string
		expected map[string]any
	}{
		{
			name: "fedora",
			s: chezmoitest.JoinLines(
				`NAME=Fedora`,
				`VERSION="17 (Beefy Miracle)"`,
				`ID=fedora`,
				`VERSION_ID=17`,
				`PRETTY_NAME="Fedora 17 (Beefy Miracle)"`,
				`ANSI_COLOR="0;34"`,
				`CPE_NAME="cpe:/o:fedoraproject:fedora:17"`,
				`HOME_URL="https://fedoraproject.org/"`,
				`BUG_REPORT_URL="https://bugzilla.redhat.com/"`,
			),
			expected: map[string]any{
				"NAME":           "Fedora",
				"VERSION":        "17 (Beefy Miracle)",
				"ID":             "fedora",
				"VERSION_ID":     "17",
				"PRETTY_NAME":    "Fedora 17 (Beefy Miracle)",
				"ANSI_COLOR":     "0;34",
				"CPE_NAME":       "cpe:/o:fedoraproject:fedora:17",
				"HOME_URL":       "https://fedoraproject.org/",
				"BUG_REPORT_URL": "https://bugzilla.redhat.com/",
			},
		},
		{
			name: "ubuntu_with_comments",
			s: chezmoitest.JoinLines(
				`NAME="Ubuntu"`,
				`VERSION="18.04.1 LTS (Bionic Beaver)"`,
				`ID=ubuntu`,
				`ID_LIKE=debian`,
				`PRETTY_NAME="Ubuntu 18.04.1 LTS"`,
				`VERSION_ID="18.04"`,
				`HOME_URL="https://www.ubuntu.com/"`,
				`SUPPORT_URL="https://help.ubuntu.com/"`,
				`BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"`,
				`PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"`,
				`# comment`,
				``,
				`  # comment`,
				`VERSION_CODENAME=bionic`,
				`UBUNTU_CODENAME=bionic`,
			),
			expected: map[string]any{
				"NAME":               "Ubuntu",
				"VERSION":            "18.04.1 LTS (Bionic Beaver)",
				"ID":                 "ubuntu",
				"ID_LIKE":            "debian",
				"PRETTY_NAME":        "Ubuntu 18.04.1 LTS",
				"VERSION_ID":         "18.04",
				"HOME_URL":           "https://www.ubuntu.com/",
				"SUPPORT_URL":        "https://help.ubuntu.com/",
				"BUG_REPORT_URL":     "https://bugs.launchpad.net/ubuntu/",
				"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
				"VERSION_CODENAME":   "bionic",
				"UBUNTU_CODENAME":    "bionic",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseOSRelease(bytes.NewBufferString(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
