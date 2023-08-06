// Package knownhosts is a thin wrapper around golang.org/x/crypto/ssh/knownhosts,
// adding the ability to obtain the list of host key algorithms for a known host.
package knownhosts

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"sort"
	"strings"

	"golang.org/x/crypto/ssh"
	xknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyCallback wraps ssh.HostKeyCallback with an additional method to
// perform host key algorithm lookups from the known_hosts entries.
type HostKeyCallback ssh.HostKeyCallback

// New creates a host key callback from the given OpenSSH host key files. The
// returned value may be used in ssh.ClientConfig.HostKeyCallback by casting it
// to ssh.HostKeyCallback, or using its HostKeyCallback method. Otherwise, it
// operates the same as the New function in golang.org/x/crypto/ssh/knownhosts.
func New(files ...string) (HostKeyCallback, error) {
	cb, err := xknownhosts.New(files...)
	return HostKeyCallback(cb), err
}

// HostKeyCallback simply casts the receiver back to ssh.HostKeyCallback, for
// use in ssh.ClientConfig.HostKeyCallback.
func (hkcb HostKeyCallback) HostKeyCallback() ssh.HostKeyCallback {
	return ssh.HostKeyCallback(hkcb)
}

// HostKeys returns a slice of known host public keys for the supplied host:port
// found in the known_hosts file(s), or an empty slice if the host is not
// already known. For hosts that have multiple known_hosts entries (for
// different key types), the result will be sorted by known_hosts filename and
// line number.
func (hkcb HostKeyCallback) HostKeys(hostWithPort string) (keys []ssh.PublicKey) {
	var keyErr *xknownhosts.KeyError
	placeholderAddr := &net.TCPAddr{IP: []byte{0, 0, 0, 0}}
	placeholderPubKey := &fakePublicKey{}
	var kkeys []xknownhosts.KnownKey
	if hkcbErr := hkcb(hostWithPort, placeholderAddr, placeholderPubKey); errors.As(hkcbErr, &keyErr) {
		kkeys = append(kkeys, keyErr.Want...)
		knownKeyLess := func(i, j int) bool {
			if kkeys[i].Filename < kkeys[j].Filename {
				return true
			}
			return (kkeys[i].Filename == kkeys[j].Filename && kkeys[i].Line < kkeys[j].Line)
		}
		sort.Slice(kkeys, knownKeyLess)
		keys = make([]ssh.PublicKey, len(kkeys))
		for n := range kkeys {
			keys[n] = kkeys[n].Key
		}
	}
	return keys
}

// HostKeyAlgorithms returns a slice of host key algorithms for the supplied
// host:port found in the known_hosts file(s), or an empty slice if the host
// is not already known. The result may be used in ssh.ClientConfig's
// HostKeyAlgorithms field, either as-is or after filtering (if you wish to
// ignore or prefer particular algorithms). For hosts that have multiple
// known_hosts entries (for different key types), the result will be sorted by
// known_hosts filename and line number.
func (hkcb HostKeyCallback) HostKeyAlgorithms(hostWithPort string) (algos []string) {
	for _, key := range hkcb.HostKeys(hostWithPort) {
		algos = append(algos, key.Type())
	}
	return algos
}

// HostKeyAlgorithms is a convenience function for performing host key algorithm
// lookups on an ssh.HostKeyCallback directly. It is intended for use in code
// paths that stay with the New method of golang.org/x/crypto/ssh/knownhosts
// rather than this package's New method.
func HostKeyAlgorithms(cb ssh.HostKeyCallback, hostWithPort string) []string {
	return HostKeyCallback(cb).HostKeyAlgorithms(hostWithPort)
}

// IsHostKeyChanged returns a boolean indicating whether the error indicates
// the host key has changed. It is intended to be called on the error returned
// from invoking a HostKeyCallback to check whether an SSH host is known.
func IsHostKeyChanged(err error) bool {
	var keyErr *xknownhosts.KeyError
	return errors.As(err, &keyErr) && len(keyErr.Want) > 0
}

// IsHostUnknown returns a boolean indicating whether the error represents an
// unknown host. It is intended to be called on the error returned from invoking
// a HostKeyCallback to check whether an SSH host is known.
func IsHostUnknown(err error) bool {
	var keyErr *xknownhosts.KeyError
	return errors.As(err, &keyErr) && len(keyErr.Want) == 0
}

// Normalize normalizes an address into the form used in known_hosts. This
// implementation includes a fix for https://github.com/golang/go/issues/53463
// and will omit brackets around ipv6 addresses on standard port 22.
func Normalize(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = "22"
	}
	entry := host
	if port != "22" {
		entry = "[" + entry + "]:" + port
	} else if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		entry = entry[1 : len(entry)-1]
	}
	return entry
}

// Line returns a line to append to the known_hosts files. This implementation
// uses the local patched implementation of Normalize in order to solve
// https://github.com/golang/go/issues/53463.
func Line(addresses []string, key ssh.PublicKey) string {
	var trimmed []string
	for _, a := range addresses {
		trimmed = append(trimmed, Normalize(a))
	}

	return strings.Join([]string{
		strings.Join(trimmed, ","),
		key.Type(),
		base64.StdEncoding.EncodeToString(key.Marshal()),
	}, " ")
}

// WriteKnownHost writes a known_hosts line to writer for the supplied hostname,
// remote, and key. This is useful when writing a custom hostkey callback which
// wraps a callback obtained from knownhosts.New to provide additional
// known_hosts management functionality. The hostname, remote, and key typically
// correspond to the callback's args.
func WriteKnownHost(w io.Writer, hostname string, remote net.Addr, key ssh.PublicKey) error {
	// Always include hostname; only also include remote if it isn't a zero value
	// and doesn't normalize to the same string as hostname.
	addresses := []string{hostname}
	remoteStr := remote.String()
	remoteStrNormalized := Normalize(remoteStr)
	if remoteStrNormalized != "[0.0.0.0]:0" && remoteStrNormalized != Normalize(hostname) {
		addresses = append(addresses, remoteStr)
	}
	line := Line(addresses, key) + "\n"
	_, err := w.Write([]byte(line))
	return err
}

// fakePublicKey is used as part of the work-around for
// https://github.com/golang/go/issues/29286
type fakePublicKey struct{}

func (fakePublicKey) Type() string {
	return "fake-public-key"
}
func (fakePublicKey) Marshal() []byte {
	return []byte("fake public key")
}
func (fakePublicKey) Verify(_ []byte, _ *ssh.Signature) error {
	return errors.New("Verify called on placeholder key")
}
