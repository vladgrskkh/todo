package domain

// TODO: check for struct alligment
type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done,omitempty"`
}

func NewTask(id int64, title string, description string, done bool) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        done,
	}
}
