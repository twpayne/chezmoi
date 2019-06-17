package cmd

import (
	"strings"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctorBinaryCheck(t *testing.T) {
	for _, tc := range []struct {
		name     string
		check    *doctorBinaryCheck
		output   string
		expected *semver.Version
	}{
		{
			name:  "gnupg_2.2.16",
			check: gpgBinaryCheck,
			output: strings.Join([]string{
				"gpg (GnuPG) 2.2.16",
				"libgcrypt 1.8.4",
				"Copyright (C) 2019 Free Software Foundation, Inc.",
				"License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>",
				"This is free software: you are free to change and redistribute it.",
				"There is NO WARRANTY, to the extent permitted by law.",
				"",
				"Home: /Users/username/.gnupg",
				"Supported algorithms:",
				"Pubkey: RSA, ELG, DSA, ECDH, ECDSA, EDDSA",
				"Cipher: IDEA, 3DES, CAST5, BLOWFISH, AES, AES192, AES256, TWOFISH,",
				"        CAMELLIA128, CAMELLIA192, CAMELLIA256",
				"Hash: SHA1, RIPEMD160, SHA256, SHA384, SHA512, SHA224",
				"Compression: Uncompressed, ZIP, ZLIB, BZIP2",
			}, "\n"),
			expected: semver.New("2.2.16"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.check.getVersionFromOutput([]byte(tc.output))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
