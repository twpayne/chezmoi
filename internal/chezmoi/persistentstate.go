package chezmoi

var (
	// ConfigStateBucket is the bucket for recording the config state.
	ConfigStateBucket = []byte("configState")

	// EntryStateBucket is the bucket for recording the entry states.
	EntryStateBucket = []byte("entryState")

	// scriptStateBucket is the bucket for recording the state of run once
	// scripts.
	scriptStateBucket = []byte("scriptState")

	stateFormat = jsonFormat{}
)

// A PersistentState is a persistent state.
type PersistentState interface {
	Close() error
	CopyTo(s PersistentState) error
	Data() (interface{}, error)
	Delete(bucket, key []byte) error
	ForEach(bucket []byte, fn func(k, v []byte) error) error
	Get(bucket, key []byte) ([]byte, error)
	Set(bucket, key, value []byte) error
}

// PersistentStateData returns the structured data in s.
func PersistentStateData(s PersistentState) (interface{}, error) {
	configStateData, err := persistentStateBucketData(s, ConfigStateBucket)
	if err != nil {
		return nil, err
	}
	entryStateData, err := persistentStateBucketData(s, EntryStateBucket)
	if err != nil {
		return nil, err
	}
	scriptStateData, err := persistentStateBucketData(s, scriptStateBucket)
	if err != nil {
		return nil, err
	}
	return struct {
		ConfigState interface{} `json:"configState" toml:"configState" yaml:"configState"`
		EntryState  interface{} `json:"entryState" toml:"entryState" yaml:"entryState"`
		ScriptState interface{} `json:"scriptState" toml:"scriptState" yaml:"scriptState"`
	}{
		ConfigState: configStateData,
		EntryState:  entryStateData,
		ScriptState: scriptStateData,
	}, nil
}

// persistentStateBucketData returns the state data in bucket in s.
func persistentStateBucketData(s PersistentState, bucket []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if err := s.ForEach(bucket, func(k, v []byte) error {
		var value map[string]interface{}
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

// persistentStateGet gets the value associated with key in bucket in s, if it exists.
func persistentStateGet(s PersistentState, bucket, key []byte, value interface{}) (bool, error) {
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

// persistentStateSet sets the value associated with key in bucket in s.
func persistentStateSet(s PersistentState, bucket, key []byte, value interface{}) error {
	data, err := stateFormat.Marshal(value)
	if err != nil {
		return err
	}
	return s.Set(bucket, key, data)
}
