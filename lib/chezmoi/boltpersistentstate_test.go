package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

var _ PersistentState = &BoltPersistentState{}

func TestBoltPersistentState(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.config/chezmoi": &vfst.Dir{Perm: 0755},
	})
	require.NoError(t, err)
	defer cleanup()

	path := "/home/user/.config/chezmoi/state.boltdb"
	b, err := NewBoltPersistentState(fs, path)
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

	b, err = NewBoltPersistentState(fs, path)
	require.NoError(t, err)

	require.NoError(t, b.Delete(bucket, key))

	actualValue, err = b.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, []byte(nil), actualValue)
}
