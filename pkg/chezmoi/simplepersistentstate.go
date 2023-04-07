package chezmoi

// FIXME use HexBytes

import (
	"errors"
	"io/fs"
)

type SimplePersistentState struct {
	system   System
	filename AbsPath
	modified bool
	state    map[string]map[string]string
}

func NewSimplePersistentState(system System, filename AbsPath) *SimplePersistentState {
	return &SimplePersistentState{
		system:   system,
		filename: filename,
	}
}

func (s *SimplePersistentState) Close() error {
	if !s.modified {
		return nil
	}
	data, err := stateFormat.Marshal(s.state)
	if err != nil {
		return err
	}
	if err := s.system.WriteFile(s.filename, data, 0o600); err != nil {
		return err
	}
	s.modified = false
	return nil
}

func (s *SimplePersistentState) CopyTo(other PersistentState) error {
	if err := s.open(); err != nil {
		return err
	}
	for bucketStr, bucketMap := range s.state {
		bucket := []byte(bucketStr)
		for keyStr, valueStr := range bucketMap {
			key := []byte(keyStr)
			value := []byte(valueStr)
			if err := other.Set(bucket, key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SimplePersistentState) Data() (any, error) {
	if err := s.open(); err != nil {
		return nil, err
	}
	return s.state, nil
}

func (s *SimplePersistentState) Delete(bucket, key []byte) error {
	if err := s.open(); err != nil {
		return err
	}
	bucketMap, ok := s.state[string(bucket)]
	if !ok {
		return nil
	}
	keyStr := string(key)
	if _, ok := bucketMap[keyStr]; !ok {
		return nil
	}
	delete(bucketMap, keyStr)
	s.modified = true
	return nil
}

func (s *SimplePersistentState) DeleteBucket(bucket []byte) error {
	if err := s.open(); err != nil {
		return err
	}
	bucketStr := string(bucket)
	if _, ok := s.state[bucketStr]; !ok {
		return nil
	}
	delete(s.state, bucketStr)
	s.modified = true
	return nil
}

func (s *SimplePersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	if err := s.open(); err != nil {
		return err
	}
	for keyStr, valueStr := range s.state[string(bucket)] {
		if err := fn([]byte(keyStr), []byte(valueStr)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SimplePersistentState) Get(bucket, key []byte) ([]byte, error) {
	if err := s.open(); err != nil {
		return nil, err
	}
	bucketMap, ok := s.state[string(bucket)]
	if !ok {
		return nil, nil
	}
	value, ok := bucketMap[string(key)]
	if !ok {
		return nil, nil
	}
	return []byte(value), nil
}

func (s *SimplePersistentState) Set(bucket, key, value []byte) error {
	if err := s.open(); err != nil {
		return err
	}
	bucketStr := string(bucket)
	bucketMap, ok := s.state[bucketStr]
	if !ok {
		bucketMap = make(map[string]string)
		s.state[bucketStr] = bucketMap
	}
	keyStr := string(key)
	valueStr := string(value)
	bucketMap[keyStr] = valueStr
	s.modified = true
	return nil
}

func (s *SimplePersistentState) open() error {
	if s.state != nil {
		return nil
	}
	switch data, err := s.system.ReadFile(s.filename); {
	case errors.Is(err, fs.ErrNotExist):
		s.state = make(map[string]map[string]string)
		return nil
	case err != nil:
		return err
	default:
		var state map[string]map[string]string
		if err := stateFormat.Unmarshal(data, &state); err != nil {
			return err
		}
		s.state = state
		return nil
	}
}
