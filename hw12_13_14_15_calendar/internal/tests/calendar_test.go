package integration_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	grpcapi "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/jackc/fake"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CalendarSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *CalendarSuite) SetupSuite() {

}

func (s *CalendarSuite) TeardownSuite() {

}

func (s *CalendarSuite) SetupTest() {

}

func (s *CalendarSuite) TeardownTest() {

}

func (s *CalendarSuite) TestAddEvent() {
	// os.Getenv("GOCLNDR_RESTADDR")
	grpcAddr := os.Getenv("GOCLNDR_GRPCADDR")
	if grpcAddr == "" {
		grpcAddr = ":8889"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	s.ctx = context.Background()
	client := grpcapi.NewCalendarClient(conn)
	start := time.Date(
		fake.Year(2023, 2025),
		time.Month((rand.Intn(12) + 1)),
		fake.Day(),
		rand.Intn(24),
		rand.Intn(59),
		rand.Intn(59),
		0,
		time.Local,
	)
	event := grpcapi.Event{
		Title:        fake.Sentence(),
		Description:  fake.Paragraph(),
		StartDate:    timestamppb.New(start),
		EndDate:      timestamppb.New(start.Add(time.Minute * time.Duration(rand.Intn(10000)))),
		NotifyBefore: durationpb.New(time.Minute * time.Duration(rand.Intn(60))),
	}
	userId := uuid.New().String()
	md := metadata.New(nil)
	md.Append(grpcapi.UUIDHeader, userId)
	ctx := metadata.NewOutgoingContext(s.ctx, md)
	res, err := client.CreateEvent(ctx, &grpcapi.EventRequest{Event: &event})
	s.Require().NoError(err)

	expected := map[string]interface{}{
		"Title":        event.Title,
		"Description":  event.Description,
		"StartDate":    event.StartDate.AsTime(),
		"EndDate":      event.EndDate.AsTime(),
		"NotifyBefore": event.NotifyBefore.AsDuration(),
	}
	actual := map[string]interface{}{
		"Title":        res.Title,
		"Description":  res.Description,
		"StartDate":    res.StartDate.AsTime(),
		"EndDate":      res.EndDate.AsTime(),
		"NotifyBefore": res.NotifyBefore.AsDuration(),
	}
	s.Require().Equal(expected, actual)

	eventId := res.Id

	res, err = client.GetEvent(ctx, &grpcapi.EventIdRequest{Id: eventId})
	s.Require().NoError(err)
	expected["Id"] = eventId
	actual = map[string]interface{}{
		"Title":        res.Title,
		"Description":  res.Description,
		"StartDate":    res.StartDate.AsTime(),
		"EndDate":      res.EndDate.AsTime(),
		"NotifyBefore": res.NotifyBefore.AsDuration(),
	}
	actual["Id"] = eventId
	s.Require().Equal(expected, actual)
}

func TestCalendarSuite(t *testing.T) {
	suite.Run(t, new(CalendarSuite))
}
