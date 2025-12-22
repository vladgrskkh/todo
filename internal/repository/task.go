package repository

import (
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/vladgrskkh/todo/internal/domain"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepo struct {
	data   map[int64]*domain.Task
	nextID int64
	mutex  sync.RWMutex
}

func NewTaskRepo() *TaskRepo {
	return &TaskRepo{
		data:  make(map[int64]*domain.Task, 0),
		mutex: sync.RWMutex{},
	}
}

func (r *TaskRepo) Get(id int64) (*domain.Task, error) {
	r.mutex.RLock()
	task, ok := r.data[id]
	if !ok {
		r.mutex.RUnlock()
		return nil, ErrTaskNotFound
	}

	r.mutex.RUnlock()

	return task, nil
}

func (r *TaskRepo) GetAll() []*domain.Task {
	r.mutex.RLock()
	tasks := slices.Collect(maps.Values(r.data))
	r.mutex.RUnlock()

	return tasks
}

func (r *TaskRepo) Insert(task *domain.Task) {
	r.mutex.Lock()
	task.ID = r.nextID
	r.data[task.ID] = task
	r.nextID++
	r.mutex.Unlock()
}

// TODO: race condition?(mb implement vesioning)
func (r *TaskRepo) Update(task *domain.Task) error {
	r.mutex.RLock()
	_, ok := r.data[task.ID]
	if !ok {
		r.mutex.RUnlock()
		return ErrTaskNotFound
	}
	r.mutex.RUnlock()

	r.mutex.Lock()
	r.data[task.ID] = task
	r.mutex.Unlock()

	return nil
}

func (r *TaskRepo) Delete(id int64) error {
	r.mutex.RLock()
	// mb just delete and then compare len before and after?
	_, ok := r.data[id]
	if !ok {
		r.mutex.RUnlock()
		return ErrTaskNotFound
	}
	r.mutex.RUnlock()

	r.mutex.Lock()
	delete(r.data, id)
	r.mutex.Unlock()

	return nil
}
