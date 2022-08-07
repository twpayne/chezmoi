package chezmoi

// A NullPersistentState is an empty PersistentState that returns the zero value
// for all reads and silently consumes all writes.
type NullPersistentState struct{}

// Close does nothing.
func (NullPersistentState) Close() error { return nil }

// CopyTo does nothing.
func (NullPersistentState) CopyTo(s PersistentState) error { return nil }

// Data does nothing.
func (NullPersistentState) Data() (any, error) { return nil, nil }

// Delete does nothing.
func (NullPersistentState) Delete(bucket, key []byte) error { return nil }

// DeleteBucket does nothing.
func (NullPersistentState) DeleteBucket(bucket []byte) error { return nil }

// ForEach does nothing.
func (NullPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error { return nil }

// Get does nothing.
func (NullPersistentState) Get(bucket, key []byte) ([]byte, error) { return nil, nil }

// Set does nothing.
func (NullPersistentState) Set(bucket, key, value []byte) error { return nil }
