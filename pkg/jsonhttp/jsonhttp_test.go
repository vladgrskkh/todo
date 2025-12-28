package jsonhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Run("writes valid JSON response", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := Envelope{
			"message": "success",
			"count":   1,
		}

		err := WriteJSON(w, http.StatusOK, data, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if response["message"] != "success" {
			t.Errorf("Expected message 'success', got '%v'", response["message"])
		}
		if response["count"] != float64(1) {
			t.Errorf("Expected count 1, got '%v'", response["count"])
		}
	})

	t.Run("includes custom headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := Envelope{"test": "data"}
		headers := http.Header{
			"X-Custom-Header": []string{"custom-value"},
			"Cache-Control":   []string{"no-cache"},
		}

		err := WriteJSON(w, http.StatusOK, data, headers)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		if w.Header().Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected custom header")
		}
		if w.Header().Get("Cache-Control") != "no-cache" {
			t.Errorf("Expected cache-control header")
		}
	})

	t.Run("appends newline at end", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := Envelope{"test": "data"}

		err := WriteJSON(w, http.StatusOK, data, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		body := w.Body.String()
		if len(body) == 0 {
			t.Error("Expected non-empty body")
		}
		if body[len(body)-1] != '\n' {
			t.Error("Expected body to end with newline")
		}
	})

	t.Run("handles empty envelope", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := Envelope{}

		err := WriteJSON(w, http.StatusOK, data, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if len(response) != 0 {
			t.Errorf("Expected empty object, got %v", response)
		}
	})

	t.Run("handles complex nested structures", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := Envelope{
			"users": []map[string]any{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
			"meta": Envelope{
				"total": 2,
				"page":  1,
			},
		}

		err := WriteJSON(w, http.StatusOK, data, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		users, ok := response["users"].([]interface{})
		if !ok {
			t.Error("Expected users to be array")
		}
		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}
	})
}

func TestReadJSON(t *testing.T) {
	t.Run("reads valid JSON", func(t *testing.T) {
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		jsonData := `{"title":"Test","description":"Description"}`
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(jsonData)))
		w := httptest.NewRecorder()

		err := ReadJSON(w, req, &input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if input.Title != "Test" {
			t.Errorf("Expected title 'Test', got '%s'", input.Title)
		}
		if input.Description != "Description" {
			t.Errorf("Expected description 'Description', got '%s'", input.Description)
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		var input struct {
			Title string `json:"title"`
		}

		jsonData := `{"title":"Test","unknown":"field"}`
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(jsonData)))
		w := httptest.NewRecorder()

		err := ReadJSON(w, req, &input)

		if err == nil {
			t.Error("Expected error for unknown field")
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		jsonData := `{"title":"Test",invalid}`
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(jsonData)))
		w := httptest.NewRecorder()

		var input map[string]string
		err := ReadJSON(w, req, &input)

		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("rejects oversized payload", func(t *testing.T) {
		largeString := strings.Repeat("a", 1_048_577)
		jsonData := `{"data":"` + largeString + `"}`
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(jsonData)))
		w := httptest.NewRecorder()

		var input map[string]string
		err := ReadJSON(w, req, &input)

		if err == nil {
			t.Error("Expected error for oversized payload")
		}
	})

	t.Run("handles empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("")))
		w := httptest.NewRecorder()

		var input map[string]string
		err := ReadJSON(w, req, &input)

		if err == nil {
			t.Error("Expected error for empty body")
		}
	})

	t.Run("handles whitespace in JSON", func(t *testing.T) {
		var input struct {
			Title string `json:"title"`
		}

		jsonData := `  {  "title"  :  "Test"  }  `
		req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(jsonData)))
		w := httptest.NewRecorder()

		err := ReadJSON(w, req, &input)

		if err != nil {
			t.Fatalf("Expected no error for whitespace, got %v", err)
		}

		if input.Title != "Test" {
			t.Errorf("Expected title 'Test', got '%s'", input.Title)
		}
	})

	t.Run("handles different data types", func(t *testing.T) {
		testCases := []struct {
			name     string
			jsonData string
			check    func(t *testing.T, data any)
		}{
			{
				name:     "string",
				jsonData: `{"value":"test"}`,
				check: func(t *testing.T, data any) {
					m := data.(map[string]any)
					if m["value"] != "test" {
						t.Errorf("Expected 'test', got %v", m["value"])
					}
				},
			},
			{
				name:     "number",
				jsonData: `{"value":42}`,
				check: func(t *testing.T, data any) {
					m := data.(map[string]any)
					if m["value"] != float64(42) {
						t.Errorf("Expected 42, got %v", m["value"])
					}
				},
			},
			{
				name:     "boolean",
				jsonData: `{"value":true}`,
				check: func(t *testing.T, data any) {
					m := data.(map[string]any)
					if m["value"] != true {
						t.Errorf("Expected true, got %v", m["value"])
					}
				},
			},
			{
				name:     "null",
				jsonData: `{"value":null}`,
				check: func(t *testing.T, data any) {
					m := data.(map[string]any)
					if m["value"] != nil {
						t.Errorf("Expected null, got %v", m["value"])
					}
				},
			},
			{
				name:     "array",
				jsonData: `{"value":[1,2,3]}`,
				check: func(t *testing.T, data any) {
					m := data.(map[string]any)
					arr := m["value"].([]any)
					if len(arr) != 3 {
						t.Errorf("Expected array length 3, got %d", len(arr))
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(tc.jsonData)))
				w := httptest.NewRecorder()

				var data map[string]interface{}
				err := ReadJSON(w, req, &data)

				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

				tc.check(t, data)
			})
		}
	})
}

func TestWriteJSONWithFail(t *testing.T) {
	t.Run("handles write errors", func(t *testing.T) {
		w := &failingWriter{}
		data := Envelope{"test": "data"}

		err := WriteJSON(w, http.StatusOK, data, nil)
		if err == nil {
			t.Error("Expected error from failing writer")
		}
	})
}

type failingWriter struct {
	header http.Header
	Code   int
}

func (w *failingWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}
