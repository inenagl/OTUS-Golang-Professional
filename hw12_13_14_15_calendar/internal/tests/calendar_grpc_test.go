//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	grpcapi "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CalendarGRPCSuite struct {
	suite.Suite
	authorizedCtx   context.Context
	unauthorizedCtx context.Context
	userID          string
	client          grpcapi.CalendarClient
	conn            *grpc.ClientConn
}

func (s *CalendarGRPCSuite) SetupSuite() {
	grpcAddr := os.Getenv("GOCLNDR_GRPCADDR")
	if grpcAddr == "" {
		grpcAddr = ":8889"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.client = grpcapi.NewCalendarClient(conn)
	s.conn = conn

	s.unauthorizedCtx = context.Background()
}

func (s *CalendarGRPCSuite) TeardownSuite() {
	err := s.conn.Close()
	s.Require().NoError(err)
}

func (s *CalendarGRPCSuite) SetupTest() {
	s.userID = uuid.New().String()
	md := metadata.New(nil)
	md.Append(grpcapi.UUIDHeader, s.userID)
	s.authorizedCtx = metadata.NewOutgoingContext(s.unauthorizedCtx, md)
}

func (s *CalendarGRPCSuite) TeardownTest() {
	s.userID = ""
	s.authorizedCtx = s.unauthorizedCtx
}

func (s *CalendarGRPCSuite) TestUnauthorizedRequest() {
	event := randomGRPCEvent()

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.CreateEvent(s.unauthorizedCtx, &grpcapi.EventRequest{Event: &event})
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.GetEvent(s.unauthorizedCtx, &grpcapi.EventIdRequest{Id: uuid.New().String()})
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.UpdateEvent(s.unauthorizedCtx, &grpcapi.EventRequest{Event: &event})
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.DeleteEvent(s.unauthorizedCtx, &grpcapi.EventIdRequest{Id: uuid.New().String()})
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.GetForDay(
				s.unauthorizedCtx,
				&grpcapi.StartDateRequest{Start: timestamppb.New(randomDay())},
			)
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.GetForWeek(
				s.unauthorizedCtx,
				&grpcapi.StartDateRequest{Start: timestamppb.New(randomDay())},
			)
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)

	s.T().Run(
		"create", func(t *testing.T) {
			_, err := s.client.GetForMonth(
				s.unauthorizedCtx,
				&grpcapi.StartDateRequest{Start: timestamppb.New(randomDay())},
			)
			s.Require().Error(err)
			s.Require().ErrorContains(err, codes.Unauthenticated.String())
		},
	)
}

func (s *CalendarGRPCSuite) TestCreateEvent() {
	event := randomGRPCEvent()

	res, err := s.client.CreateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &event})
	s.Require().NoError(err)
	eventID := res.Id

	// Проверяем ответ метода создания.
	expected := mapFromGRPCEvent(&event)
	actual := mapFromGRPCEvent(res)
	s.Require().Equal(expected, actual)

	// Проверяем, что будут возвращаться правильные данные.
	res, err = s.client.GetEvent(s.authorizedCtx, &grpcapi.EventIdRequest{Id: eventID})
	s.Require().NoError(err)
	expected["Id"] = eventID
	actual = mapFromGRPCEvent(res)
	actual["Id"] = res.Id
	s.Require().Equal(expected, actual)
}

func (s *CalendarGRPCSuite) TestUpdateEvent() {
	event := randomGRPCEvent()

	res, err := s.client.CreateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &event})
	s.Require().NoError(err)
	eventID := res.Id

	// Обновим созданное новыми данными.
	event = randomGRPCEvent()
	event.Id = eventID
	res, err = s.client.UpdateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &event})
	s.Require().NoError(err)

	// Проверяем ответ метода обновления.
	expected := mapFromGRPCEvent(&event)
	actual := mapFromGRPCEvent(res)
	s.Require().Equal(expected, actual)

	// Проверяем, что возвращаются обновленные данные.
	res, err = s.client.GetEvent(s.authorizedCtx, &grpcapi.EventIdRequest{Id: eventID})
	s.Require().NoError(err)
	expected["Id"] = eventID
	actual = mapFromGRPCEvent(res)
	actual["Id"] = res.Id
	s.Require().Equal(expected, actual)
}

func (s *CalendarGRPCSuite) TestDeleteEvent() {
	event := randomGRPCEvent()

	res, err := s.client.CreateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &event})
	s.Require().NoError(err)
	eventID := res.Id

	resp, err := s.client.DeleteEvent(s.authorizedCtx, &grpcapi.EventIdRequest{Id: eventID})
	s.Require().NoError(err)
	expected := grpcapi.DeleteEventResponse{}
	s.Require().IsType(&expected, resp)

	// Проверяем, что после удаления ничего не возвращается.
	_, err = s.client.GetEvent(s.authorizedCtx, &grpcapi.EventIdRequest{Id: eventID})
	s.Require().Error(err)
	s.Require().ErrorContains(err, codes.NotFound.String())
}

