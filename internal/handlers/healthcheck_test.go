package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHealthCheckHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("returns health check data successfully", func(t *testing.T) {
		env := "development"
		version := "1.0.0"
		handler := NewHealthCheckHandler(logger, env, version)

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["status"] != "avaliable" {
			t.Errorf("Expected status 'avaliable', got '%s'", response["status"])
		}
		if response["env"] != env {
			t.Errorf("Expected env '%s', got '%s'", env, response["env"])
		}
		if response["version"] != version {
			t.Errorf("Expected version '%s', got '%s'", version, response["version"])
		}
	})
	t.Run("works with different environments", func(t *testing.T) {
		testCases := []struct {
			env     string
			version string
		}{
			{"development", "1.0.0"},
			{"staging", "2.0.0"},
			{"production", "3.0.0"},
			{"test", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.env, func(t *testing.T) {
				handler := NewHealthCheckHandler(logger, tc.env, tc.version)

				req := httptest.NewRequest("GET", "/healthcheck", nil)
				w := httptest.NewRecorder()

				handler(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				}

				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["env"] != tc.env {
					t.Errorf("Expected env '%s', got '%s'", tc.env, response["env"])
				}
				if response["version"] != tc.version {
					t.Errorf("Expected version '%s', got '%s'", tc.version, response["version"])
				}
			})
		}
	})
}
