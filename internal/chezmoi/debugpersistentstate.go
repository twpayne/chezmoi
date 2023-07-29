package chezmoi

import (
	"github.com/rs/zerolog"
)

// A DebugPersistentState logs calls to a PersistentState.
type DebugPersistentState struct {
	logger          *zerolog.Logger
	persistentState PersistentState
}

// NewDebugPersistentState returns a new debugPersistentState that logs methods
// on persistentState to logger.
func NewDebugPersistentState(
	persistentState PersistentState,
	logger *zerolog.Logger,
) *DebugPersistentState {
	return &DebugPersistentState{
		logger:          logger,
		persistentState: persistentState,
	}
}

// Close implements PersistentState.Close.
func (s *DebugPersistentState) Close() error {
	err := s.persistentState.Close()
	s.logger.Err(err).
		Msg("Close")
	return err
}

// CopyTo implements PersistentState.CopyTo.
func (s *DebugPersistentState) CopyTo(p PersistentState) error {
	err := s.persistentState.CopyTo(p)
	s.logger.Err(err).
		Msg("CopyTo")
	return err
}

// Data implements PersistentState.Data.
func (s *DebugPersistentState) Data() (any, error) {
	data, err := s.persistentState.Data()
	s.logger.Err(err).
		Interface("data", data).
		Msg("Data")
	return data, err
}

// Delete implements PersistentState.Delete.
func (s *DebugPersistentState) Delete(bucket, key []byte) error {
	err := s.persistentState.Delete(bucket, key)
	s.logger.Err(err).
		Bytes("bucket", bucket).
		Bytes("key", key).
		Msg("Delete")
	return err
}

// DeleteBucket implements PersistentState.DeleteBucket.
func (s *DebugPersistentState) DeleteBucket(bucket []byte) error {
	err := s.persistentState.DeleteBucket(bucket)
	s.logger.Err(err).
		Bytes("bucket", bucket).
		Msg("DeleteBucket")
	return err
}

// ForEach implements PersistentState.ForEach.
func (s *DebugPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	err := s.persistentState.ForEach(bucket, func(k, v []byte) error {
		err := fn(k, v)
		s.logger.Err(err).
			Bytes("bucket", bucket).
			Bytes("key", k).
			Bytes("value", v).
			Msg("ForEach")
		return err
	})
	s.logger.Err(err).
		Bytes("bucket", bucket).
		Msg("ForEach")
	return err
}

// Get implements PersistentState.Get.
func (s *DebugPersistentState) Get(bucket, key []byte) ([]byte, error) {
	value, err := s.persistentState.Get(bucket, key)
	s.logger.Err(err).
		Bytes("bucket", bucket).
		Bytes("key", key).
		Bytes("value", value).
		Msg("Get")
	return value, err
}

// Set implements PersistentState.Set.
func (s *DebugPersistentState) Set(bucket, key, value []byte) error {
	err := s.persistentState.Set(bucket, key, value)
	s.logger.Err(err).
		Bytes("bucket", bucket).
		Bytes("key", key).
		Bytes("value", value).
		Msg("Set")
	return err
}
