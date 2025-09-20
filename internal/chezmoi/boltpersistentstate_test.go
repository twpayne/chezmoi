package chezmoi

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"chezmoi.io/chezmoi/internal/chezmoitest"
)

var _ PersistentState = &BoltPersistentState{}

func TestBoltPersistentState(t *testing.T) {
	chezmoitest.WithTestFS(t, nil, func(fileSystem vfs.FS) {
		var (
			system = NewRealSystem(fileSystem)
			path   = NewAbsPath("/home/user/.config/chezmoi/chezmoistate.boltdb")
			bucket = []byte("bucket")
			key    = []byte("key")
			value  = []byte("value")
		)

		b1, err := NewBoltPersistentState(system, path, BoltPersistentStateReadWrite)
		assert.NoError(t, err)

		// Test that getting a key from an non-existent state does not create
		// the state.
		actualValue, err := b1.Get(bucket, key)
		assert.NoError(t, err)
		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath(path.String(),
				vfst.TestDoesNotExist(),
			),
		)
		assert.Equal(t, []byte(nil), actualValue)

		// Test that deleting a key from a non-existent state does not create
		// the state.
		assert.NoError(t, b1.Delete(bucket, key))
		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath(path.String(),
				vfst.TestDoesNotExist(),
			),
		)

		// Test that setting a key creates the state.
		assert.NoError(t, b1.Set(bucket, key, value))
		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath(path.String(),
				vfst.TestModeIsRegular(),
			),
		)
		actualValue, err = b1.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value, actualValue)

		visited := false
		assert.NoError(t, b1.ForEach(bucket, func(k, v []byte) error {
			visited = true
			assert.Equal(t, key, k)
			assert.Equal(t, value, v)
			return nil
		}))
		assert.True(t, visited)

		data, err := b1.Data()
		assert.NoError(t, err)
		assert.Equal(t, map[string]map[string]string{
			string(bucket): {
				string(key): string(value),
			},
		}, data)

		assert.NoError(t, b1.Close())

		b2, err := NewBoltPersistentState(system, path, BoltPersistentStateReadWrite)
		assert.NoError(t, err)

		assert.NoError(t, b2.Delete(bucket, key))

		actualValue, err = b2.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, []byte(nil), actualValue)
	})
}

func TestBoltPersistentStateMock(t *testing.T) {
	chezmoitest.WithTestFS(t, nil, func(fileSystem vfs.FS) {
		var (
			system = NewRealSystem(fileSystem)
			path   = NewAbsPath("/home/user/.config/chezmoi/chezmoistate.boltdb")
			bucket = []byte("bucket")
			key    = []byte("key")
			value1 = []byte("value1")
			value2 = []byte("value2")
		)

		b, err := NewBoltPersistentState(system, path, BoltPersistentStateReadWrite)
		assert.NoError(t, err)
		assert.NoError(t, b.Set(bucket, key, value1))

		m := NewMockPersistentState()
		assert.NoError(t, b.CopyTo(m), err)

		actualValue, err := m.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value1, actualValue)

		assert.NoError(t, m.Set(bucket, key, value2))
		actualValue, err = m.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value2, actualValue)
		actualValue, err = b.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value1, actualValue)

		assert.NoError(t, m.Delete(bucket, key))
		actualValue, err = m.Get(bucket, key)
		assert.NoError(t, err)
		assert.Zero(t, actualValue)
		actualValue, err = b.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value1, actualValue)

		assert.NoError(t, b.Close())
	})
}

func TestBoltPersistentStateGeneric(t *testing.T) {
	system := NewRealSystem(vfs.OSFS)
	var tempDirs []string
	defer func() {
		for _, tempDir := range tempDirs {
			assert.NoError(t, os.RemoveAll(tempDir))
		}
	}()
	testPersistentState(t, func() PersistentState {
		tempDir := t.TempDir()
		absPath := NewAbsPath(tempDir).JoinString("chezmoistate.boltdb")
		b, err := NewBoltPersistentState(system, absPath, BoltPersistentStateReadWrite)
		assert.NoError(t, err)
		return b
	})
}

func TestBoltPersistentStateReadOnly(t *testing.T) {
	chezmoitest.WithTestFS(t, nil, func(fileSystem vfs.FS) {
		var (
			system = NewRealSystem(fileSystem)
			path   = NewAbsPath("/home/user/.config/chezmoi/chezmoistate.boltdb")
			bucket = []byte("bucket")
			key    = []byte("key")
			value  = []byte("value")
		)

		b1, err := NewBoltPersistentState(system, path, BoltPersistentStateReadWrite)
		assert.NoError(t, err)
		assert.NoError(t, b1.Set(bucket, key, value))
		assert.NoError(t, b1.Close())

		b2, err := NewBoltPersistentState(system, path, BoltPersistentStateReadOnly)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, b2.Close())
		}()

		b3, err := NewBoltPersistentState(system, path, BoltPersistentStateReadOnly)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, b3.Close())
		}()

		actualValueB, err := b2.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value, actualValueB)

		actualValueC, err := b3.Get(bucket, key)
		assert.NoError(t, err)
		assert.Equal(t, value, actualValueC)

		assert.Error(t, b2.Set(bucket, key, value))
		assert.Error(t, b3.Set(bucket, key, value))

		assert.NoError(t, b2.Close())
		assert.NoError(t, b3.Close())
	})
}
