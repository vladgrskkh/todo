package domain

import (
	"testing"

	"github.com/vladgrskkh/todo/pkg/validator"
)

func TestTaskUpdate(t *testing.T) {
	t.Run("updates task successfully", func(t *testing.T) {
		task := NewTask(1, "Original", "Original Description")
		v := validator.New()

		task.Update(v, "Updated", "Updated Description", true)

		if task.Title != "Updated" {
			t.Errorf("Expected title 'Updated', got '%s'", task.Title)
		}
		if task.Description != "Updated Description" {
			t.Errorf("Expected description 'Updated Description', got '%s'", task.Description)
		}
		if !task.Done {
			t.Error("Expected task to be done")
		}
		if task.version != 2 {
			t.Errorf("Expected version 2, got %d", task.version)
		}
		if !v.Valid() {
			t.Errorf("Expected validator to be valid, got errors: %v", v.Errors)
		}
	})

	t.Run("fails to update completed task", func(t *testing.T) {
		task := NewTask(1, "Original", "Original Description")
		task.Done = true
		v := validator.New()

		task.Update(v, "Updated", "Updated Description", false)

		if v.Valid() {
			t.Error("Expected validator to be invalid for completed task")
		}
		if _, exists := v.Errors["done"]; !exists {
			t.Error("Expected 'done' error to exist")
		}
	})

	t.Run("updates task without changing done status", func(t *testing.T) {
		task := NewTask(1, "Original", "Original Description")
		v := validator.New()

		task.Update(v, "Updated", "Updated Description", false)

		if task.Done {
			t.Error("Expected task to remain not done")
		}
	})
}

func TestValidateTask(t *testing.T) {
	tests := []struct {
		name  string
		task  *Task
		valid bool
	}{
		{
			name:  "valid task",
			task:  NewTask(1, "Valid Title", "Valid Description"),
			valid: true,
		},
		{
			name:  "zero ID task",
			task:  NewTask(0, "Valid Title", "Valid Description"),
			valid: false,
		},
		{
			name:  "negative ID task",
			task:  NewTask(-1, "Valid Title", "Valid Description"),
			valid: false,
		},
		{
			name:  "empty title task",
			task:  NewTask(1, "", "Valid Description"),
			valid: false,
		},
		{
			name:  "empty description task",
			task:  NewTask(1, "Valid Title", ""),
			valid: true,
		},
		{
			name:  "multiple errors task",
			task:  NewTask(-99, "", ""),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateTask(v, tt.task)
			if tt.valid && !v.Valid() {
				t.Errorf("Expected task to be valid, got errors: %v", v.Errors)
			}
			if !tt.valid && v.Valid() {
				t.Error("Expected task to be invalid")
			}
		})
	}
	t.Run("title exceeds 100 characters", func(t *testing.T) {
		longTitle := ""
		for i := 0; i < 101; i++ {
			longTitle += "a"
		}
		task := &Task{
			ID:          1,
			Title:       longTitle,
			Description: "Valid Description",
			Done:        false,
		}
		v := validator.New()

		ValidateTask(v, task)

		if v.Valid() {
			t.Error("Expected task with long title to be invalid")
		}
		if _, exists := v.Errors["title"]; !exists {
			t.Error("Expected 'title' error to exist")
		}
	})

	t.Run("accepts task with title exactly 100 characters", func(t *testing.T) {
		exactTitle := ""
		for i := 0; i < 100; i++ {
			exactTitle += "a"
		}
		task := &Task{
			ID:          1,
			Title:       exactTitle,
			Description: "Valid Description",
			Done:        false,
		}
		v := validator.New()

		ValidateTask(v, task)

		if !v.Valid() {
			t.Errorf("Expected task with 100-char title to be valid, got errors: %v", v.Errors)
		}
	})

	t.Run("rejects task with description exceeding 2000 characters", func(t *testing.T) {
		longDesc := ""
		for i := 0; i < 2001; i++ {
			longDesc += "a"
		}
		task := &Task{
			ID:          1,
			Title:       "Valid Title",
			Description: longDesc,
			Done:        false,
		}
		v := validator.New()

		ValidateTask(v, task)

		if v.Valid() {
			t.Error("Expected task with long description to be invalid")
		}
		if _, exists := v.Errors["description"]; !exists {
			t.Error("Expected 'description' error to exist")
		}
	})

	t.Run("accepts task with description exactly 2000 characters", func(t *testing.T) {
		exactDesc := ""
		for i := 0; i < 2000; i++ {
			exactDesc += "a"
		}
		task := &Task{
			ID:          1,
			Title:       "Valid Title",
			Description: exactDesc,
			Done:        false,
		}
		v := validator.New()

		ValidateTask(v, task)

		if !v.Valid() {
			t.Errorf("Expected task with 2000-char description to be valid, got errors: %v", v.Errors)
		}
	})
}
