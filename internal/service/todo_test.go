package service

import (
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/internal/handlers/dto"
	"github.com/vladgrskkh/todo/internal/repository"
	"github.com/vladgrskkh/todo/pkg/inmemorydb"
	"github.com/vladgrskkh/todo/pkg/validator"
)

func setupTestEnvironment(t *testing.T) (*repository.TaskRepo, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := inmemorydb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	repo := repository.NewTaskRepo(db)

	return repo, func() {
		e := db.Close()
		if e != nil {
			t.Errorf("Close failed: %v", e)
		}
	}
}

func TestTodoServiceGetTask(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("returns task successfully", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(1, "Test", "Description")
		err := repo.Insert(task)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}

		result, err := service.GetTask(1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.ID != 1 {
			t.Errorf("Expected task ID 1, got %d", result.ID)
		}
		if result.Title != "Test" {
			t.Errorf("Expected title 'Test', got '%s'", result.Title)
		}
	})

	t.Run("returns error for zero ID", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		_, err := service.GetTask(0)
		if err == nil {
			t.Error("Expected error for ID 0")
		}
		if !errors.Is(err, ErrInvalidID) {
			t.Errorf("Expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("returns error for negative ID", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		_, err := service.GetTask(-1)
		if err == nil {
			t.Error("Expected error for negative ID")
		}
		if !errors.Is(err, ErrInvalidID) {
			t.Errorf("Expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("returns error when task not found", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		_, err := service.GetTask(999)
		if err == nil {
			t.Error("Expected error for non-existent task")
		}
		if !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestTodoServiceGetAllTasks(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("returns empty list when no tasks", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		tasks, err := service.GetAllTasks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("Expected empty list, got %d tasks", len(tasks))
		}
	})

	t.Run("returns all tasks", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		for i := 1; i <= 2; i++ {
			err := repo.Insert(domain.NewTask(int64(i), "Test", "Test"))
			if err != nil {
				t.Fatalf("Failed to insert task: %v", err)
			}
		}

		tasks, err := service.GetAllTasks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(tasks) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(tasks))
		}
	})
}

func TestTodoServiceCreateTask(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("creates task successfully", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(1, "New Task", "New Description")
		err := service.CreateTask(task)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		savedTask, err := repo.Get(task.ID)
		if err != nil {
			t.Error("Task was not saved to repository")
		}
		if savedTask.Title != "New Task" {
			t.Errorf("Expected title 'New Task', got '%s'", savedTask.Title)
		}
	})

	t.Run("fails to create task with invalid data", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(0, "", "Description")
		err := service.CreateTask(task)
		if err == nil {
			t.Error("Expected error for invalid task")
		}

		var validationErr *validator.Validator
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected validator error, got %T", err)
		}
	})

	t.Run("fails to create task with duplicate ID", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task1 := domain.NewTask(1, "Task 1", "Description 1")
		err := service.CreateTask(task1)
		if err != nil {
			t.Errorf("Expected no error for first task, got %v", err)
		}

		task2 := domain.NewTask(1, "Task 2", "Description 2")
		err = service.CreateTask(task2)
		if err == nil {
			t.Error("Expected error for duplicate ID")
		}
		if !errors.Is(err, ErrTaskExists) {
			t.Errorf("Expected ErrTaskExists, got %v", err)
		}
	})
}

func TestTodoServiceUpdateTask(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("updates task successfully", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(1, "Original", "Original Description")
		err := repo.Insert(task)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}

		input := dto.UpdateTaskInput{
			Title:       "Updated",
			Description: "Updated Description",
			Done:        true,
		}

		updatedTask, err := service.UpdateTask(1, input)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if updatedTask.Title != "Updated" {
			t.Errorf("Expected title 'Updated', got '%s'", updatedTask.Title)
		}
		if !updatedTask.Done {
			t.Error("Expected task to be done")
		}
	})

	t.Run("fails to update non-existent task", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		input := dto.UpdateTaskInput{
			Title:       "Updated",
			Description: "Updated Description",
			Done:        false,
		}

		_, err := service.UpdateTask(999, input)
		if err == nil {
			t.Error("Expected error for non-existent task")
		}
		if !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("fails to update with invalid data", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(1, "Original", "Original Description")
		err := repo.Insert(task)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}

		input := dto.UpdateTaskInput{
			Title:       "", // Invalid: empty title
			Description: "Updated Description",
			Done:        false,
		}

		_, err = service.UpdateTask(1, input)
		if err == nil {
			t.Error("Expected error for invalid data")
		}

		var validationErr *validator.Validator
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected validator error, got %T", err)
		}
	})

	t.Run("fails to update completed task", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(1, "Original", "Original Description")
		task.Done = true
		err := repo.Insert(task)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}

		input := dto.UpdateTaskInput{
			Title:       "Updated",
			Description: "Updated Description",
			Done:        false,
		}

		_, err = service.UpdateTask(1, input)
		if err == nil {
			t.Error("Expected error for completed task")
		}

		var validationErr *validator.Validator
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected validator error, got %T", err)
		}
	})
}

func TestTodoServiceDeleteTask(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("deletes task successfully", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		task := domain.NewTask(1, "Task", "Description")
		err := repo.Insert(task)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}

		err = service.DeleteTask(1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		_, err = repo.Get(task.ID)
		if err == nil {
			t.Error("Task was not deleted from repository")
		}
	})

	t.Run("fails to delete with invalid ID", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		err := service.DeleteTask(0)
		if err == nil {
			t.Error("Expected error for ID 0")
		}
		if !errors.Is(err, ErrInvalidID) {
			t.Errorf("Expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("fails to delete non-existent task", func(t *testing.T) {
		repo, cleanup := setupTestEnvironment(t)
		defer cleanup()

		service := NewTodoService(logger, repo)

		err := service.DeleteTask(999)
		if err == nil {
			t.Error("Expected error for non-existent task")
		}
		if !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}
