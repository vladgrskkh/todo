// Package inmemorydb provides a persistent in-memory key-value database.
// It maintains data in memory while persisting operations to disk in a log-based format.
// All operations are thread-safe using RWMutex. The database stores arbitrary byte slices
// as values, allowing users to serialize custom types using encoding/gob or other mechanisms.
package inmemorydb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	ErrNotFound      = errors.New("key not found")
	ErrAlreadyExists = errors.New("key already exists")
	ErrInvalidType   = errors.New("invalid type")
	ErrClose         = errors.New("database is closed")
)

// DB represents an in-memory key-value database with persistent storage.
// All operations on DB are thread-safe.
type DB struct {
	FilePath string
	closed   bool
	data     map[string][]byte
	mutex    sync.RWMutex
	file     *os.File
	writer   *bufio.Writer
}

// Open creates and returns a new database instance. It loads existing data from the file
// at filePath, creating the file if it doesn't exist.
// The returned DB should be closed with Close() when no longer needed.
//
// Dont open the same file twice. Opening the same file twice will result in ub.
func Open(filePath string) (*DB, error) {
	db := &DB{
		data:     make(map[string][]byte),
		FilePath: filePath,
	}
	err := db.load()
	if err != nil {
		return nil, fmt.Errorf("inmemorydb: failed to load database: %w", err)
	}

	return db, nil
}

// PutObject stores a value in the database at the given key. Returns ErrAlreadyExists
// if the key already exists. The operation is persisted to disk.
func (db *DB) PutObject(key string, value []byte) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if db.closed {
		return ErrClose
	}

	// Check if key already exists
	if _, exists := db.data[key]; exists {
		return ErrAlreadyExists
	}

	db.data[key] = value
	return db.appendEntry(newEntry(Put, key, value))
}

// GetObject retrieves the value associated with the given key.
// Returns ErrNotFound if the key does not exist.
func (db *DB) GetObject(key string) ([]byte, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	if db.closed {
		return nil, ErrClose
	}

	data, exists := db.data[key]
	if !exists {
		return data, ErrNotFound
	}

	return data, nil
}

// DeleteObject removes the value associated with the given key from the database.
// Returns ErrNotFound if the key does not exist. The operation is persisted to disk.
func (db *DB) DeleteObject(key string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if db.closed {
		return ErrClose
	}

	if _, exists := db.data[key]; !exists {
		return ErrNotFound
	}

	delete(db.data, key)
	return db.appendEntry(newEntry(Del, key, nil))
}

// Has returns true if the given key exists in the database, false otherwise.
func (db *DB) Has(key string) bool {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	if db.closed {
		return false
	}

	_, exists := db.data[key]
	return exists
}

// Clear removes all entries from the database. This only clears the in-memory data;
// previously persisted entries will be reloaded on the next Open.
func (db *DB) Clear() {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.data = make(map[string][]byte)
}

// Size returns the number of keys currently stored in the database.
func (db *DB) Size() int {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	return len(db.data)
}
