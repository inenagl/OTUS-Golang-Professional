package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/app"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/logger"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/jackc/fake"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testHost = "localhost"
	testPort = 8082
)

var (
	testClient  CalendarClient
	testStorage app.Storage
	testApp     app.Application
)

func TestMain(m *testing.M) {
	logg, err := logger.New("prod", "fatal", "json", []string{"stdout"}, []string{"stdout"})
	if err != nil {
		os.Exit(1)
	}

	testStorage = memorystorage.New()
	testApp = app.New(*logg, testStorage)
	server := NewServer(testHost, testPort, *logg, testApp)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		server.Stop(ctx)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		server.Start(ctx)
	}()

	if err != nil {
		log.Fatal(err)
	}

	tm := time.NewTimer(3 * time.Second)
	var conn *grpc.ClientConn
	for conn, err = grpc.Dial(
		fmt.Sprintf("%s:%d", testHost, testPort), grpc.WithTransportCredentials(insecure.NewCredentials()),
	); err != nil; {
		select {
		case <-tm.C:
			break
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if err != nil {
		os.Exit(2)
	}

	testClient = NewCalendarClient(conn)

	m.Run()

	cancel()
	wg.Wait()
}

func requestContext(userID uuid.UUID) context.Context {
	md := metadata.New(nil)
	md.Append(UUIDHeader, userID.String())
	return metadata.NewOutgoingContext(context.Background(), md)
}

func errCode(t *testing.T, err error) codes.Code {
	t.Helper()
	st, ok := status.FromError(err)
	if !ok {
		t.Error("can't process response code")
	}
	return st.Code()
}

func TestGetEvent(t *testing.T) {
	userID := uuid.New()
	event := storage.Event{
		ID:           uuid.UUID{},
		Title:        fake.Sentence(),
		StartDate:    time.Date(2023, time.January, 10, 10, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2023, time.January, 10, 11, 0, 0, 0, time.UTC),
		Description:  fake.Paragraph(),
		UserID:       userID,
		NotifyBefore: 24 * time.Hour,
	}
	event, _ = testStorage.AddEvent(event)

	req := EventIdRequest{Id: event.ID.String()}
	res, err := testClient.GetEvent(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))
	result, _ := unmarshalEvent(res, userID)
	require.Equal(t, event, result)
}

func TestUpdateEvent(t *testing.T) {
	userID := uuid.New()
	event := storage.Event{
		ID:           uuid.UUID{},
		Title:        fake.Sentence(),
		StartDate:    time.Date(2023, time.January, 10, 10, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2023, time.January, 10, 11, 0, 0, 0, time.UTC),
		Description:  fake.Paragraph(),
		UserID:       userID,
		NotifyBefore: 24 * time.Hour,
	}
	event, _ = testStorage.AddEvent(event)

	updatedEvent := event
	updatedEvent.Title = fake.Sentence()
	updatedEvent.StartDate = time.Date(2023, time.February, 21, 12, 0, 0, 0, time.UTC)
	updatedEvent.EndDate = time.Date(2023, time.February, 21, 12, 30, 0, 0, time.UTC)
	updatedEvent.Description = fake.Paragraph()
	updatedEvent.NotifyBefore = 15 * time.Minute

	req := EventRequest{Event: marshalEvent(updatedEvent)}
	res, err := testClient.UpdateEvent(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))
	result, _ := unmarshalEvent(res, userID)
	require.Equal(t, updatedEvent, result)

	storedEvent, _ := testStorage.GetEvent(updatedEvent.ID)
	require.Equal(t, storedEvent, result)
}

func TestCreateEvent(t *testing.T) {
	userID := uuid.New()
	event := storage.Event{
		ID:           uuid.UUID{},
		Title:        fake.Sentence(),
		StartDate:    time.Date(2023, time.January, 10, 10, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2023, time.January, 10, 11, 0, 0, 0, time.UTC),
		Description:  fake.Paragraph(),
		UserID:       userID,
		NotifyBefore: 24 * time.Hour,
	}

	req := EventRequest{Event: marshalEvent(event)}
	res, err := testClient.CreateEvent(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))
	result, _ := unmarshalEvent(res, userID)
	event.ID = result.ID
	require.Equal(t, event, result)

	storedEvent, _ := testStorage.GetEvent(result.ID)
	require.Equal(t, storedEvent, result)
}

