package validator

import (
	"testing"
	"unicode/utf8"
)

func TestValid(t *testing.T) {
	t.Run("returns true when no errors", func(t *testing.T) {
		v := New()

		if !v.Valid() {
			t.Error("Expected Valid() to return true for validator with no errors")
		}
	})

	t.Run("returns false when errors exist", func(t *testing.T) {
		v := New()
		v.AddError("field", "error message")

		if v.Valid() {
			t.Error("Expected Valid() to return false for validator with errors")
		}
	})
}

func TestAddError(t *testing.T) {
	t.Run("adds error to empty validator", func(t *testing.T) {
		v := New()

		v.AddError("field", "error message")

		if len(v.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(v.Errors))
		}

		if msg, exists := v.Errors["field"]; !exists {
			t.Error("Expected error for key 'field' to exist")
		} else if msg != "error message" {
			t.Errorf("Expected error message 'error message', got '%s'", msg)
		}
	})

	t.Run("does not overwrite existing error for same key", func(t *testing.T) {
		v := New()

		v.AddError("field", "first error")
		v.AddError("field", "second error")

		if len(v.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(v.Errors))
		}

		if msg := v.Errors["field"]; msg != "first error" {
			t.Errorf("Expected original error message 'first error', got '%s'", msg)
		}
	})

	t.Run("handles empty key and message", func(t *testing.T) {
		v := New()

		v.AddError("", "")

		if len(v.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(v.Errors))
		}

		if msg, exists := v.Errors[""]; !exists {
			t.Error("Expected error for empty key to exist")
		} else if msg != "" {
			t.Errorf("Expected empty error message, got '%s'", msg)
		}
	})
}

func TestCheck(t *testing.T) {
	t.Run("adds error when condition is false", func(t *testing.T) {
		v := New()

		v.Check(false, "field", "validation failed")

		if len(v.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(v.Errors))
		}

		if msg, exists := v.Errors["field"]; !exists {
			t.Error("Expected error for key 'field' to exist")
		} else if msg != "validation failed" {
			t.Errorf("Expected error message 'validation failed', got '%s'", msg)
		}
	})

	t.Run("does not add error when condition is true", func(t *testing.T) {
		v := New()

		v.Check(true, "field", "validation failed")

		if len(v.Errors) != 0 {
			t.Errorf("Expected 0 errors, got %d", len(v.Errors))
		}
	})
}

func TestValidatorIntegration(t *testing.T) {
	t.Run("typical validation workflow", func(t *testing.T) {
		v := New()

		id := -1
		title := ""
		description := "123"

		v.Check(id > 0, "id", "must be a positive integer")

		v.Check(title != "", "title", "must be provided")
		v.Check(utf8.RuneCountInString(title) <= 100, "title", "must not be more than 100 symbols long")

		v.Check(utf8.RuneCountInString(description) <= 2000, "description", "must not be more than 2000 symbols long")
		if v.Valid() {
			t.Error("Expected validation to fail")
		}

		if len(v.Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(v.Errors))
		}

		expectedErrors := map[string]string{
			"id":    "must be a positive integer",
			"title": "must be provided",
		}

		for key, expectedMsg := range expectedErrors {
			if msg, exists := v.Errors[key]; !exists {
				t.Errorf("Expected error for key '%s'", key)
			} else if msg != expectedMsg {
				t.Errorf("For key '%s', expected '%s', got '%s'", key, expectedMsg, msg)
			}
		}
	})

	t.Run("successful validation workflow", func(t *testing.T) {
		v := New()

		id := 1
		title := "DLQ implementation"
		description := "Implement dead letter queue for the task service"

		v.Check(id > 0, "id", "must be a positive integer")

		v.Check(title != "", "title", "must be provided")
		v.Check(utf8.RuneCountInString(title) <= 100, "title", "must not be more than 100 symbols long")

		v.Check(utf8.RuneCountInString(description) <= 2000, "description", "must not be more than 2000 symbols long")

		if !v.Valid() {
			t.Error("Expected validation to succeed")
		}

		if len(v.Errors) != 0 {
			t.Errorf("Expected 0 errors, got %d: %v", len(v.Errors), v.Errors)
		}
	})
}
