package chezmoigit

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestParseStatusPorcelainV2(t *testing.T) {
	for _, tc := range []struct {
		name           string
		outputStr      string
		expectedEmpty  bool
		expectedStatus *Status
	}{
		{
			name:           "empty",
			outputStr:      "",
			expectedStatus: &Status{},
		},
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
						MI:   0o100644,
						MW:   0o100644,
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
						MH:   0o100644,
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
			name: "copied",
			outputStr: chezmoitest.JoinLines(
				"2 C. N... 100644 100644 100644 4a58007052a65fbc2fc3f910f2855f45a4058e74 4a58007052a65fbc2fc3f910f2855f45a4058e74 C100 c\tb",
				"2 R. N... 100644 100644 100644 4a58007052a65fbc2fc3f910f2855f45a4058e74 4a58007052a65fbc2fc3f910f2855f45a4058e74 R100 d\tb",
			),
			expectedStatus: &Status{
				RenamedOrCopied: []RenamedOrCopiedStatus{
					{
						X:        'C',
						Y:        '.',
						Sub:      "N...",
						MH:       0o100644,
						MI:       0o100644,
						MW:       0o100644,
						HH:       "4a58007052a65fbc2fc3f910f2855f45a4058e74",
						HI:       "4a58007052a65fbc2fc3f910f2855f45a4058e74",
						RC:       'C',
						Score:    100,
						Path:     "c",
						OrigPath: "b",
					},
					{
						X:        'R',
						Y:        '.',
						Sub:      "N...",
						MH:       0o100644,
						MI:       0o100644,
						MW:       0o100644,
						HH:       "4a58007052a65fbc2fc3f910f2855f45a4058e74",
						HI:       "4a58007052a65fbc2fc3f910f2855f45a4058e74",
						RC:       'R',
						Score:    100,
						Path:     "d",
						OrigPath: "b",
					},
				},
			},
		},
		{
			name:      "update",
			outputStr: "1 .M N... 100644 100644 100644 353dbbb3c29a80fb44d4e26dac111739d25294db 353dbbb3c29a80fb44d4e26dac111739d25294db cmd/git.go\n",
			expectedStatus: &Status{
				Ordinary: []OrdinaryStatus{
					{
						X:    '.',
						Y:    'M',
						Sub:  "N...",
						MH:   0o100644,
						MI:   0o100644,
						MW:   0o100644,
						HH:   "353dbbb3c29a80fb44d4e26dac111739d25294db",
						HI:   "353dbbb3c29a80fb44d4e26dac111739d25294db",
						Path: "cmd/git.go",
					},
				},
			},
		},
		{
			name:      "renamed",
			outputStr: "2 R. N... 100644 100644 100644 9d06c86ecba40e1c695e69b55a40843df6a79cef 9d06c86ecba40e1c695e69b55a40843df6a79cef R100 chezmoi_rename.go\tchezmoi.go\n",
			expectedStatus: &Status{
				RenamedOrCopied: []RenamedOrCopiedStatus{
					{
						X:        'R',
						Y:        '.',
						Sub:      "N...",
						MH:       0o100644,
						MI:       0o100644,
						MW:       0o100644,
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
			name:      "renamed_2",
			outputStr: "2 R. N... 100644 100644 100644 ddbd961d7e4db2bb6615a9e8ce86364fa65e732d ddbd961d7e4db2bb6615a9e8ce86364fa65e732d R100 dot_config/chezmoi/private_chezmoi.toml\tdot_config/chezmoi/chezmoi.toml", //nolint:dupword
			expectedStatus: &Status{
				RenamedOrCopied: []RenamedOrCopiedStatus{
					{
						X:        82,
						Y:        46,
						Sub:      "N...",
						MH:       0o100644,
						MI:       0o100644,
						MW:       0o100644,
						HH:       "ddbd961d7e4db2bb6615a9e8ce86364fa65e732d",
						HI:       "ddbd961d7e4db2bb6615a9e8ce86364fa65e732d",
						RC:       'R',
						Score:    100,
						Path:     "dot_config/chezmoi/private_chezmoi.toml",
						OrigPath: "dot_config/chezmoi/chezmoi.toml",
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
						MH:   0o100644,
						MI:   0o100644,
						MW:   0o100644,
						HH:   "5716ca5987cbf97d6bb54920bea6adde242d87e6",
						HI:   "5716ca5987cbf97d6bb54920bea6adde242d87e6",
						Path: "foo",
					},
				},
			},
		},
		{
			name:      "unmerged",
			outputStr: "u UU N... 100644 100644 100644 100644 78981922613b2afb6025042ff6bd878ac1994e85 0f7bc766052a5a0ee28a393d51d2370f96d8ceb8 422c2b7ab3b3c668038da977e4e93a5fc623169c README.md\n",
			expectedStatus: &Status{
				Unmerged: []UnmergedStatus{
					{
						X:    'U',
						Y:    'U',
						Sub:  "N...",
						M1:   0o100644,
						M2:   0o100644,
						M3:   0o100644,
						MW:   0o100644,
						H1:   "78981922613b2afb6025042ff6bd878ac1994e85",
						H2:   "0f7bc766052a5a0ee28a393d51d2370f96d8ceb8",
						H3:   "422c2b7ab3b3c668038da977e4e93a5fc623169c",
						Path: "README.md",
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
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, actualStatus)
		})
	}
}
