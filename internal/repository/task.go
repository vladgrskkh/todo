package repository

import (
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/vladgrskkh/todo/internal/domain"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrEditConflict = errors.New("edit conflict")
)

type TaskRepo struct {
	data  map[int64]*domain.Task
	mutex sync.RWMutex
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
		return nil, ErrNotFound
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
	r.data[task.ID] = task
	r.mutex.Unlock()
}

func (r *TaskRepo) Update(task *domain.Task) error {
	r.mutex.Lock()
	// retriving task here to prevent data race(optimistic locking)
	t, ok := r.data[task.ID]
	if !ok {
		return ErrNotFound
	}

	if t.Version != task.Version {
		r.mutex.Unlock()
		return ErrEditConflict
	}

	task.Version++
	r.data[task.ID] = task
	r.mutex.Unlock()

	return nil
}

func (r *TaskRepo) Delete(id int64) error {
	r.mutex.Lock()
	before := len(r.data)
	delete(r.data, id)

	if len(r.data) == before {
		r.mutex.Unlock()
		return ErrNotFound
	}
	r.mutex.Unlock()

	return nil
}
