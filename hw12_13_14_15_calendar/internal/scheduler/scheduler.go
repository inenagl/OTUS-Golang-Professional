package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/json"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/queue"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"go.uber.org/zap"
)

var marshalledFields = []json.EventField{
	json.EventID,
	json.EventTitle,
	json.EventStartDate,
	json.EventEndDate,
	json.EventDescription,
	json.EventNotifyBefore,
}

type Scheduler interface {
	Start(ctx context.Context) error
	Stop() error
}

type S struct {
	workCycle  time.Duration
	expiration time.Duration
	logger     zap.Logger
	storage    Storage
	producer   queue.Producer
	cancel     context.CancelFunc
}

type Storage interface {
	SetEventsNotified(ids []uuid.UUID, notified time.Time) error
	NotificationNeededEvents(t time.Time) ([]storage.Event, error)
	DeleteEvents(filter []storage.EventCondition) error
}

func New(
	workCycle time.Duration,
	expiration time.Duration,
	logger zap.Logger,
	storage Storage,
	producer queue.Producer,
) *S {
	return &S{
		workCycle:  workCycle,
		expiration: expiration,
		logger:     logger,
		storage:    storage,
		producer:   producer,
	}
}

func (s *S) Start(ctx context.Context) error {
	s.logger.Debug("trying connect to AMPQ...")
	if err := s.producer.Connect(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	wg := sync.WaitGroup{}
	wg.Add(2)
	s.logger.Debug("starting notification goroutine...")
	go func(ctx context.Context) {
		defer wg.Done()
		t := time.NewTimer(time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				s.logger.Debug("done in notifying")
				return
			case <-t.C:
				// default:
				err := s.notify()
				if err != nil {
					err = fmt.Errorf("notify: %w", err)
					s.logger.Error(err.Error())
					return
				}
				t.Reset(s.workCycle)
			}
		}
	}(ctx)

	s.logger.Debug("starting deletion goroutine...")
	go func(ctx context.Context) {
		defer wg.Done()
		t := time.NewTimer(time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				s.logger.Debug("done in deleting")
				return
			case <-t.C:
				err := s.deleteOldEvents()
				if err != nil {
					err = fmt.Errorf("delete events: %w", err)
					s.logger.Error(err.Error())
					return
				}
				t.Reset(s.workCycle)
			}
		}
	}(ctx)

	wg.Wait()

	return nil
}

func (s *S) Stop() error {
	s.logger.Debug("cancel in stop")
	s.cancel()
	err := s.producer.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *S) notify() error {
	s.logger.Debug("notifying...")

	t := time.Now()
	events, err := s.storage.NotificationNeededEvents(t)
	if err != nil {
		return err
	}
	s.logger.Debug(strconv.Itoa(len(events)) + " events")

	errs := make([]string, 0)
	for _, event := range events {
		err = s.producer.Publish(json.MarshallEvent(event, marshalledFields))
		if err != nil {
			errs = append(errs, fmt.Sprintf("event %s: %s", event.ID.String(), err.Error()))
			continue
		}
		s.logger.Debug("published " + event.ID.String())
		err = s.storage.SetEventsNotified([]uuid.UUID{event.ID}, t)
		if err != nil {
			errs = append(errs, fmt.Sprintf("event %s: %s", event.ID.String(), err.Error()))
			continue
		}
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (s *S) deleteOldEvents() error {
	s.logger.Debug("deleting...")
	filter := []storage.EventCondition{
		{Field: storage.EventEndDate, Type: storage.TypeLess, Sample: time.Now().Add(-1 * s.expiration)},
	}

	return s.storage.DeleteEvents(filter)
}
