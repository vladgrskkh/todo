package repository

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/pkg/inmemorydb"
)

func setupTestEnvironment(t *testing.T) (*inmemorydb.DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := inmemorydb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	return db, func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}
}

func TestTaskRepoInsertAndGet(t *testing.T) {
	db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := NewTaskRepo(db)

	t.Run("inserts and retrieves task", func(t *testing.T) {
		task := domain.NewTask(1, "Test Task", "Test Description")
		err := repo.Insert(task)
		if err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}

		retrieved, err := repo.Get(1)
		if err != nil {
			t.Fatalf("Failed to get: %v", err)
		}
		if retrieved.ID != 1 || retrieved.Title != "Test Task" {
			t.Error("Retrieved task doesn't match")
		}
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		_, err := repo.Get(999)
		if err == nil {
			t.Error("Expected error for non-existent task")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskRepoGetAll(t *testing.T) {
	db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := NewTaskRepo(db)

	t.Run("returns empty list for empty database", func(t *testing.T) {
		tasks, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("Expected empty list, got %d tasks", len(tasks))
		}
	})

	t.Run("returns all tasks", func(t *testing.T) {
		repo.Insert(domain.NewTask(1, "Task 1", "Desc 1"))
		repo.Insert(domain.NewTask(2, "Task 2", "Desc 2"))
		repo.Insert(domain.NewTask(3, "Task 3", "Desc 3"))

		tasks, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all: %v", err)
		}
		if len(tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(tasks))
		}
	})
}

func TestTaskRepoUpdate(t *testing.T) {
	db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := NewTaskRepo(db)

	t.Run("updates task successfully", func(t *testing.T) {
		task := domain.NewTask(1, "Original", "Original Description")
		repo.Insert(task)

		task.Title = "Updated"
		task.Description = "Updated Description"
		task.Done = true

		err := repo.Update(task)
		if err != nil {
			t.Fatalf("Failed to update: %v", err)
		}

		updated, err := repo.Get(1)
		if err != nil {
			t.Fatalf("Failed to get updated task: %v", err)
		}
		if updated.Title != "Updated" || !updated.Done {
			t.Error("Task not updated correctly")
		}
	})
}

func TestTaskRepoDelete(t *testing.T) {
	db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := NewTaskRepo(db)

	t.Run("deletes task successfully", func(t *testing.T) {
		err := repo.Insert(domain.NewTask(1, "Task", "Description"))
		if err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}

		err = repo.Delete(1)
		if err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}

		_, err = repo.Get(1)
		if err == nil {
			t.Error("Expected error after deletion")
		}
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		err := repo.Delete(999)
		if err == nil {
			t.Error("Expected error for non-existent task")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}
