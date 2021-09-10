//go:build !windows
// +build !windows

package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestFQDNHostname(t *testing.T) {
	for _, tc := range []struct {
		name     string
		root     interface{}
		expected string
	}{
		{
			name: "empty",
		},
		{
			name: "etc_hosts",
			root: map[string]interface{}{
				"/etc/hosts": chezmoitest.JoinLines(
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
			},
			expected: "thishost.mydomain.org",
		},
		{
			name: "etc_hosts_whitespace_and_comments",
			root: map[string]interface{}{
				"/etc/hosts": chezmoitest.JoinLines(
					" \t127.0.1.1 \tthishost.mydomain.org# comment",
				),
			},
			expected: "thishost.mydomain.org",
		},
		{
			name: "etc_hosts_missing_canonical_hostname",
			root: map[string]interface{}{
				"/etc/hosts": chezmoitest.JoinLines(
					`127.0.1.1`,
					`127.0.1.1 thishost.mydomain.org`,
				),
			},
			expected: "thishost.mydomain.org",
		},
		{
			name: "etc_hostname",
			root: map[string]interface{}{
				"/etc/hostname": chezmoitest.JoinLines(
					`# comment`,
					` hostname.example.com # comment`,
				),
			},
			expected: "hostname.example.com",
		},
		{
			name: "etc_hosts_and_etc_hostname",
			root: map[string]interface{}{
				"/etc/hosts":    "127.0.1.1 hostname.example.com hostname\n",
				"/etc/hostname": "hostname\n",
			},
			expected: "hostname.example.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				assert.Equal(t, tc.expected, FQDNHostname(fileSystem))
			})
		})
	}
}

func TestUmask(t *testing.T) {
	require.Equal(t, chezmoitest.Umask, Umask)
}
