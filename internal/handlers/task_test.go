package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/internal/handlers/dto"
	"github.com/vladgrskkh/todo/internal/handlers/middleware/metrics"
	"github.com/vladgrskkh/todo/internal/handlers/mocks"
	"github.com/vladgrskkh/todo/internal/repository"
	"github.com/vladgrskkh/todo/internal/service"
	"github.com/vladgrskkh/todo/pkg/validator"
)

func init() {
	// Initialize metrics before any tests run
	// Check if already initialized to avoid panic
	if metrics.TotalTasksCreated == nil {
		metrics.InitMetrics()
	}
}

func TestNewGetTaskHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tests := []struct {
		name            string
		taskId          int64
		tastTitle       string
		taskDescription string
		getErr          error
		getAllErr       error
		url             string
		excpectedCode   int
	}{
		{
			name:          "returns bad request for invalid ID(not int)",
			url:           "/todos/invalid",
			excpectedCode: http.StatusBadRequest,
		},
		{
			name:          "returns bad request for invalid ID(less than 1)",
			getErr:        service.ErrInvalidID,
			url:           "/todos/-100",
			excpectedCode: http.StatusBadRequest,
		},
		{
			name:          "returns not found for missing task",
			getErr:        repository.ErrNotFound,
			url:           "/todos/1",
			excpectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTaskGetter(domain.NewTask(tt.taskId, tt.tastTitle, tt.taskDescription), nil, tt.getErr, tt.getAllErr)
			handler := NewGetTaskHandler(logger, mockService)

			req := httptest.NewRequest("GET", tt.url, nil)
			req.SetPathValue("id", strings.Split(tt.url, "/")[2])
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.excpectedCode {
				t.Errorf("Expected status %d, got %d", tt.excpectedCode, w.Code)
			}
		})
	}

	t.Run("returns task successfully", func(t *testing.T) {
		mockService := mocks.NewMockTaskGetter(domain.NewTask(1, "Test Task", "Test Description"), nil, nil, nil)
		handler := NewGetTaskHandler(logger, mockService)

		req := httptest.NewRequest("GET", "/todos/1", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]domain.Task
		json.Unmarshal(w.Body.Bytes(), &response)
		task := response["task"]
		if task.ID != 1 || task.Title != "Test Task" {
			t.Error("Task data incorrect")
		}
	})
}

func TestNewGetAllTasksHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tests := []struct {
		name          string
		tasks         []*domain.Task
		getAllErr     error
		excpectedCode int
	}{
		{
			name: "returns all tasks",
			tasks: []*domain.Task{
				domain.NewTask(1, "Task 1", "Description 1"),
				domain.NewTask(2, "Task 2", "Description 2"),
			},
			excpectedCode: http.StatusOK,
		},
		{
			name:          "returns nil tasks",
			excpectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTaskGetter(nil, tt.tasks, nil, tt.getAllErr)
			handler := NewGetAllTasksHandler(logger, mockService)

			req := httptest.NewRequest("GET", "/todos", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.excpectedCode {
				t.Errorf("Expected status %d, got %d", tt.excpectedCode, w.Code)
			}

			var response map[string][]domain.Task
			json.Unmarshal(w.Body.Bytes(), &response)
			if len(response["tasks"]) != len(tt.tasks) {
				t.Errorf("Expected %d tasks, got %d", len(tt.tasks), len(response["tasks"]))
			}
		})
	}
}

func TestNewPostTaskHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tests := []struct {
		name         string
		input        dto.CreateTaskInput
		createErr    error
		expectedCode int
	}{
		{
			name: "creates task successfully",
			input: dto.CreateTaskInput{
				ID:          1,
				Title:       "New Task",
				Description: "New Description",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "returns conflict for duplicate task",
			input: dto.CreateTaskInput{
				ID:          1,
				Title:       "Duplicate",
				Description: "Duplicate",
			},
			createErr:    service.ErrTaskExists,
			expectedCode: http.StatusConflict,
		},
		{
			name: "failed validation for task",
			input: dto.CreateTaskInput{
				ID:          -1,
				Description: "Duplicate",
			},
			createErr:    validator.New(),
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTaskCreator(tt.createErr)
			handler := NewPostTaskHandler(logger, mockService)

			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %v", err)
			}

			req := httptest.NewRequest("POST", "/todos", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		mockService := mocks.NewMockTaskCreator(nil)
		handler := NewPostTaskHandler(logger, mockService)

		req := httptest.NewRequest("POST", "/todos", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestNewTaskUpdater(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tests := []struct {
		name         string
		task         *domain.Task
		input        dto.UpdateTaskInput
		updateErr    error
		expectedCode int
		url          string
	}{
		{
			name: "updates task successfully",
			task: domain.NewTask(1, "Updated", "Updated Description"),
			input: dto.UpdateTaskInput{
				Title:       "Updated",
				Description: "Updated Description",
				Done:        true,
			},
			expectedCode: http.StatusOK,
			url:          "/todos/1",
		},
		{
			name:         "returns bad request for invalid ID",
			expectedCode: http.StatusBadRequest,
			url:          "/todos/invalid",
		},
		{
			name:         "returns not found for missing task",
			expectedCode: http.StatusNotFound,
			updateErr:    repository.ErrNotFound,
			url:          "/todos/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTaskUpdater(tt.task, tt.updateErr)
			handler := NewTaskUpdater(logger, mockService)

			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %v", err)
			}

			req := httptest.NewRequest("PUT", tt.url, bytes.NewReader(body))
			req.SetPathValue("id", strings.Split(tt.url, "/")[2])
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.task != nil {
				var response map[string]*domain.Task
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				if response["task"].Title != tt.task.Title || response["task"].Description != tt.task.Description || response["task"].Done != tt.task.Done {
					t.Errorf("Expected task %v, got %v", tt.task, response["task"])
				}
			}
		})
	}
}

func TestNewDeleteTaskHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tests := []struct {
		name         string
		deleteErr    error
		expectedCode int
		url          string
	}{
		{
			name:         "deletes task successfully",
			expectedCode: http.StatusOK,
			url:          "/todos/1",
		},
		{
			name:         "returns bad request for invalid ID(not int)",
			expectedCode: http.StatusBadRequest,
			url:          "/todos/invalid",
		},
		{
			name:         "returns not found for missing task",
			expectedCode: http.StatusNotFound,
			deleteErr:    repository.ErrNotFound,
			url:          "/todos/1",
		},
		{
			name:         "returns bad request for invalid ID(less than 1)",
			expectedCode: http.StatusBadRequest,
			deleteErr:    service.ErrInvalidID,
			url:          "/todos/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTaskDeleter(tt.deleteErr)
			handler := NewDeleteTaskHandler(logger, mockService)

			req := httptest.NewRequest("DELETE", tt.url, nil)
			req.SetPathValue("id", strings.Split(tt.url, "/")[2])
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}