func (s *CalendarGRPCSuite) TestGetForDay() {
	day := randomDay()
	var anotherDay time.Time
	for anotherDay = randomDay(); anotherDay.Unix() == day.Unix(); {
		continue
	}
	events := make([]*grpcapi.Event, 5)
	var ev grpcapi.Event
	for i := 0; i < 5; i++ {
		ev = randomGRPCEvent()
		// первые три в текущем дне, остальные - в другом дне
		if i < 3 {
			ev.StartDate = timestamppb.New(day.Add(time.Duration(i) * time.Hour))
			ev.EndDate = timestamppb.New(day.Add(time.Duration(i+1) * time.Hour))
		} else {
			ev.StartDate = timestamppb.New(anotherDay.Add(time.Duration(i) * time.Hour))
			ev.EndDate = timestamppb.New(anotherDay.Add(time.Duration(i+1) * time.Hour))
		}
		res, err := s.client.CreateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &ev})
		s.Require().NoError(err)
		events[i] = res
	}

	res, err := s.client.GetForDay(s.authorizedCtx, &grpcapi.StartDateRequest{Start: timestamppb.New(day)})
	s.Require().NoError(err)

	expected := []map[string]interface{}{
		mapFromGRPCEvent(events[0]),
		mapFromGRPCEvent(events[1]),
		mapFromGRPCEvent(events[2]),
	}
	actual := make([]map[string]interface{}, len(res.GetEvents()))
	for i, ev := range res.GetEvents() {
		actual[i] = mapFromGRPCEvent(ev)
	}
	s.Require().EqualValues(expected, actual)
}

func (s *CalendarGRPCSuite) TestGetForWeek() {
	firstDay := randomDay()
	// Делаем 10 подряд идущих дней.
	days := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		days[i] = firstDay.AddDate(0, 0, i)
	}

	events := make([]*grpcapi.Event, 10)
	var ev grpcapi.Event
	var start time.Time
	// Делаем по событию в каждом дне.
	for i := 0; i < 10; i++ {
		ev = randomGRPCEvent()
		start = days[i].Add(randomDuration())
		ev.StartDate = timestamppb.New(start)
		ev.EndDate = timestamppb.New(start.Add(time.Hour))

		res, err := s.client.CreateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &ev})
		s.Require().NoError(err)
		events[i] = res
	}

	res, err := s.client.GetForWeek(s.authorizedCtx, &grpcapi.StartDateRequest{Start: timestamppb.New(firstDay)})
	s.Require().NoError(err)

	// Ожидаем в ответе первые 7 событий (в течение недели).
	expected := make([]map[string]interface{}, 7)
	for i := 0; i < 7; i++ {
		expected[i] = mapFromGRPCEvent(events[i])
	}
	actual := make([]map[string]interface{}, len(res.GetEvents()))
	for i, ev := range res.GetEvents() {
		actual[i] = mapFromGRPCEvent(ev)
	}
	s.Require().EqualValues(expected, actual)
}

func (s *CalendarGRPCSuite) TestGetForMonth() {
	firstDay := randomDay()
	days := []time.Time{
		firstDay,                   // 0: ok first day of month
		firstDay.AddDate(0, 0, 1),  // 1: ok
		firstDay.AddDate(0, 2, 5),  // 2: fail
		firstDay.AddDate(0, 0, 5),  // 3: ok
		firstDay.AddDate(0, 1, 1),  // 4: fail
		firstDay.AddDate(0, 1, 11), // 5: fail
		firstDay.AddDate(0, 0, 15), // 6: ok
		firstDay.AddDate(0, 1, 31), // 7: fail
		firstDay.AddDate(0, 0, 25), // 8: ok
		firstDay.AddDate(0, 1, -1), // 9: ok day before next month
	}

	events := make([]*grpcapi.Event, 10)
	var ev grpcapi.Event
	var start time.Time
	// Делаем по событию в каждом дне.
	for i := 0; i < 10; i++ {
		ev = randomGRPCEvent()
		start = days[i].Add(randomDuration())
		ev.StartDate = timestamppb.New(start)
		ev.EndDate = timestamppb.New(start.Add(time.Hour))

		res, err := s.client.CreateEvent(s.authorizedCtx, &grpcapi.EventRequest{Event: &ev})
		s.Require().NoError(err)
		events[i] = res
	}

	res, err := s.client.GetForMonth(s.authorizedCtx, &grpcapi.StartDateRequest{Start: timestamppb.New(firstDay)})
	s.Require().NoError(err)

	expected := []map[string]interface{}{
		mapFromGRPCEvent(events[0]),
		mapFromGRPCEvent(events[1]),
		mapFromGRPCEvent(events[3]),
		mapFromGRPCEvent(events[6]),
		mapFromGRPCEvent(events[8]),
		mapFromGRPCEvent(events[9]),
	}
	actual := make([]map[string]interface{}, len(res.GetEvents()))
	for i, ev := range res.GetEvents() {
		actual[i] = mapFromGRPCEvent(ev)
	}
	s.Require().EqualValues(expected, actual)
}

func TestCalendarGRPCSuite(t *testing.T) {
	suite.Run(t, new(CalendarGRPCSuite))
}
