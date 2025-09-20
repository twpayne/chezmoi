package cmd

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/chezmoitest"
)

func TestLastpassParseNote(t *testing.T) {
	for i, tc := range []struct {
		note     string
		expected map[string]string
	}{
		{
			note: "Foo:bar\n",
			expected: map[string]string{
				"foo": "bar\n",
			},
		},
		{
			note: chezmoitest.JoinLines(
				"Foo:bar",
				"baz",
			),
			expected: map[string]string{
				"foo": chezmoitest.JoinLines(
					"bar",
					"baz",
				),
			},
		},
		{
			note: chezmoitest.JoinLines(
				"NoteType:SSH Key",
				"Language:en-US",
				"Bit Strength:2048",
				"Format:RSA",
				"Passphrase:Passphrase",
				"Private Key:-----BEGIN OPENSSH PRIVATE KEY-----",
				"-----END OPENSSH PRIVATE KEY-----",
				"Public Key:ssh-rsa public-key you@example",
				"Hostname:Hostname",
				"Date:Date",
			) + "Notes:",
			expected: map[string]string{
				"noteType":    "SSH Key\n",
				"language":    "en-US\n",
				"bitStrength": "2048\n",
				"format":      "RSA\n",
				"passphrase":  "Passphrase\n",
				"privateKey":  "-----BEGIN OPENSSH PRIVATE KEY-----\n-----END OPENSSH PRIVATE KEY-----\n",
				"publicKey":   "ssh-rsa public-key you@example\n",
				"hostname":    "Hostname\n",
				"date":        "Date\n",
				"notes":       "\n",
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expected, lastpassParseNote(tc.note))
		})
	}
}
