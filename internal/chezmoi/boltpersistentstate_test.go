package chezmoi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
	bolt "go.etcd.io/bbolt"
)

var _ PersistentState = &BoltPersistentState{}

func TestBoltPersistentState(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.config/chezmoi": &vfst.Dir{Perm: 0o755},
	})
	require.NoError(t, err)
	defer cleanup()

	path := "/home/user/.config/chezmoi/chezmoistate.boltdb"
	b, err := NewBoltPersistentState(fs, path, nil)
	require.NoError(t, err)
	vfst.RunTests(t, fs, "",
		vfst.TestPath(path,
			vfst.TestDoesNotExist,
		),
	)

	var (
		bucket = []byte("bucket")
		key    = []byte("key")
		value  = []byte("value")
	)

	require.NoError(t, b.Delete(bucket, key))
	vfst.RunTests(t, fs, "",
		vfst.TestPath(path,
			vfst.TestDoesNotExist,
		),
	)

	actualValue, err := b.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, []byte(nil), actualValue)
	vfst.RunTests(t, fs, "",
		vfst.TestPath(path,
			vfst.TestDoesNotExist,
		),
	)

	assert.NoError(t, b.Set(bucket, key, value))
	vfst.RunTests(t, fs, "",
		vfst.TestPath(path,
			vfst.TestModeIsRegular,
		),
	)

	actualValue, err = b.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, b.Close())

	b, err = NewBoltPersistentState(fs, path, nil)
	require.NoError(t, err)

	require.NoError(t, b.Delete(bucket, key))

	actualValue, err = b.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, []byte(nil), actualValue)
}

func TestBoltPersistentStateReadOnly(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.config/chezmoi": &vfst.Dir{Perm: 0o755},
	})
	require.NoError(t, err)
	defer cleanup()

	path := "/home/user/.config/chezmoi/chezmoistate.boltdb"
	bucket := []byte("bucket")
	key := []byte("key")
	value := []byte("value")

	a, err := NewBoltPersistentState(fs, path, nil)
	require.NoError(t, err)
	require.NoError(t, a.Set(bucket, key, value))
	require.NoError(t, a.Close())

	b, err := NewBoltPersistentState(fs, path, &bolt.Options{
		ReadOnly: true,
		Timeout:  1 * time.Second,
	})
	require.NoError(t, err)
	defer b.Close()

	c, err := NewBoltPersistentState(fs, path, &bolt.Options{
		ReadOnly: true,
		Timeout:  1 * time.Second,
	})
	require.NoError(t, err)
	defer c.Close()

	actualValueB, err := b.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValueB)

	actualValueC, err := c.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValueC)

	assert.Error(t, b.Set(bucket, key, value))
	assert.Error(t, c.Set(bucket, key, value))

	require.NoError(t, b.Close())
	require.NoError(t, c.Close())
}
