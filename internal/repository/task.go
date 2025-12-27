package repository

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"

	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/pkg/inmemorydb"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrEditConflict  = errors.New("edit conflict")
	ErrAlreadyExists = errors.New("resource already exists")
)

type TaskRepo struct {
	db *inmemorydb.DB
}

func NewTaskRepo(db *inmemorydb.DB) *TaskRepo {
	return &TaskRepo{
		db: db,
	}
}

func (r *TaskRepo) Get(id int64) (*domain.Task, error) {
	key := strconv.FormatInt(id, 10)
	obj, err := r.db.GetObject(key)
	if err != nil {
		switch {
		case errors.Is(err, inmemorydb.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	dec := gob.NewDecoder(bytes.NewReader(obj))

	var task domain.Task
	err = dec.Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepo) GetAll() ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0, r.db.Size())
	data := r.db.GetAllObjects()

	for _, v := range data {
		dec := gob.NewDecoder(bytes.NewReader(v))

		var task domain.Task
		err := dec.Decode(&task)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *TaskRepo) Insert(task *domain.Task) error {
	key := strconv.FormatInt(task.ID, 10)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(task)
	if err != nil {
		return err
	}

	err = r.db.PutObject(key, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (r *TaskRepo) Update(task *domain.Task) error {
	key := strconv.FormatInt(task.ID, 10)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(task)
	if err != nil {
		return err
	}

	err = r.db.PutObject(key, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (r *TaskRepo) Delete(id int64) error {
	key := strconv.FormatInt(id, 10)
	err := r.db.DeleteObject(key)
	if err != nil {
		switch {
		case errors.Is(err, inmemorydb.ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}
