package chezmoi

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"syscall"
	"time"

	"go.etcd.io/bbolt"
)

// A BoltPersistentStateMode is a mode for opening a PersistentState.
type BoltPersistentStateMode int

// Persistent state modes.
const (
	BoltPersistentStateReadOnly BoltPersistentStateMode = iota
	BoltPersistentStateReadWrite
)

// A BoltPersistentState is a state persisted with bolt.
type BoltPersistentState struct {
	system  System
	empty   bool
	path    AbsPath
	options bbolt.Options
	db      *bbolt.DB
}

// NewBoltPersistentState returns a new BoltPersistentState.
func NewBoltPersistentState(
	system System,
	path AbsPath,
	mode BoltPersistentStateMode,
) (*BoltPersistentState, error) {
	empty := false
	switch _, err := system.Stat(path); {
	case errors.Is(err, fs.ErrNotExist):
		// We need to simulate an empty persistent state because Bolt's
		// read-only mode is only supported for databases that already exist.
		//
		// If the database does not already exist, then Bolt will open it with
		// O_RDONLY and then attempt to initialize it, which leads to EBADF
		// errors on Linux. See also https://github.com/etcd-io/bbolt/issues/98.
		empty = true
	case err != nil:
		return nil, err
	}

	options := bbolt.Options{
		OpenFile: func(name string, flag int, perm fs.FileMode) (*os.File, error) {
			rawPath, err := system.RawPath(NewAbsPath(name))
			if err != nil {
				return nil, err
			}
			return os.OpenFile(rawPath.String(), flag, perm)
		},
		ReadOnly: mode == BoltPersistentStateReadOnly,
		Timeout:  time.Second,
	}

	return &BoltPersistentState{
		system:  system,
		empty:   empty,
		path:    path,
		options: options,
	}, nil
}

// Close closes b.
func (b *BoltPersistentState) Close() error {
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			return err
		}
		b.db = nil
	}
	return nil
}

// CopyTo copies b to p.
func (b *BoltPersistentState) CopyTo(p PersistentState) error {
	if b.empty {
		return nil
	}
	if err := b.open(); err != nil {
		return err
	}

	return b.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucket []byte, b *bbolt.Bucket) error {
			return b.ForEach(func(key, value []byte) error {
				return p.Set(slices.Clone(bucket), slices.Clone(key), slices.Clone(value))
			})
		})
	})
}

// Delete deletes the value associate with key in bucket. If bucket or key does
// not exist then Delete does nothing.
func (b *BoltPersistentState) Delete(bucket, key []byte) error {
	if b.empty {
		return nil
	}
	if err := b.open(); err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.Delete(key)
	})
}

// DeleteBucket deletes the bucket.
func (b *BoltPersistentState) DeleteBucket(bucket []byte) error {
	if b.empty {
		return nil
	}
	if err := b.open(); err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(bucket)
	})
}

// Data returns all the data in b.
func (b *BoltPersistentState) Data() (any, error) {
	if b.empty {
		return nil, nil
	}
	if err := b.open(); err != nil {
		return nil, err
	}

	data := make(map[string]map[string]string)
	err := b.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			bucketName := string(name)
			bucket, ok := data[bucketName]
			if !ok {
				bucket = make(map[string]string)
				data[bucketName] = bucket
			}
			return b.ForEach(func(k, v []byte) error {
				bucket[string(k)] = string(v)
				return nil
			})
		})
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ForEach calls fn for each key, value pair in bucket.
func (b *BoltPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	if b.empty {
		return nil
	}
	if err := b.open(); err != nil {
		return err
	}

	return b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			return fn(slices.Clone(k), slices.Clone(v))
		})
	})
}

// Get returns the value associated with key in bucket.
func (b *BoltPersistentState) Get(bucket, key []byte) ([]byte, error) {
	if b.empty {
		return nil, nil
	}
	if err := b.open(); err != nil {
		return nil, err
	}

	var value []byte
	if err := b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		value = slices.Clone(b.Get(key))
		return nil
	}); err != nil {
		return nil, err
	}
	return value, nil
}

// Set sets the value associated with key in bucket. bucket will be created if
// it does not already exist.
func (b *BoltPersistentState) Set(bucket, key, value []byte) error {
	if err := b.open(); err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

// open opens b's database if it is not already open, creating it if needed.
func (b *BoltPersistentState) open() error {
	if b.db != nil {
		return nil
	}
	if err := MkdirAll(b.system, b.path.Dir(), fs.ModePerm); err != nil {
		return err
	}
	switch db, err := bbolt.Open(b.path.String(), 0o600, &b.options); {
	case errors.Is(err, syscall.EINVAL):
		// Assume that any EINVAL error is because flock(2) failed.
		return fmt.Errorf("open %s: failed to acquire lock: %w", b.path, err)
	case err != nil:
		return fmt.Errorf("open %s: %w", b.path, err)
	default:
		b.empty = false
		b.db = db
		return nil
	}
}
