package chezmoi

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockPersistentState(t *testing.T) {
	var (
		bucket = []byte("bucket")
		key    = []byte("key")
		value  = []byte("value")
	)

	s1 := NewMockPersistentState()

	require.NoError(t, s1.Delete(bucket, value))

	actualValue, err := s1.Get(bucket, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)

	require.NoError(t, s1.Set(bucket, key, value))

	actualValue, err = s1.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s1.ForEach(bucket, func(k, v []byte) error {
		assert.Equal(t, key, k)
		assert.Equal(t, value, v)
		return nil
	}))

	assert.Equal(t, io.EOF, s1.ForEach(bucket, func(k, v []byte) error {
		return io.EOF
	}))

	s2 := NewMockPersistentState()
	require.NoError(t, s1.CopyTo(s2))
	actualValue, err = s2.Get(bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s1.Close())

	actualValue, err = s1.Get(bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s1.Delete(bucket, key))
	actualValue, err = s1.Get(bucket, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)
}
