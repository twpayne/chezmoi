package chezmoi

import (
	"os"
	"time"

	"go.etcd.io/bbolt"
)

// A BoltPersistentStateMode is a mode for opening a PersistentState.
type BoltPersistentStateMode int

// PersistentStateModes.
const (
	BoltPersistentStateReadOnly BoltPersistentStateMode = iota
	BoltPersistentStateReadWrite
)

// A BoltPersistentState is a state persisted with bolt.
type BoltPersistentState struct {
	db *bbolt.DB
}

// NewBoltPersistentState returns a new BoltPersistentState.
func NewBoltPersistentState(system System, path AbsPath, mode BoltPersistentStateMode) (*BoltPersistentState, error) {
	if _, err := system.Stat(path); os.IsNotExist(err) {
		if mode == BoltPersistentStateReadOnly {
			return &BoltPersistentState{}, nil
		}
		if err := MkdirAll(system, path.Dir(), 0o777); err != nil {
			return nil, err
		}
	}
	options := bbolt.Options{
		OpenFile: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			rawPath, err := system.RawPath(AbsPath(name))
			if err != nil {
				return nil, err
			}
			return os.OpenFile(string(rawPath), flag, perm)
		},
		ReadOnly: mode == BoltPersistentStateReadOnly,
		Timeout:  time.Second,
	}
	db, err := bbolt.Open(string(path), 0o600, &options)
	if err != nil {
		return nil, err
	}
	return &BoltPersistentState{
		db: db,
	}, nil
}

// Close closes b.
func (b *BoltPersistentState) Close() error {
	if b.db == nil {
		return nil
	}
	return b.db.Close()
}

// CopyTo copies b to p.
func (b *BoltPersistentState) CopyTo(p PersistentState) error {
	if b.db == nil {
		return nil
	}

	return b.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucket []byte, b *bbolt.Bucket) error {
			return b.ForEach(func(key, value []byte) error {
				return p.Set(copyByteSlice(bucket), copyByteSlice(key), copyByteSlice(value))
			})
		})
	})
}

// Delete deletes the value associate with key in bucket. If bucket or key does
// not exist then Delete does nothing.
func (b *BoltPersistentState) Delete(bucket, key []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.Delete(key)
	})
}

// ForEach calls fn for each key, value pair in bucket.
func (b *BoltPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	if b.db == nil {
		return nil
	}

	return b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			return fn(copyByteSlice(k), copyByteSlice(v))
		})
	})
}

// Get returns the value associated with key in bucket.
func (b *BoltPersistentState) Get(bucket, key []byte) ([]byte, error) {
	if b.db == nil {
		return nil, nil
	}

	var value []byte
	if err := b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		value = copyByteSlice(b.Get(key))
		return nil
	}); err != nil {
		return nil, err
	}
	return value, nil
}

// Set sets the value associated with key in bucket. bucket will be created if
// it does not already exist.
func (b *BoltPersistentState) Set(bucket, key, value []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func copyByteSlice(value []byte) []byte {
	if value == nil {
		return nil
	}
	result := make([]byte, len(value))
	copy(result, value)
	return result
}
