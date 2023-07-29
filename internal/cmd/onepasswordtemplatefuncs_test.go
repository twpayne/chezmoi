package cmd

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestOnepasswordAccountMap(t *testing.T) {
	for _, tc := range []struct {
		name     string
		accounts []onepasswordAccount
		expected map[string]string
	}{
		{
			name: "single_account_without_shorthand",
			accounts: []onepasswordAccount{
				{
					URL:         "account1.1password.ca",
					Email:       "my@email.com",
					UserUUID:    "some-user-uuid",
					AccountUUID: "some-account-uuid",
				},
			},
			expected: map[string]string{
				"account1.1password.ca":    "some-account-uuid",
				"account1":                 "some-account-uuid",
				"my@account1.1password.ca": "some-account-uuid",
				"my@account1":              "some-account-uuid",
				"my@email.com":             "some-account-uuid",
				"my":                       "some-account-uuid",
				"some-account-uuid":        "some-account-uuid",
				"some-user-uuid":           "some-account-uuid",
			},
		},
		{
			name: "single_account_with_shorthand",
			accounts: []onepasswordAccount{
				{
					URL:         "account1.1password.ca",
					Email:       "my@email.com",
					UserUUID:    "some-user-uuid",
					AccountUUID: "some-account-uuid",
					Shorthand:   "some-account-shorthand",
				},
			},
			expected: map[string]string{
				"account1.1password.ca":    "some-account-uuid",
				"account1":                 "some-account-uuid",
				"my@account1.1password.ca": "some-account-uuid",
				"my@account1":              "some-account-uuid",
				"my@email.com":             "some-account-uuid",
				"my":                       "some-account-uuid",
				"some-account-shorthand":   "some-account-uuid",
				"some-account-uuid":        "some-account-uuid",
				"some-user-uuid":           "some-account-uuid",
			},
		},
		{
			name: "multiple_unambiguous_accounts",
			accounts: []onepasswordAccount{
				{
					URL:         "account1.1password.ca",
					Email:       "my@email.com",
					UserUUID:    "some-user-uuid",
					AccountUUID: "some-account-uuid",
					Shorthand:   "some-account-shorthand",
				},
				{
					URL:         "account2.1password.ca",
					Email:       "me@otheremail.org",
					UserUUID:    "some-other-user-uuid",
					AccountUUID: "some-other-account-uuid",
					Shorthand:   "some-other-account-shorthand",
				},
			},
			expected: map[string]string{
				"account1.1password.ca":        "some-account-uuid",
				"account1":                     "some-account-uuid",
				"account2.1password.ca":        "some-other-account-uuid",
				"account2":                     "some-other-account-uuid",
				"me@account2.1password.ca":     "some-other-account-uuid",
				"me@account2":                  "some-other-account-uuid",
				"me@otheremail.org":            "some-other-account-uuid",
				"me":                           "some-other-account-uuid",
				"my@account1.1password.ca":     "some-account-uuid",
				"my@account1":                  "some-account-uuid",
				"my@email.com":                 "some-account-uuid",
				"my":                           "some-account-uuid",
				"some-account-shorthand":       "some-account-uuid",
				"some-account-uuid":            "some-account-uuid",
				"some-other-account-shorthand": "some-other-account-uuid",
				"some-other-account-uuid":      "some-other-account-uuid",
				"some-other-user-uuid":         "some-other-account-uuid",
				"some-user-uuid":               "some-account-uuid",
			},
		},
		{
			name: "multiple_ambiguous_accounts",
			accounts: []onepasswordAccount{
				{
					URL:         "account1.1password.ca",
					Email:       "my@email.com",
					UserUUID:    "some-user-uuid",
					AccountUUID: "some-account-uuid",
					Shorthand:   "some-account-shorthand",
				},
				{
					URL:         "account1.1password.ca",
					Email:       "your@email.com",
					UserUUID:    "some-other-user-uuid",
					AccountUUID: "some-other-account-uuid",
					Shorthand:   "some-other-account-shorthand",
				},
			},
			expected: map[string]string{
				"my@account1.1password.ca":     "some-account-uuid",
				"my@account1":                  "some-account-uuid",
				"my@email.com":                 "some-account-uuid",
				"my":                           "some-account-uuid",
				"some-account-shorthand":       "some-account-uuid",
				"some-account-uuid":            "some-account-uuid",
				"some-other-account-shorthand": "some-other-account-uuid",
				"some-other-account-uuid":      "some-other-account-uuid",
				"some-other-user-uuid":         "some-other-account-uuid",
				"some-user-uuid":               "some-account-uuid",
				"your@account1.1password.ca":   "some-other-account-uuid",
				"your@account1":                "some-other-account-uuid",
				"your@email.com":               "some-other-account-uuid",
				"your":                         "some-other-account-uuid",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := onepasswordAccountMap(tc.accounts)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
