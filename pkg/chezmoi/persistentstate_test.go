package chezmoi

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPersistentState(t *testing.T, constructor func() PersistentState) {
	t.Helper()

	var (
		bucket1 = []byte("bucket1")
		bucket2 = []byte("bucket2")
		key     = []byte("key1")
		value   = []byte("value")
	)

	s1 := constructor()

	require.NoError(t, s1.Delete(bucket1, value))

	actualValue, err := s1.Get(bucket1, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)

	require.NoError(t, s1.Set(bucket1, key, value))

	actualValue, err = s1.Get(bucket1, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s1.ForEach(bucket1, func(k, v []byte) error {
		assert.Equal(t, key, k)
		assert.Equal(t, value, v)
		return nil
	}))

	assert.Equal(t, io.EOF, s1.ForEach(bucket1, func(k, v []byte) error {
		return io.EOF
	}))

	s2 := constructor()
	require.NoError(t, s1.CopyTo(s2))
	actualValue, err = s2.Get(bucket1, key)
	assert.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s2.Close())

	actualValue, err = s1.Get(bucket1, key)
	assert.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s1.Delete(bucket1, key))
	actualValue, err = s1.Get(bucket1, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)

	require.NoError(t, s1.Set(bucket2, key, value))
	actualValue, err = s1.Get(bucket2, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValue)
	require.NoError(t, s1.DeleteBucket(bucket2))
	actualValue, err = s1.Get(bucket2, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)
}
