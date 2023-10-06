package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"go.uber.org/zap"
)

type App struct {
	logger  zap.Logger
	storage Storage
}

type Storage interface {
	AddEvent(event storage.Event) (uuid.UUID, error)
	UpdateEvent(event storage.Event) error
	DeleteEvent(id uuid.UUID) error
	GetEvent(id uuid.UUID) (storage.Event, error)
	GetEvents(filter []storage.EventCondition, sort []storage.EventSort) ([]storage.Event, error)
}

func New(logger zap.Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error { //nolint: revive
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

// TODO
