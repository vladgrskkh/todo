package mocks

import (
	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/internal/handlers/dto"
)

type mockTaskGetter struct {
	task      *domain.Task
	tasks     []*domain.Task
	getErr    error
	getAllErr error
}

func NewMockTaskGetter(task *domain.Task, tasks []*domain.Task, getErr, getAllErr error) *mockTaskGetter {
	return &mockTaskGetter{task, tasks, getErr, getAllErr}
}

func (m *mockTaskGetter) GetTask(id int64) (*domain.Task, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.task, nil
}

func (m *mockTaskGetter) GetAllTasks() ([]*domain.Task, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.tasks, nil
}

type mockTaskCreater struct {
	createErr error
}

func NewMockTaskCreator(createErr error) *mockTaskCreater {
	return &mockTaskCreater{createErr}
}

func (m *mockTaskCreater) CreateTask(task *domain.Task) error {
	return m.createErr
}

type mockTaskUpdater struct {
	task      *domain.Task
	updateErr error
}

func NewMockTaskUpdater(task *domain.Task, updateErr error) *mockTaskUpdater {
	return &mockTaskUpdater{task, updateErr}
}

func (m *mockTaskUpdater) UpdateTask(id int64, input dto.UpdateTaskInput) (*domain.Task, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}

	m.task.Title = input.Title
	m.task.Description = input.Description
	m.task.Done = input.Done
	return m.task, nil
}

type mockTaskDeleter struct {
	deleteErr error
}

func NewMockTaskDeleter(deleteErr error) *mockTaskDeleter {
	return &mockTaskDeleter{deleteErr}
}

func (m *mockTaskDeleter) DeleteTask(id int64) error {
	return m.deleteErr
}
