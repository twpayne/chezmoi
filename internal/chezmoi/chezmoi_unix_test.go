// +build !windows

package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestEtcHostsFQDNHostname(t *testing.T) {
	for _, tc := range []struct {
		name                string
		etcHostsContentsStr string
		expected            string
	}{
		{
			name: "empty",
		},
		{
			name: "linux_example",
			etcHostsContentsStr: chezmoitest.JoinLines(
				`# The following lines are desirable for IPv4 capable hosts`,
				`127.0.0.1       localhost`,
				``,
				`# 127.0.1.1 is often used for the FQDN of the machine`,
				`127.0.1.1       thishost.mydomain.org  thishost`,
				`192.168.1.10    foo.mydomain.org       foo`,
				`192.168.1.13    bar.mydomain.org       bar`,
				`146.82.138.7    master.debian.org      master`,
				`209.237.226.90  www.opensource.org`,
				``,
				`# The following lines are desirable for IPv6 capable hosts`,
				`::1             localhost ip6-localhost ip6-loopback`,
				`ff02::1         ip6-allnodes`,
				`ff02::2         ip6-allrouters`,
			),
			expected: "thishost.mydomain.org",
		},
		{
			name: "whitespace_and_comments",
			etcHostsContentsStr: chezmoitest.JoinLines(
				" \t127.0.1.1 \tthishost.mydomain.org# comment",
			),
			expected: "thishost.mydomain.org",
		},
		{
			name: "missing_canonical_hostname",
			etcHostsContentsStr: chezmoitest.JoinLines(
				`127.0.1.1`,
				`127.0.1.1 thishost.mydomain.org`,
			),
			expected: "thishost.mydomain.org",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]interface{}{
				"/etc/hosts": tc.etcHostsContentsStr,
			}, func(fs vfs.FS) {
				actual, err := FQDNHostname(fs)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
		})
	}
}

func TestUmask(t *testing.T) {
	require.Equal(t, chezmoitest.Umask, Umask)
}
