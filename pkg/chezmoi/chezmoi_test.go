package chezmoi

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-vfs/v4"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: true,
	})
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

func TestEtcHostsFQDNHostname(t *testing.T) {
	for _, tc := range []struct {
		name     string
		root     any
		f        func(vfs.FS) (string, error)
		expected string
	}{
		{
			name: "etc_hosts",
			root: map[string]any{
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
			f:        etcHostsFQDNHostname,
			expected: "thishost.mydomain.org",
		},
		{
			name: "etc_hosts_whitespace_and_comments",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					" \t127.0.1.1 \tthishost.mydomain.org# comment",
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "thishost.mydomain.org",
		},
		{
			name: "etc_hosts_missing_canonical_hostname",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					`127.0.1.1`,
					`127.0.1.1 thishost.mydomain.org`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "thishost.mydomain.org",
		},
		{
			name: "etc_hostname",
			root: map[string]any{
				"/etc/hostname": chezmoitest.JoinLines(
					`# comment`,
					` hostname.example.com # comment`,
				),
			},
			f:        etcHostnameFQDNHostname,
			expected: "hostname.example.com",
		},
		{
			name: "etc_myname",
			root: map[string]any{
				"/etc/myname": chezmoitest.JoinLines(
					"# comment",
					"",
					"hostname.example.com",
				),
			},
			f:        etcMynameFQDNHostname,
			expected: "hostname.example.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				fqdnHostname, err := tc.f(fileSystem)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, fqdnHostname)
			})
		})
	}
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}
