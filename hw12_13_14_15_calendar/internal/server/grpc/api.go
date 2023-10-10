package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/app"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const UUIDHeader = "x-api-user"

type SearchPeriod int

const (
	Undefined SearchPeriod = iota
	Day
	Week
	Month
)

type Service struct {
	UnimplementedCalendarServer
	app    app.Application
	logger zap.Logger
}

func NewService(app app.Application, l zap.Logger) *Service {
	return &Service{app: app, logger: l}
}

func (s *Service) CreateEvent(ctx context.Context, r *EventRequest) (*Event, error) {
	uid, err := s.getUserFromMeta(ctx)
	if err != nil {
		return nil, err
	}

	event, err := unmarshalEvent(r.GetEvent(), uid)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	res, err := s.app.CreateEvent(uid, event)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return marshalEvent(res), nil
}

func (s *Service) UpdateEvent(ctx context.Context, r *EventRequest) (*Event, error) {
	uid, err := s.getUserFromMeta(ctx)
	if err != nil {
		return nil, err
	}

	event, err := unmarshalEvent(r.GetEvent(), uid)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	res, err := s.app.UpdateEvent(event.ID, uid, event)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrNotFound):
			return nil, status.Errorf(codes.NotFound, "%s", err)
		case errors.Is(err, app.ErrAccessDenied):
			return nil, status.Errorf(codes.PermissionDenied, "%s", err)
		default:
			s.logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, "%s", err)
		}
	}

	return marshalEvent(res), nil
}

func (s *Service) DeleteEvent(ctx context.Context, r *EventIdRequest) (*DeleteEventResponse, error) {
	uid, err := s.getUserFromMeta(ctx)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(r.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	err = s.app.DeleteEvent(eid, uid)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &DeleteEventResponse{}, nil
}

func (s *Service) GetEvent(ctx context.Context, r *EventIdRequest) (*Event, error) {
	uid, err := s.getUserFromMeta(ctx)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(r.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	event, err := s.app.GetEvent(eid, uid)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrNotFound):
			return nil, status.Errorf(codes.NotFound, "%s", err)
		case errors.Is(err, app.ErrAccessDenied):
			return nil, status.Errorf(codes.PermissionDenied, "%s", err)
		default:
			s.logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, "%s", err)
		}
	}

	return marshalEvent(event), nil
}

func (s *Service) getForPeriod(ctx context.Context, r *StartDateRequest, period SearchPeriod) (*Events, error) {
	uid, err := s.getUserFromMeta(ctx)
	if err != nil {
		return nil, err
	}

	start := r.GetStart().AsTime()

	var end time.Time
	switch period {
	case Day, Undefined:
		end = start.AddDate(0, 0, 1).Add(-1 * time.Nanosecond)
	case Week:
		end = start.AddDate(0, 0, 7).Add(-1 * time.Nanosecond)
	case Month:
		end = start.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	}

	events, err := s.app.GetEventsForPeriod(uid, start, end)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return marshalEvents(events), nil
}

func (s *Service) GetForDay(ctx context.Context, r *StartDateRequest) (*Events, error) {
	return s.getForPeriod(ctx, r, Day)
}

func (s *Service) GetForWeek(ctx context.Context, r *StartDateRequest) (*Events, error) {
	return s.getForPeriod(ctx, r, Week)
}

func (s *Service) GetForMonth(ctx context.Context, r *StartDateRequest) (*Events, error) {
	return s.getForPeriod(ctx, r, Month)
}

func (s *Service) getUserFromMeta(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Error(fmt.Sprintf("can't get request metadata from '%v'", md))
		return uuid.UUID{}, status.Errorf(codes.Internal, "can't get request metadata")
	}
	if len(md.Get(UUIDHeader)) == 0 {
		return uuid.UUID{}, status.Errorf(codes.Unauthenticated, `request hasn't header "%s"`, UUIDHeader)
	}
	uid, err := uuid.Parse(md.Get(UUIDHeader)[0])
	if err != nil {
		return uuid.UUID{}, status.Errorf(codes.Unauthenticated, `"%s" is not valid UUID: %s`, md.Get(UUIDHeader)[0], err)
	}

	return uid, nil
}

func marshalEvent(e storage.Event) *Event {
	return &Event{
		Id:           e.ID.String(),
		Title:        e.Title,
		Description:  e.Description,
		StartDate:    timestamppb.New(e.StartDate),
		EndDate:      timestamppb.New(e.EndDate),
		NotifyBefore: durationpb.New(e.NotifyBefore),
	}
}

func unmarshalEvent(e *Event, uid uuid.UUID) (storage.Event, error) {
	eid := uuid.UUID{}
	if e.Id != "" {
		id, err := uuid.Parse(e.Id)
		if err != nil {
			return storage.Event{}, status.Errorf(codes.InvalidArgument, "%s", err)
		}
		eid = id
	}
	return storage.Event{
		ID:           eid,
		Title:        e.Title,
		Description:  e.Description,
		StartDate:    e.StartDate.AsTime(),
		EndDate:      e.EndDate.AsTime(),
		NotifyBefore: e.NotifyBefore.AsDuration(),
		UserID:       uid,
	}, nil
}

func marshalEvents(events []storage.Event) *Events {
	count := len(events)
	res := Events{
		Events: make([]*Event, count),
	}
	for i := 0; i < count; i++ {
		res.Events[i] = marshalEvent(events[i])
	}
	return &res
}
