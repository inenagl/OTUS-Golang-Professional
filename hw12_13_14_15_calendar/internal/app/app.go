package app

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"go.uber.org/zap"
)

type Application interface {
	GetEvent(id uuid.UUID, userID uuid.UUID) (storage.Event, error)
	UpdateEvent(id uuid.UUID, userID uuid.UUID, event storage.Event) (storage.Event, error)
	CreateEvent(userID uuid.UUID, event storage.Event) (storage.Event, error)
	DeleteEvent(id uuid.UUID, userID uuid.UUID) error
	GetEventsForPeriod(userID uuid.UUID, start, end time.Time) ([]storage.Event, error)
}

type App struct {
	logger  zap.Logger
	storage Storage
}

type Storage interface {
	AddEvent(event storage.Event) (storage.Event, error)
	UpdateEvent(event storage.Event) error
	DeleteEvent(id uuid.UUID) error
	GetEvent(id uuid.UUID) (storage.Event, error)
	GetEvents(filter []storage.EventCondition, sort []storage.EventSort) ([]storage.Event, error)
}

var (
	ErrNotFound     = errors.New("not found")
	ErrAccessDenied = errors.New("access denied")
)

func New(logger zap.Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(userID uuid.UUID, event storage.Event) (storage.Event, error) {
	event.UserID = userID
	return a.storage.AddEvent(event)
}

func (a *App) GetEvent(id uuid.UUID, userID uuid.UUID) (storage.Event, error) {
	event, err := a.storage.GetEvent(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, storage.ErrEventNotFound) {
			err = ErrNotFound
		}
		return storage.Event{}, err
	}
	if event.UserID != userID {
		return storage.Event{}, ErrAccessDenied
	}

	return event, nil
}

func (a *App) UpdateEvent(id uuid.UUID, userID uuid.UUID, event storage.Event) (storage.Event, error) {
	// проверка на наличие в хранилище и на принадлежность пользователю
	_, err := a.GetEvent(id, userID)
	if err != nil {
		return storage.Event{}, err
	}
	// допускаем, что пришлют объект с другим ID и userID, и чтобы не проапдейтить не то событие
	// затрём эти айдишки целевыми значениями
	event.ID = id
	event.UserID = userID

	err = a.storage.UpdateEvent(event)
	if err != nil {
		return storage.Event{}, err
	}

	return event, nil
}

func (a *App) DeleteEvent(id uuid.UUID, userID uuid.UUID) error {
	event, err := a.storage.GetEvent(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, storage.ErrEventNotFound) {
			err = ErrNotFound
		}
		return err
	}
	if event.UserID != userID {
		return ErrAccessDenied
	}

	return a.storage.DeleteEvent(id)
}

func (a *App) GetEventsForPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]storage.Event, error) {
	filter := []storage.EventCondition{
		{Field: storage.EventUserID, Type: storage.TypeEq, Sample: userID},
		{Field: storage.EventStartDate, Type: storage.TypeMoreOrEq, Sample: startDate},
		{Field: storage.EventStartDate, Type: storage.TypeLessOrEq, Sample: endDate},
	}
	sort := []storage.EventSort{
		{Field: storage.EventStartDate, Direction: storage.DirectionAsc},
		{Field: storage.EventEndDate, Direction: storage.DirectionAsc},
	}

	return a.storage.GetEvents(filter, sort)
}
