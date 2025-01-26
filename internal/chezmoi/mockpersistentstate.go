package chezmoi

// A MockPersistentState is a mock persistent state.
type MockPersistentState struct {
	buckets map[string]map[string][]byte
}

// NewMockPersistentState returns a new PersistentState.
func NewMockPersistentState() *MockPersistentState {
	return &MockPersistentState{
		buckets: make(map[string]map[string][]byte),
	}
}

// Close closes s.
func (s *MockPersistentState) Close() error {
	return nil
}

// CopyTo implements PersistentState.CopyTo.
func (s *MockPersistentState) CopyTo(p PersistentState) error {
	for bucket, bucketMap := range s.buckets {
		for key, value := range bucketMap {
			if err := p.Set([]byte(bucket), []byte(key), value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Data implements PersistentState.Data.
func (s *MockPersistentState) Data() (map[string]map[string]string, error) {
	data := make(map[string]map[string]string)
	for bucket, bucketMap := range s.buckets {
		dataBucketMap := make(map[string]string, len(bucketMap))
		for key, value := range bucketMap {
			dataBucketMap[key] = string(value)
		}
		data[bucket] = dataBucketMap
	}
	return data, nil
}

// Delete implements PersistentState.Delete.
func (s *MockPersistentState) Delete(bucket, key []byte) error {
	bucketMap, ok := s.buckets[string(bucket)]
	if !ok {
		return nil
	}
	delete(bucketMap, string(key))
	return nil
}

// DeleteBucket implements PersistentState.DeleteBucket.
func (s *MockPersistentState) DeleteBucket(bucket []byte) error {
	delete(s.buckets, string(bucket))
	return nil
}

// ForEach implements PersistentState.ForEach.
func (s *MockPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	for k, v := range s.buckets[string(bucket)] {
		if err := fn([]byte(k), v); err != nil {
			return err
		}
	}
	return nil
}

// Get implements PersistentState.Get.
func (s *MockPersistentState) Get(bucket, key []byte) ([]byte, error) {
	bucketMap, ok := s.buckets[string(bucket)]
	if !ok {
		return nil, nil
	}
	return bucketMap[string(key)], nil
}

// Set implements PersistentState.Set.
func (s *MockPersistentState) Set(bucket, key, value []byte) error {
	bucketMap, ok := s.buckets[string(bucket)]
	if !ok {
		bucketMap = make(map[string][]byte)
		s.buckets[string(bucket)] = bucketMap
	}
	bucketMap[string(key)] = value
	return nil
}
