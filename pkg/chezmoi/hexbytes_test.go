package chezmoi

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestHexBytes(t *testing.T) {
	for i, tc := range []struct {
		b           HexBytes
		expectedStr string
	}{
		{
			b:           nil,
			expectedStr: "\"\"\n",
		},
		{
			b:           []byte{0},
			expectedStr: "\"00\"\n",
		},
		{
			b:           []byte{0, 1, 2, 3},
			expectedStr: "\"00010203\"\n",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for _, format := range []Format{
				formatJSON{},
				formatYAML{},
			} {
				t.Run(format.Name(), func(t *testing.T) {
					actual, err := format.Marshal(tc.b)
					require.NoError(t, err)
					assert.Equal(t, []byte(tc.expectedStr), actual)
					var actualHexBytes HexBytes
					require.NoError(t, format.Unmarshal(actual, &actualHexBytes))
					assert.Equal(t, tc.b, actualHexBytes)
				})
			}
		})
	}
}
