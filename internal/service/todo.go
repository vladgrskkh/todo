package service

import (
	"fmt"
	"log/slog"

	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/internal/repository"
	"github.com/vladgrskkh/todo/pkg/validator"
)

var (
	ErrInvalidID  = fmt.Errorf("invalid id param")
	ErrTaskExists = fmt.Errorf("task with this id already exists")
)

type TodoService struct {
	logger   *slog.Logger
	taskRepo *repository.TaskRepo
}

func NewTodoService(logger *slog.Logger, taskRepo *repository.TaskRepo) *TodoService {
	return &TodoService{
		logger:   logger,
		taskRepo: taskRepo,
	}
}

func (s *TodoService) GetTask(id int64) (*domain.Task, error) {
	if id < 1 {
		return nil, ErrInvalidID
	}

	task, err := s.taskRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error getting task with %d id: %w", id, err)
	}

	return task, nil
}

func (s *TodoService) GetAllTasks() []*domain.Task {
	return s.taskRepo.GetAll()
}

func (s *TodoService) CreateTask(task *domain.Task) error {
	validator := validator.New()

	domain.ValidateTask(validator, task)

	if !validator.Valid() {
		return validator
	}

	// checking if task with this id already exists
	_, err := s.taskRepo.Get(task.ID)
	if err == nil {
		return ErrTaskExists
	}

	s.taskRepo.Insert(task)
	return nil
}

func (s *TodoService) UpdateTask(task *domain.Task) error {
	validator := validator.New()

	domain.ValidateTask(validator, task)

	if !validator.Valid() {
		return validator
	}

	// checking if task with this id already exists
	t, err := s.taskRepo.Get(task.ID)
	if err != nil {
		return err
	}

	if t.Done {
		validator.AddError("done", "cannot change done task")
		return validator
	}

	err = s.taskRepo.Update(task)
	if err != nil {
		return fmt.Errorf("error updating task with %d id: %w", task.ID, err)
	}

	return nil
}

func (s *TodoService) DeleteTask(id int64) error {
	if id < 1 {
		return ErrInvalidID
	}

	err := s.taskRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting task with %d id: %w", id, err)
	}

	return nil
}