func TestDeleteEvent(t *testing.T) {
	userID := uuid.New()
	event := storage.Event{
		ID:           uuid.UUID{},
		Title:        fake.Sentence(),
		StartDate:    time.Date(2023, time.January, 10, 10, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2023, time.January, 10, 11, 0, 0, 0, time.UTC),
		Description:  fake.Paragraph(),
		UserID:       userID,
		NotifyBefore: 24 * time.Hour,
	}
	event, _ = testStorage.AddEvent(event)

	req := EventIdRequest{Id: event.ID.String()}
	res, err := testClient.DeleteEvent(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))
	require.IsType(t, &DeleteEventResponse{}, res)

	_, err = testStorage.GetEvent(event.ID)
	require.ErrorIs(t, err, storage.ErrEventNotFound)
}

func TestGetForDay(t *testing.T) {
	userID := uuid.New()
	events := []storage.Event{
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 10, 13, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 10, 14, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 10, 10, 30, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 10, 11, 30, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 11, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 11, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
	}
	for i, event := range events {
		event, _ = testStorage.AddEvent(event)
		events[i] = event
	}

	start := timestamppb.New(events[0].StartDate)
	req := StartDateRequest{Start: start}
	res, err := testClient.GetForDay(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))

	result := make([]storage.Event, 2)
	for i := 0; i < 2; i++ {
		result[i], _ = unmarshalEvent(res.Events[i], userID)
	}

	expected := []storage.Event{events[1], events[0]}
	require.Equal(t, expected, result)
}

func TestGetForWeek(t *testing.T) {
	userID := uuid.New()

	events := []storage.Event{
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 16, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 16, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 20, 10, 30, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 20, 11, 30, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 10, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 10, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 11, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 11, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
	}
	for i, event := range events {
		event, _ = testStorage.AddEvent(event)
		events[i] = event
	}

	start := timestamppb.New(events[2].StartDate)
	req := StartDateRequest{Start: start}
	res, err := testClient.GetForWeek(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))

	result := make([]storage.Event, 3)
	for i := 0; i < 3; i++ {
		result[i], _ = unmarshalEvent(res.Events[i], userID)
	}

	expected := []storage.Event{events[2], events[3], events[0]}
	require.Equal(t, expected, result)
}

func TestGetForMonth(t *testing.T) {
	userID := uuid.New()

	events := []storage.Event{
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.January, 16, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.January, 16, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.February, 20, 10, 30, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.February, 20, 11, 30, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.March, 10, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.March, 10, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.UUID{},
			Title:        fake.Sentence(),
			StartDate:    time.Date(2023, time.February, 15, 10, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2023, time.February, 15, 11, 0, 0, 0, time.UTC),
			Description:  fake.Paragraph(),
			UserID:       userID,
			NotifyBefore: 24 * time.Hour,
		},
	}
	for i, event := range events {
		event, _ = testStorage.AddEvent(event)
		events[i] = event
	}

	start := timestamppb.New(events[0].StartDate)
	req := StartDateRequest{Start: start}
	res, err := testClient.GetForMonth(requestContext(userID), &req)
	require.Equal(t, codes.OK, errCode(t, err))

	result := make([]storage.Event, 2)
	for i := 0; i < 2; i++ {
		result[i], _ = unmarshalEvent(res.Events[i], userID)
	}

	expected := []storage.Event{events[0], events[3]}
	require.Equal(t, expected, result)
}

func TestGetEventAccessDenied(t *testing.T) {
	userID := uuid.New()

	event := storage.Event{
		ID:           uuid.UUID{},
		Title:        fake.Sentence(),
		StartDate:    time.Date(2023, time.January, 10, 10, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2023, time.January, 10, 11, 0, 0, 0, time.UTC),
		Description:  fake.Paragraph(),
		UserID:       userID,
		NotifyBefore: 24 * time.Hour,
	}
	event, _ = testStorage.AddEvent(event)

	req := EventIdRequest{Id: event.ID.String()}
	res, err := testClient.GetEvent(requestContext(uuid.New()), &req)
	require.Equal(t, codes.PermissionDenied, errCode(t, err))
	require.Nil(t, res)

	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	res, err = testClient.GetEvent(ctx, &req)
	require.Equal(t, codes.Unauthenticated, errCode(t, err))
	require.Nil(t, res)
}
