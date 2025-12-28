package integrationtest

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/internal/handlers/dto"
	"github.com/vladgrskkh/todo/internal/handlers/middleware/metrics"
	"github.com/vladgrskkh/todo/internal/handlers/routes"
	"github.com/vladgrskkh/todo/internal/repository"
	"github.com/vladgrskkh/todo/internal/service"
	"github.com/vladgrskkh/todo/pkg/inmemorydb"
)

func init() {
	if metrics.TotalTasksCreated == nil {
		metrics.InitMetrics()
	}
}

func setupTestEnvironment(t *testing.T) (*service.TodoService, *repository.TaskRepo, *inmemorydb.DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := inmemorydb.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	repo := repository.NewTaskRepo(db)
	s := service.NewTodoService(slog.New(slog.NewTextHandler(os.Stdout, nil)), repo)

	cleanup := func() {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}

	return s, repo, db, cleanup
}

func TestIntegrationFullTaskWorkflow(t *testing.T) {
	s, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := routes.Routes(logger, s, "test", "1.0.0")

	t.Run("complete task lifecycle", func(t *testing.T) {
		// Create a task
		createInput := dto.CreateTaskInput{
			ID:          1,
			Title:       "Integration Test Task",
			Description: "Testing full lifecycle",
		}
		createBody, _ := json.Marshal(createInput)

		createReq := httptest.NewRequest("POST", "/todos", bytes.NewReader(createBody))
		createW := httptest.NewRecorder()
		handler.ServeHTTP(createW, createReq)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, createW.Code, createW.Body.String())
		}

		var createResponse map[string]domain.Task
		if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}
		createdTask := createResponse["task"]
		if createdTask.ID != 1 {
			t.Errorf("Expected task ID 1, got %d", createdTask.ID)
		}

		// Get the created task
		getReq := httptest.NewRequest("GET", "/todos/1", nil)
		getReq.SetPathValue("id", "1")
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, getW.Code)
		}

		var getResponse map[string]domain.Task
		if err := json.Unmarshal(getW.Body.Bytes(), &getResponse); err != nil {
			t.Fatalf("Failed to unmarshal get response: %v", err)
		}
		retrievedTask := getResponse["task"]
		if retrievedTask.Title != "Integration Test Task" {
			t.Errorf("Expected title 'Integration Test Task', got '%s'", retrievedTask.Title)
		}

		// Update the task
		updateInput := dto.UpdateTaskInput{
			Title:       "Updated Integration Task",
			Description: "Updated description",
			Done:        true,
		}
		updateBody, _ := json.Marshal(updateInput)

		updateReq := httptest.NewRequest("PUT", "/todos/1", bytes.NewReader(updateBody))
		updateReq.SetPathValue("id", "1")
		updateW := httptest.NewRecorder()
		handler.ServeHTTP(updateW, updateReq)

		if updateW.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, updateW.Code)
		}

		var updateResponse map[string]domain.Task
		if err := json.Unmarshal(updateW.Body.Bytes(), &updateResponse); err != nil {
			t.Fatalf("Failed to unmarshal update response: %v", err)
		}
		updatedTask := updateResponse["task"]
		if updatedTask.Title != "Updated Integration Task" {
			t.Errorf("Expected updated title 'Updated Integration Task', got '%s'", updatedTask.Title)
		}
		if !updatedTask.Done {
			t.Error("Expected task to be done")
		}

		// Get all tasks
		getAllReq := httptest.NewRequest("GET", "/todos", nil)
		getAllW := httptest.NewRecorder()
		handler.ServeHTTP(getAllW, getAllReq)

		if getAllW.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, getAllW.Code)
		}

		var getAllResponse map[string][]domain.Task
		if err := json.Unmarshal(getAllW.Body.Bytes(), &getAllResponse); err != nil {
			t.Fatalf("Failed to unmarshal get all response: %v", err)
		}
		if len(getAllResponse["tasks"]) != 1 {
			t.Errorf("Expected 1 task, got %d", len(getAllResponse["tasks"]))
		}

		// Delete the task
		deleteReq := httptest.NewRequest("DELETE", "/todos/1", nil)
		deleteReq.SetPathValue("id", "1")
		deleteW := httptest.NewRecorder()
		handler.ServeHTTP(deleteW, deleteReq)

		if deleteW.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, deleteW.Code)
		}

		// Verify task is deleted
		getDeletedReq := httptest.NewRequest("GET", "/todos/1", nil)
		getDeletedReq.SetPathValue("id", "1")
		getDeletedW := httptest.NewRecorder()
		handler.ServeHTTP(getDeletedW, getDeletedReq)

		if getDeletedW.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for deleted task, got %d", http.StatusNotFound, getDeletedW.Code)
		}
	})
}

