package domain

import (
	"unicode/utf8"

	"github.com/vladgrskkh/todo/pkg/validator"
)

// TODO: check for struct alligment
type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

func NewTask(id int64, title string, description string, done bool) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        done,
	}
}

func ValidateTask(v *validator.Validator, task *Task) {
	v.Check(task.ID > 0, "id", "must be a positive integer")

	v.Check(task.Title != "", "title", "must be provided")
	v.Check(utf8.RuneCountInString(task.Title) <= 100, "title", "must not be more than 100 symbols long")

	v.Check(utf8.RuneCountInString(task.Description) <= 2000, "description", "must not be more than 2000 symbols long")
}
