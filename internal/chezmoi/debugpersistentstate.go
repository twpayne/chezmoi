package chezmoi

import (
	"log/slog"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

// A DebugPersistentState logs calls to a PersistentState.
type DebugPersistentState struct {
	logger          *slog.Logger
	persistentState PersistentState
}

// NewDebugPersistentState returns a new debugPersistentState that logs methods
// on persistentState to logger.
func NewDebugPersistentState(persistentState PersistentState, logger *slog.Logger) *DebugPersistentState {
	return &DebugPersistentState{
		logger:          logger,
		persistentState: persistentState,
	}
}

// Close implements PersistentState.Close.
func (s *DebugPersistentState) Close() error {
	err := s.persistentState.Close()
	chezmoilog.InfoOrError(s.logger, "Close", err)
	return err
}

// CopyTo implements PersistentState.CopyTo.
func (s *DebugPersistentState) CopyTo(p PersistentState) error {
	err := s.persistentState.CopyTo(p)
	chezmoilog.InfoOrError(s.logger, "CopyTo", err)
	return err
}

// Data implements PersistentState.Data.
func (s *DebugPersistentState) Data() (map[string]map[string]string, error) {
	data, err := s.persistentState.Data()
	chezmoilog.InfoOrError(s.logger, "Data", err,
		slog.Any("data", data),
	)
	return data, err
}

// Delete implements PersistentState.Delete.
func (s *DebugPersistentState) Delete(bucket, key []byte) error {
	err := s.persistentState.Delete(bucket, key)
	chezmoilog.InfoOrError(s.logger, "Delete", err,
		chezmoilog.Bytes("bucket", bucket),
		chezmoilog.Bytes("key", key),
	)
	return err
}

// DeleteBucket implements PersistentState.DeleteBucket.
func (s *DebugPersistentState) DeleteBucket(bucket []byte) error {
	err := s.persistentState.DeleteBucket(bucket)
	chezmoilog.InfoOrError(s.logger, "DeleteBucket", err,
		chezmoilog.Bytes("bucket", bucket),
	)
	return err
}

// ForEach implements PersistentState.ForEach.
func (s *DebugPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	err := s.persistentState.ForEach(bucket, func(k, v []byte) error {
		err := fn(k, v)
		chezmoilog.InfoOrError(s.logger, "ForEach", err,
			chezmoilog.Bytes("bucket", bucket),
			chezmoilog.Bytes("key", k),
			chezmoilog.Bytes("value", v),
		)
		return err
	})
	chezmoilog.InfoOrError(s.logger, "ForEach", err,
		chezmoilog.Bytes("bucket", bucket),
	)
	return err
}

// Get implements PersistentState.Get.
func (s *DebugPersistentState) Get(bucket, key []byte) ([]byte, error) {
	value, err := s.persistentState.Get(bucket, key)
	chezmoilog.InfoOrError(s.logger, "Get", err,
		chezmoilog.Bytes("bucket", bucket),
		chezmoilog.Bytes("key", key),
		chezmoilog.Bytes("value", value),
	)
	return value, err
}

// Set implements PersistentState.Set.
func (s *DebugPersistentState) Set(bucket, key, value []byte) error {
	err := s.persistentState.Set(bucket, key, value)
	chezmoilog.InfoOrError(s.logger, "Set", err,
		chezmoilog.Bytes("bucket", bucket),
		chezmoilog.Bytes("key", key),
		chezmoilog.Bytes("value", value),
	)
	return err
}