func TestIntegrationMultipleTasks(t *testing.T) {
	s, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := routes.Routes(logger, s, "test", "1.0.0")

	t.Run("create and manage multiple tasks", func(t *testing.T) {
		// Create multiple tasks
		tasks := []dto.CreateTaskInput{
			{ID: 1, Title: "Task 1", Description: "Description 1"},
			{ID: 2, Title: "Task 2", Description: "Description 2"},
			{ID: 3, Title: "Task 3", Description: "Description 3"},
		}

		for _, task := range tasks {
			body, _ := json.Marshal(task)
			req := httptest.NewRequest("POST", "/todos", bytes.NewReader(body))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Fatalf("Failed to create task %d: %d", task.ID, w.Code)
			}
		}

		// Get all tasks
		req := httptest.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string][]domain.Task
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if len(response["tasks"]) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(response["tasks"]))
		}

		// Update task 2
		updateInput := dto.UpdateTaskInput{
			Title:       "Updated Task 2",
			Description: "Updated Description 2",
			Done:        true,
		}
		updateBody, _ := json.Marshal(updateInput)
		updateReq := httptest.NewRequest("PUT", "/todos/2", bytes.NewReader(updateBody))
		updateReq.SetPathValue("id", "2")
		updateW := httptest.NewRecorder()
		handler.ServeHTTP(updateW, updateReq)

		if updateW.Code != http.StatusOK {
			t.Fatalf("Failed to update task 2: %d", updateW.Code)
		}

		// Delete task 1
		deleteReq := httptest.NewRequest("DELETE", "/todos/1", nil)
		deleteReq.SetPathValue("id", "1")
		deleteW := httptest.NewRecorder()
		handler.ServeHTTP(deleteW, deleteReq)

		if deleteW.Code != http.StatusOK {
			t.Fatalf("Failed to delete task 1: %d", deleteW.Code)
		}

		// Verify final state
		finalReq := httptest.NewRequest("GET", "/todos", nil)
		finalW := httptest.NewRecorder()
		handler.ServeHTTP(finalW, finalReq)

		var finalResponse map[string][]domain.Task
		err := json.Unmarshal(finalW.Body.Bytes(), &finalResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if len(finalResponse["tasks"]) != 2 {
			t.Errorf("Expected 2 remaining tasks, got %d", len(finalResponse["tasks"]))
		}

		// Verify task 2 is updated
		var task2Found bool
		for _, task := range finalResponse["tasks"] {
			if task.ID == 2 {
				task2Found = true
				if task.Title != "Updated Task 2" {
					t.Errorf("Expected task 2 title 'Updated Task 2', got '%s'", task.Title)
				}
				if !task.Done {
					t.Error("Expected task 2 to be done")
				}
			}
			if task.ID == 1 {
				t.Error("Task 1 should have been deleted")
			}
		}
		if !task2Found {
			t.Error("Task 2 not found in final results")
		}
	})
}

func TestIntegrationErrorScenarios(t *testing.T) {
	s, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := routes.Routes(logger, s, "test", "1.0.0")

	t.Run("duplicate ID error", func(t *testing.T) {
		// Create first task
		createInput := dto.CreateTaskInput{
			ID:          1,
			Title:       "Task 1",
			Description: "Description",
		}
		body, _ := json.Marshal(createInput)
		req := httptest.NewRequest("POST", "/todos", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("First task creation failed: %d", w.Code)
		}

		// Try to create task with same ID
		req = httptest.NewRequest("POST", "/todos", bytes.NewReader(body))
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected conflict status for duplicate ID, got %d", w.Code)
		}
	})

	t.Run("invalid ID parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/todos/invalid", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected bad request for invalid ID, got %d", w.Code)
		}
	})

	t.Run("update non-existent task", func(t *testing.T) {
		updateInput := dto.UpdateTaskInput{
			Title:       "Updated",
			Description: "Updated",
			Done:        false,
		}
		body, _ := json.Marshal(updateInput)
		req := httptest.NewRequest("PUT", "/todos/999", bytes.NewReader(body))
		req.SetPathValue("id", "999")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected not found for non-existent task, got %d", w.Code)
		}
	})

	t.Run("delete non-existent task", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/todos/999", nil)
		req.SetPathValue("id", "999")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected not found for non-existent task, got %d", w.Code)
		}
	})

	t.Run("get non-existent task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/todos/999", nil)
		req.SetPathValue("id", "999")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected not found for non-existent task, got %d", w.Code)
		}
	})

	t.Run("invalid JSON in create request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/todos", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected bad request for invalid JSON, got %d", w.Code)
		}
	})

	t.Run("invalid JSON in update request", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/todos/1", bytes.NewReader([]byte("invalid json")))
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected bad request for invalid JSON, got %d", w.Code)
		}
	})
}
