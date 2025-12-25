package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/vladgrskkh/todo/internal/apierrors"
	"github.com/vladgrskkh/todo/internal/domain"
	"github.com/vladgrskkh/todo/internal/paramutil"
	"github.com/vladgrskkh/todo/internal/repository"
	s "github.com/vladgrskkh/todo/internal/service"
	"github.com/vladgrskkh/todo/pkg/jsonhttp"
	"github.com/vladgrskkh/todo/pkg/validator"
)

type TaskGetter interface {
	GetTask(id int64) (*domain.Task, error)
	GetAllTasks() []*domain.Task
}

func NewGetTaskHandler(logger *slog.Logger, service TaskGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := paramutil.ReadIDParam(r)
		if err != nil {
			apierrors.BadRequestResponse(logger, w, r, err)
			return
		}

		task, err := service.GetTask(id)
		if err != nil {
			switch {
			case errors.Is(err, s.ErrInvalidID):
				apierrors.BadRequestResponse(logger, w, r, err)
			case errors.Is(err, repository.ErrNotFound):
				apierrors.NotFoundResponse(logger, w, r)
			default:
				apierrors.ServerErrorResponse(logger, w, r, err)
			}

			return
		}

		err = jsonhttp.WriteJSON(w, http.StatusOK, jsonhttp.Envelope{"task": task}, nil)
		if err != nil {
			apierrors.ServerErrorResponse(logger, w, r, err)
		}
	}
}

func NewGetAllTasksHandler(logger *slog.Logger, service TaskGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks := service.GetAllTasks()

		err := jsonhttp.WriteJSON(w, http.StatusOK, jsonhttp.Envelope{"tasks": tasks}, nil)
		if err != nil {
			apierrors.ServerErrorResponse(logger, w, r, err)
		}
	}
}

type TaskCreater interface {
	CreateTask(task *domain.Task) error
}

func NewPostTaskHandler(logger *slog.Logger, service TaskCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		err := jsonhttp.ReadJSON(w, r, &input)
		if err != nil {
			apierrors.BadRequestResponse(logger, w, r, err)
			return
		}

		task := domain.NewTask(input.ID, input.Title, input.Description)

		err = service.CreateTask(task)
		if err != nil {
			var validationErr *validator.Validator
			switch {
			case errors.As(err, &validationErr):
				apierrors.FailedValidationResponse(logger, w, r, validationErr.Errors)
			default:
				apierrors.ServerErrorResponse(logger, w, r, err)
			}

			return
		}

		err = jsonhttp.WriteJSON(w, http.StatusCreated, jsonhttp.Envelope{"task": task}, nil)
		if err != nil {
			apierrors.ServerErrorResponse(logger, w, r, err)
		}
	}
}

type TaskUpdater interface {
	GetTask(id int64) (*domain.Task, error)
	UpdateTask(task *domain.Task) error
}

func NewTaskUpdater(logger *slog.Logger, service TaskUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := paramutil.ReadIDParam(r)
		if err != nil {
			apierrors.BadRequestResponse(logger, w, r, err)
			return
		}

		// mb this should be done in service layer
		// TODO: duplicates in service layer(need to decide what to do)
		task, err := service.GetTask(id)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				apierrors.NotFoundResponse(logger, w, r)
			default:
				apierrors.ServerErrorResponse(logger, w, r, err)
			}

			return
		}

		var input struct {
			Title       *string `json:"title"`
			Description *string `json:"description"`
			Done        *bool   `json:"done"`
		}

		err = jsonhttp.ReadJSON(w, r, &input)
		if err != nil {
			apierrors.BadRequestResponse(logger, w, r, err)
			return
		}

		if input.Title != nil {
			task.Title = *input.Title
		}

		if input.Description != nil {
			task.Description = *input.Description
		}

		if input.Done != nil {
			task.Done = *input.Done
		}

		err = service.UpdateTask(task)
		if err != nil {
			var validationErr *validator.Validator
			switch {
			case errors.As(err, &validationErr):
				apierrors.FailedValidationResponse(logger, w, r, validationErr.Errors)
			// TODO: handle this
			case errors.Is(err, repository.ErrNotFound):
				apierrors.NotFoundResponse(logger, w, r)
			default:
				apierrors.ServerErrorResponse(logger, w, r, err)
			}

			return
		}

		err = jsonhttp.WriteJSON(w, http.StatusOK, jsonhttp.Envelope{"task": task}, nil)
		if err != nil {
			apierrors.ServerErrorResponse(logger, w, r, err)
		}
	}
}

type TaskDeleter interface {
	DeleteTask(id int64) error
}

func NewDeleteTaskHandler(logger *slog.Logger, service TaskDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := paramutil.ReadIDParam(r)
		if err != nil {
			apierrors.BadRequestResponse(logger, w, r, err)
			return
		}

		err = service.DeleteTask(id)
		if err != nil {
			switch {
			case errors.Is(err, s.ErrInvalidID):
				apierrors.BadRequestResponse(logger, w, r, err)
			case errors.Is(err, repository.ErrNotFound):
				apierrors.NotFoundResponse(logger, w, r)
			default:
				apierrors.ServerErrorResponse(logger, w, r, err)
			}

			return
		}

		err = jsonhttp.WriteJSON(w, http.StatusOK, jsonhttp.Envelope{"message": "task successfully deleted"}, nil)
		if err != nil {
			apierrors.ServerErrorResponse(logger, w, r, err)
		}
	}
}
