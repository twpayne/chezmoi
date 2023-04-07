package chezmoi

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite" // Import pure Go SQLite driver
)

const (
	createSchemaQuery = `
		CREATE TABLE IF NOT EXISTS state (
			bucket BLOB NOT NULL,
			key BLOB NOT NULL,
			value BLOB NOT NULL,
			PRIMARY KEY (bucket, key)
		);
	`
	dataQuery = `
		SELECT bucket, key, value FROM state;
	`
	deleteQuery = `
		DELETE FROM state WHERE bucket = $1 AND key = $2;
	`
	deleteBucketQuery = `
		DELETE FROM state WHERE bucket = $1;
	`
	forEachQuery = `
		SELECT key, value FROM state WHERE bucket = $1;
	`
	getQuery = `
		SELECT value FROM state WHERE bucket = $1 AND key = $2;
	`
	setQuery = `
		INSERT OR REPLACE INTO state (bucket, key, value) VALUES ($1, $2, $3);
	`
)

type SQLitePersistentState struct {
	db *sql.DB
}

func NewSQLitePersistentState(dataSourceName string) (*SQLitePersistentState, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(createSchemaQuery); err != nil {
		return nil, err
	}
	return &SQLitePersistentState{
		db: db,
	}, nil
}

func (s *SQLitePersistentState) Close() error {
	return s.db.Close()
}

func (s *SQLitePersistentState) CopyTo(other PersistentState) error {
	return s.forAll(other.Set)
}

func (s *SQLitePersistentState) Data() (any, error) {
	data := make(map[string]map[string]string)
	if err := s.forAll(func(bucket, key, value []byte) error {
		bucketMap, ok := data[string(bucket)]
		if !ok {
			bucketMap = make(map[string]string)
			data[string(bucket)] = bucketMap
		}
		bucketMap[string(key)] = string(value)
		return nil
	}); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *SQLitePersistentState) Delete(bucket, key []byte) error {
	_, err := s.db.Exec(deleteQuery, bucket, key)
	return err
}

func (s *SQLitePersistentState) DeleteBucket(bucket []byte) error {
	_, err := s.db.Exec(deleteBucketQuery, bucket)
	return err
}

func (s *SQLitePersistentState) ForEach(bucket []byte, fn func([]byte, []byte) error) error {
	rows, err := s.db.Query(forEachQuery, bucket)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return err
		}
		if err := fn(key, value); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *SQLitePersistentState) Get(bucket, key []byte) ([]byte, error) {
	var value []byte
	switch err := s.db.QueryRow(getQuery, bucket, key).Scan(&value); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return value, nil
	}
}

func (s *SQLitePersistentState) Set(bucket, key, value []byte) error {
	_, err := s.db.Exec(setQuery, bucket, key, value)
	return err
}

func (s *SQLitePersistentState) forAll(fn func([]byte, []byte, []byte) error) error {
	rows, err := s.db.Query(dataQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var bucket, key, value []byte
		if err := rows.Scan(&bucket, &key, &value); err != nil {
			return err
		}
		if err := fn(bucket, key, value); err != nil {
			return err
		}
	}
	return rows.Err()
}
