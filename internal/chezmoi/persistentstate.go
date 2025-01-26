package chezmoi

var (
	// ConfigStateBucket is the bucket for recording the config state.
	ConfigStateBucket = []byte("configState")

	// EntryStateBucket is the bucket for recording the entry states.
	EntryStateBucket = []byte("entryState")

	// GitRepoExternalStateBucket is the bucket for recording the state of commands
	// that modify directories.
	GitRepoExternalStateBucket = []byte("gitRepoExternalState")

	// ScriptStateBucket is the bucket for recording the state of run once
	// scripts.
	ScriptStateBucket = []byte("scriptState")

	stateFormat = formatJSON{}
)

// A PersistentState is a persistent state.
type PersistentState interface {
	Close() error
	CopyTo(s PersistentState) error
	Data() (map[string]map[string]string, error)
	Delete(bucket, key []byte) error
	DeleteBucket(bucket []byte) error
	ForEach(bucket []byte, fn func(k, v []byte) error) error
	Get(bucket, key []byte) ([]byte, error)
	Set(bucket, key, value []byte) error
}

// PersistentStateBucketData returns the state data in bucket in s.
func PersistentStateBucketData(s PersistentState, bucket []byte) (map[string]any, error) {
	result := make(map[string]any)
	if err := s.ForEach(bucket, func(k, v []byte) error {
		var value map[string]any
		if err := stateFormat.Unmarshal(v, &value); err != nil {
			return err
		}
		result[string(k)] = value
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// PersistentStateData returns the structured data in s.
func PersistentStateData(s PersistentState, buckets map[string][]byte) (map[string]any, error) {
	result := make(map[string]any)
	for bucketName, bucketKey := range buckets {
		stateData, err := PersistentStateBucketData(s, bucketKey)
		if err != nil {
			return nil, err
		}
		result[bucketName] = stateData
	}
	return result, nil
}

// PersistentStateGet gets the value associated with key in bucket in s, if it exists.
func PersistentStateGet(s PersistentState, bucket, key []byte, value any) (bool, error) {
	data, err := s.Get(bucket, key)
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil
	}
	if err := stateFormat.Unmarshal(data, value); err != nil {
		return false, err
	}
	return true, nil
}

// PersistentStateSet sets the value associated with key in bucket in s.
func PersistentStateSet(s PersistentState, bucket, key []byte, value any) error {
	data, err := stateFormat.Marshal(value)
	if err != nil {
		return err
	}
	return s.Set(bucket, key, data)
}
