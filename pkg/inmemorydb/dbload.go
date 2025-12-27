package inmemorydb

import (
	"bufio"
	"fmt"
	"os"
)

// load reads the database file and reconstructs the in-memory state.
func (db *DB) load() error {
	// Check if file exists
	if _, err := os.Stat(db.FilePath); os.IsNotExist(err) {
		file, err := os.Create(db.FilePath)
		if err != nil {
			return fmt.Errorf("inmemorydb: failed file creation: %w", err)
		}
		db.file = file
		db.writer = bufio.NewWriter(db.file)
		return nil
	}

	file, err := os.Open(db.FilePath)
	if err != nil {
		return fmt.Errorf("inmemorydb: failed file opening: %w", err)
	}

	db.file = file
	db.writer = bufio.NewWriter(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry, err := newEntryFromLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("inmemorydb: error reading entry at line: %w", err)
		}

		switch entry.action {
		case Put:
			db.data[string(entry.key)] = entry.value
		case Del:
			delete(db.data, string(entry.key))
		default:
			return ErrBadFormat
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("inmemorydb: failed to scan file: %w", err)
	}
	return db.Shrink()
}

// Close flushes pending writes to disk and closes the database file.
// After Close is called, the database should not be used. The in-memory data is cleared.
func (db *DB) Close() error {
	db.mutex.Lock()
	if db.closed {
		db.mutex.Unlock()
		return ErrClose
	}
	errFlush := db.writer.Flush()

	// want to close file even if flush fails
	errClose := db.file.Close()
	if errFlush != nil {
		return fmt.Errorf("inmemorydb: unable to flush writer: %w", errFlush)
	}

	db.file = nil
	db.writer = nil
	db.closed = true
	db.mutex.Unlock()

	db.Clear()
	if errClose != nil {
		return fmt.Errorf("inmemorydb: unable to close file: %w", errClose)
	}
	return nil
}

// Shrink compacts the database file by removing delete operations and rewriting only
// the current state (Put operations). This is called automatically during Load().
func (db *DB) Shrink() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.file.Close()
	if err != nil {
		return fmt.Errorf("inmemorydb: unable to close file while shrinking: %w", err)
	}

	err = os.Rename(db.FilePath, db.FilePath+".bak")
	if err != nil {
		return fmt.Errorf("inmemorydb: unable to rename %s to %s.bak while shrinking: %w", db.FilePath, db.FilePath, err)
	}

	db.file, err = os.Create(db.FilePath)
	if err != nil {
		return fmt.Errorf("inmemorydb: unable to create file while shrinking: %w", err)
	}

	db.writer = bufio.NewWriter(db.file)

	for key, value := range db.data {
		entry := newEntry(Put, key, value)
		err := db.appendEntry(entry)
		if err != nil {
			return fmt.Errorf("inmemorydb: unable to append entry: %w", err)
		}
	}

	return nil
}

func (db *DB) appendEntry(entry *entry) error {
	_, err := db.writer.Write(entry.toBytes())
	return err
}
