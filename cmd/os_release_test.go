package cmd

import (
	"bytes"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/twpayne/go-vfs/vfst"
)

func TestGetOSRelease(t *testing.T) {
	for _, tc := range []struct {
		name string
		root map[string]interface{}
		want map[string]string
	}{
		{
			name: "fedora",
			root: map[string]interface{}{
				"/etc/os-release": `NAME=Fedora
VERSION="17 (Beefy Miracle)"
ID=fedora
VERSION_ID=17
PRETTY_NAME="Fedora 17 (Beefy Miracle)"
ANSI_COLOR="0;34"
CPE_NAME="cpe:/o:fedoraproject:fedora:17"
HOME_URL="https://fedoraproject.org/"
BUG_REPORT_URL="https://bugzilla.redhat.com/"`,
			},
			want: map[string]string{
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
			root: map[string]interface{}{
				"/usr/lib/os-release": `NAME="Ubuntu"
VERSION="18.04.1 LTS (Bionic Beaver)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 18.04.1 LTS"
VERSION_ID="18.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=bionic
UBUNTU_CODENAME=bionic`,
			},
			want: map[string]string{
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
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			if err != nil {
				t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			defer cleanup()
			got, gotErr := getOSRelease(fs)
			diff, equal := messagediff.PrettyDiff(tc.want, got)
			if gotErr != nil || !equal {
				t.Errorf("getOSRelease(fs) == %v, %v, want %v, <nil>\n%s", got, gotErr, tc.want, diff)
			}
		})
	}
}

func TestParseOSRelease(t *testing.T) {
	for _, tc := range []struct {
		s    string
		want map[string]string
	}{
		{
			s: `NAME=Fedora
VERSION="17 (Beefy Miracle)"
ID=fedora
VERSION_ID=17
PRETTY_NAME="Fedora 17 (Beefy Miracle)"
ANSI_COLOR="0;34"
CPE_NAME="cpe:/o:fedoraproject:fedora:17"
HOME_URL="https://fedoraproject.org/"
BUG_REPORT_URL="https://bugzilla.redhat.com/"`,
			want: map[string]string{
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
			s: `NAME="Ubuntu"
VERSION="18.04.1 LTS (Bionic Beaver)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 18.04.1 LTS"
VERSION_ID="18.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=bionic
UBUNTU_CODENAME=bionic`,
			want: map[string]string{
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
		got, gotErr := parseOSRelease(bytes.NewBufferString(tc.s))
		diff, equal := messagediff.PrettyDiff(tc.want, got)
		if gotErr != nil || !equal {
			t.Errorf("parseOSRelease(bytes.NewBufferString(%q)) == %+v, %v, want %+v, <nil>\n%s", tc.s, got, gotErr, tc.want, diff)
		}
	}
}

func TestUppercaseSnakeToCamelCase(t *testing.T) {
	for _, tc := range []struct {
		s    string
		want string
	}{
		{s: "ID", want: "id"},
		{s: "NAME", want: "name"},
		{s: "ID_LIKE", want: "idLike"},
		{s: "PRETTY_NAME", want: "prettyName"},
		{s: "ANSI_COLOR", want: "ansiColor"},
		{s: "BUG_REPORT_URL", want: "bugReportURL"},
	} {
		if got := upperSnakeCaseToCamelCase(tc.s); got != tc.want {
			t.Errorf("uppercaseSnakeToCamelCase(%q) == %q, want %q", tc.s, got, tc.want)
		}
	}
}
