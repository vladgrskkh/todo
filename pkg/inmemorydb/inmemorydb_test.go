package inmemorydb

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Done        bool
}

func encodeTask(task *Task) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(task)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decodeTask(data []byte) (*Task, error) {
	var task Task
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func TestPutAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	task := &Task{ID: 1, Title: "Buy groceries", Description: "Get milk and bread"}
	buf, err := encodeTask(task)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = db.PutObject(strconv.Itoa(int(task.ID)), buf)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}

	retrievedData, err := db.GetObject("1")
	if err != nil {
		t.Errorf("GetObject failed: %v", err)
	}

	retrieved, err := decodeTask(retrievedData)
	if err != nil {
		t.Errorf("decode error: %v", err)
	}

	if retrieved.ID != task.ID || retrieved.Title != task.Title || retrieved.Description != task.Description {
		t.Errorf("Retrieved task doesn't match original. Got: %+v, Want: %+v", retrieved, task)
	}
}

func TestPutAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	task := &Task{ID: 1, Title: "Test Task", Description: "Test Description"}
	taskData, err := encodeTask(task)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = db.PutObject("key", taskData)
	if err != nil {
		t.Errorf("First PutObject failed: %v", err)
	}

	err = db.PutObject("key", taskData)
	if err != ErrAlreadyExists {
		t.Errorf("Expected ErrAlreadyExists, got: %v", err)
	}
}

func TestGetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	_, err = db.GetObject("nonexistent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestDeleteObject(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	task := &Task{ID: 1, Title: "Delete Test", Description: "Task to delete"}
	taskData, err := encodeTask(task)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}
	err = db.PutObject("key", taskData)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}

	err = db.DeleteObject("key")
	if err != nil {
		t.Errorf("DeleteObject failed: %v", err)
	}

	if db.Has("key") {
		t.Error("Key should not exist after deletion")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	err = db.DeleteObject("nonexistent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestHas(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	task := &Task{ID: 1, Title: "Has Test", Description: "Test Has"}
	taskData, err := encodeTask(task)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}
	err = db.PutObject("key", taskData)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}

	if !db.Has("key") {
		t.Error("Has should return true for existing key")
	}

	if db.Has("nonexistent") {
		t.Error("Has should return false for non-existent key")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	task1 := &Task{ID: 1, Title: "Clear Test 1", Description: "First"}
	task2 := &Task{ID: 2, Title: "Clear Test 2", Description: "Second"}

	task1Data, err := encodeTask(task1)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}
	task2Data, err := encodeTask(task2)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = db.PutObject("key1", task1Data)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}
	err = db.PutObject("key2", task2Data)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}
	db.Clear()

	if db.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got: %d", db.Size())
	}
}

func TestSize(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	task1 := &Task{ID: 1, Title: "Size Test 1", Description: "First"}
	task2 := &Task{ID: 2, Title: "Size Test 2", Description: "Second"}

	task1Data, err := encodeTask(task1)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}
	task2Data, err := encodeTask(task2)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = db.PutObject("key1", task1Data)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}
	err = db.PutObject("key2", task2Data)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}
	_ = db.PutObject("key2", task2Data)

	if db.Size() != 2 {
		t.Errorf("Expected size 2, got: %d", db.Size())
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_db.dat")

	db1, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}
	task1 := &Task{ID: 1, Title: "Task 1", Description: "First task"}
	task2 := &Task{ID: 2, Title: "Task 2", Description: "Second task"}

	buf1, err := encodeTask(task1)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	buf2, err := encodeTask(task2)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = db1.PutObject("task:1", buf1)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}
	err = db1.PutObject("task:2", buf2)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}
	err = db1.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	db2, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db2.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	retrieved1, err := db2.GetObject("task:1")
	if err != nil {
		t.Errorf("Failed to get task:1 after load: %v", err)
	}

	retrieved2, err := db2.GetObject("task:2")
	if err != nil {
		t.Errorf("Failed to get task:2 after load: %v", err)
	}

	decodedTask1, err := decodeTask(retrieved1)
	if err != nil {
		t.Errorf("Failed to decode task:1 after load: %v", err)
	}

	decodedTask2, err := decodeTask(retrieved2)
	if err != nil {
		t.Errorf("Failed to decode task:2 after load: %v", err)
	}

	if decodedTask1.Title != "Task 1" || decodedTask2.Title != "Task 2" {
		t.Error("Data not loaded correctly")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent_db.dat")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Open should have created the file")
	}

	if db.Size() != 0 {
		t.Errorf("Expected empty database after loading non-existent file, got size: %d", db.Size())
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_db.dat")

	db, err := Open(dbPath)
	if err != nil {
		t.Errorf("Open failed: %v", err)
	}
	task := &Task{ID: 1, Title: "Close Test", Description: "Task to save"}
	taskData, err := encodeTask(task)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = db.PutObject("key", taskData)
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if _, err := db.GetObject("key"); err == nil {
		t.Error("Database should be closed after Close()")
	}

	if err := db.PutObject("key", []byte{0}); err == nil {
		t.Error("Database should be closed after Close()")
	}

	if err := db.DeleteObject("key"); err == nil {
		t.Error("Database should be closed after Close()")
	}

	// Test that Has() returns false after Close()
	if ok := db.Has("key"); ok == true {
		t.Error("Database should be closed after Close()")
	}
}
