package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatusPorcelainV2(t *testing.T) {
	for _, tc := range []struct {
		name           string
		outputStr      string
		expectedStatus *Status
	}{
		{
			name:      "added",
			outputStr: "1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 cea5c3500651a923bacd80f960dd20f04f71d509 main.go\n",
			expectedStatus: &Status{
				Ordinary: []OrdinaryStatus{
					{
						X:    'A',
						Y:    '.',
						Sub:  "N...",
						MH:   0,
						MI:   0100644,
						MW:   0100644,
						HH:   "0000000000000000000000000000000000000000",
						HI:   "cea5c3500651a923bacd80f960dd20f04f71d509",
						Path: "main.go",
					},
				},
			},
		},
		{
			name:      "removed",
			outputStr: "1 D. N... 100644 000000 000000 cea5c3500651a923bacd80f960dd20f04f71d509 0000000000000000000000000000000000000000 main.go\n",
			expectedStatus: &Status{
				Ordinary: []OrdinaryStatus{
					{
						X:    'D',
						Y:    '.',
						Sub:  "N...",
						MH:   0100644,
						MI:   0,
						MW:   0,
						HH:   "cea5c3500651a923bacd80f960dd20f04f71d509",
						HI:   "0000000000000000000000000000000000000000",
						Path: "main.go",
					},
				},
			},
		},
		{
			name:      "update",
			outputStr: "1 .M N... 100644 100644 100644 353dbbb3c29a80fb44d4e26dac111739d25294db 353dbbb3c29a80fb44d4e26dac111739d25294db cmd/gitvcs.go\n",
			expectedStatus: &Status{
				Ordinary: []OrdinaryStatus{
					{
						X:    '.',
						Y:    'M',
						Sub:  "N...",
						MH:   0100644,
						MI:   0100644,
						MW:   0100644,
						HH:   "353dbbb3c29a80fb44d4e26dac111739d25294db",
						HI:   "353dbbb3c29a80fb44d4e26dac111739d25294db",
						Path: "cmd/gitvcs.go",
					},
				},
			},
		},
		{
			name:      "renamed",
			outputStr: "2 R. N... 100644 100644 100644 9d06c86ecba40e1c695e69b55a40843df6a79cef 9d06c86ecba40e1c695e69b55a40843df6a79cef R100 chezmoi_rename.go chezmoi.go\n",
			expectedStatus: &Status{
				RenamedOrCopied: []RenamedOrCopiedStatus{
					{
						X:        'R',
						Y:        '.',
						Sub:      "N...",
						MH:       0100644,
						MI:       0100644,
						MW:       0100644,
						HH:       "9d06c86ecba40e1c695e69b55a40843df6a79cef",
						HI:       "9d06c86ecba40e1c695e69b55a40843df6a79cef",
						RC:       'R',
						Score:    100,
						Path:     "chezmoi_rename.go",
						OrigPath: "chezmoi.go",
					},
				},
			},
		},
		{
			name:      "modified_2",
			outputStr: "1 .M N... 100644 100644 100644 5716ca5987cbf97d6bb54920bea6adde242d87e6 5716ca5987cbf97d6bb54920bea6adde242d87e6 foo\n",
			expectedStatus: &Status{
				Ordinary: []OrdinaryStatus{
					{
						X:    '.',
						Y:    'M',
						Sub:  "N...",
						MH:   0100644,
						MI:   0100644,
						MW:   0100644,
						HH:   "5716ca5987cbf97d6bb54920bea6adde242d87e6",
						HI:   "5716ca5987cbf97d6bb54920bea6adde242d87e6",
						Path: "foo",
					},
				},
			},
		},
		{
			name:      "untracked",
			outputStr: "? chezmoi.go\n",
			expectedStatus: &Status{
				Untracked: []UntrackedStatus{
					{
						Path: "chezmoi.go",
					},
				},
			},
		},
		{
			name:      "ignored",
			outputStr: "! chezmoi.go\n",
			expectedStatus: &Status{
				Ignored: []IgnoredStatus{
					{
						Path: "chezmoi.go",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualStatus, err := ParseStatusPorcelainV2([]byte(tc.outputStr))
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, actualStatus)
		})
	}
}
