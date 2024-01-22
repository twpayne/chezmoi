package chezmoi

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: true,
	})
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack //nolint:reassign
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
					`127.0.1.1       host.example.com   host`,
					`192.168.1.10    foo.example.com    foo`,
					`192.168.1.13    bar.example.com    bar`,
					`146.82.138.7    master.debian.org  master`,
					`209.237.226.90  www.example.org`,
					``,
					`# The following lines are desirable for IPv6 capable hosts`,
					`::1             localhost ip6-localhost ip6-loopback`,
					`ff02::1         ip6-allnodes`,
					`ff02::2         ip6-allrouters`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
		},
		{
			name: "etc_hosts_loopback_ipv4",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					`invalid localhost`,
					`127.0.0.1 localhost`,
					`::1 localhost`,
					`127.0.0.2 host.example.com host`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
		},
		{
			name: "etc_hosts_loopback_ipv4_localhost_dot_localdomain",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					`127.0.0.1 localhost.localdomain`,
					`127.0.0.2 host.example.com host`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
		},
		{
			name: "etc_hosts_loopback_ipv6",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					`127.0.0.1 localhost`,
					`::1 localhost`,
					`::1 host.example.com host`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
		},
		{
			name: "etc_hosts_whitespace_and_comments",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					" \t127.0.1.1 \thost.example.com# comment",
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
		},
		{
			name: "etc_hosts_missing_canonical_hostname",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					`127.0.1.1`,
					`127.0.1.1 host.example.com`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
		},
		{
			name: "etc_hosts_kubernetes_docker_internal",
			root: map[string]any{
				"/etc/hosts": chezmoitest.JoinLines(
					`##`,
					`# Host Database`,
					`#`,
					`# localhost is used to configure the loopback interface`,
					`# when the system is booting.  Do not change this entry.`,
					`##`,
					`127.0.0.1	localhost`,
					`255.255.255.255	broadcasthost`,
					`::1             localhost`,
					`# Added by Docker Desktop`,
					`# To allow the same kube context to work on the host and the container:`,
					`127.0.0.1 kubernetes.docker.internal`,
					`# End of section`,
					`127.0.1.1`,
					`127.0.1.1 host.example.com`,
				),
			},
			f:        etcHostsFQDNHostname,
			expected: "host.example.com",
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

func TestUniqueAbbreviations(t *testing.T) {
	for _, tc := range []struct {
		values   []string
		expected map[string]string
	}{
		{
			values:   nil,
			expected: map[string]string{},
		},
		{
			values: []string{
				"yes",
				"no",
				"all",
				"quit",
			},
			expected: map[string]string{
				"y":    "yes",
				"ye":   "yes",
				"yes":  "yes",
				"n":    "no",
				"no":   "no",
				"a":    "all",
				"al":   "all",
				"all":  "all",
				"q":    "quit",
				"qu":   "quit",
				"qui":  "quit",
				"quit": "quit",
			},
		},
		{
			values: []string{
				"ale",
				"all",
				"abort",
			},
			expected: map[string]string{
				"ale":   "ale",
				"all":   "all",
				"ab":    "abort",
				"abo":   "abort",
				"abor":  "abort",
				"abort": "abort",
			},
		},
		{
			values: []string{
				"no",
				"now",
				"nope",
			},
			expected: map[string]string{
				"no":   "no",
				"now":  "now",
				"nop":  "nope",
				"nope": "nope",
			},
		},
	} {
		t.Run(strings.Join(tc.values, "_"), func(t *testing.T) {
			actual := UniqueAbbreviations(tc.values)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
