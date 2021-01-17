package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileMode(t *testing.T) {
	for _, tc := range []struct {
		s              string
		expectedErr    bool
		expected       fileMode
		expectedString string
	}{
		{
			s:              "0",
			expected:       0,
			expectedString: "000",
		},
		{
			s:              "644",
			expected:       0o644,
			expectedString: "644",
		},
		{
			s:              "755",
			expected:       0o755,
			expectedString: "755",
		},
		{
			s:              "0",
			expected:       0,
			expectedString: "000",
		},
		{
			s:           "-0",
			expectedErr: true,
		},
		{
			s:           "s",
			expectedErr: true,
		},
		{
			s:           "008",
			expectedErr: true,
		},
		{
			s:           "01000",
			expectedErr: true,
		},
		{
			s:           "-0",
			expectedErr: true,
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			var p fileMode
			err := p.Set(tc.s)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedString, p.String())
				assert.Equal(t, "file mode", p.Type())
			}
		})
	}
}
