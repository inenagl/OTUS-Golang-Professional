package internalhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/app"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/json"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/logger"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/jackc/fake"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	testHost = "localhost"
	testPort = 8081
	testURI  = "http://localhost:8081"
)

var (
	testClient  *http.Client
	testStorage app.Storage
	testApp     app.Application
)

type testAPIMethod int

const (
	testMethodUndefined testAPIMethod = iota
	testMethodGetEvent
	testMethodCreateEvent
	testMethodUpdateEvent
	testMethodDeleteEvent
	testMethodGetForDay
	testMethodGetForWeek
	testMethodGetForMonth
)

var testUris = map[testAPIMethod]string{
	testMethodUndefined:   testURI + "/hello",
	testMethodGetEvent:    testURI + "/event/%s",
	testMethodCreateEvent: testURI + "/event",
	testMethodUpdateEvent: testURI + "/event/%s",
	testMethodDeleteEvent: testURI + "/event/%s",
	testMethodGetForDay:   testURI + "/events/day/%s",
	testMethodGetForWeek:  testURI + "/events/week/%s",
	testMethodGetForMonth: testURI + "/events/month/%s",
}

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

	testClient = &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(contextTimeout(), http.MethodGet, testUris[testMethodUndefined], nil)
	if err != nil {
		os.Exit(2)
	}
	tm := time.NewTimer(3 * time.Second)
	for res, err := testClient.Do(req); err != nil; {
		res.Body.Close()
		select {
		case <-tm.C:
			break
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if err != nil {
		os.Exit(3)
	}

	m.Run()

	cancel()
	wg.Wait()
}

func contextTimeout() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second) //nolint: govet
	return ctx
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

	uri := fmt.Sprintf(testUris[testMethodGetEvent], event.ID.String())
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodGet, uri, nil)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var result storage.Event
	err = json.UnmarshallEvent(string(body), &result, unmarshalledFields)
	require.NoError(t, err)
	result.ID = event.ID
	result.UserID = userID

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

	reqBody := bytes.NewReader([]byte(json.MarshallEvent(updatedEvent, marshalledFields)))
	uri := fmt.Sprintf(testUris[testMethodUpdateEvent], event.ID.String())
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodPost, uri, reqBody)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var result storage.Event
	err = json.UnmarshallEvent(string(body), &result, unmarshalledFields)
	require.NoError(t, err)
	result.ID = event.ID
	result.UserID = userID

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

	reqBody := bytes.NewReader([]byte(json.MarshallEvent(event, marshalledFields)))
	uri := testUris[testMethodCreateEvent]
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodPost, uri, reqBody)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	eventID, _ := uuid.Parse(gjson.Get(string(body), string(json.EventID)).String())
	event.ID = eventID

	var result storage.Event
	err = json.UnmarshallEvent(string(body), &result, unmarshalledFields)
	require.NoError(t, err)
	result.ID = eventID
	result.UserID = userID

	require.Equal(t, event, result)

	storedEvent, _ := testStorage.GetEvent(eventID)
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

	uri := fmt.Sprintf(testUris[testMethodDeleteEvent], event.ID.String())
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodDelete, uri, nil)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	require.Equal(t, `{"status":"ok"}`, string(body))

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

	uri := fmt.Sprintf(testUris[testMethodGetForDay], events[0].StartDate.Format(time.DateOnly))
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodGet, uri, nil)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	result := make([]storage.Event, 2)
	for i, val := range gjson.Parse(string(body)).Array() {
		err = json.UnmarshallEvent(val.Raw, &result[i], unmarshalledFields)
		require.NoError(t, err)
		result[i].ID, _ = uuid.Parse(val.Get("ID").String())
		result[i].UserID = userID
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

	uri := fmt.Sprintf(testUris[testMethodGetForWeek], events[2].StartDate.Format(time.DateOnly))
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodGet, uri, nil)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	result := make([]storage.Event, 3)
	for i, val := range gjson.Parse(string(body)).Array() {
		err = json.UnmarshallEvent(val.Raw, &result[i], unmarshalledFields)
		require.NoError(t, err)
		result[i].ID, _ = uuid.Parse(val.Get("ID").String())
		result[i].UserID = userID
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

	uri := fmt.Sprintf(testUris[testMethodGetForMonth], events[0].StartDate.Format(time.DateOnly))
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodGet, uri, nil)
	req.Header.Add(UserIDHeader, userID.String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	result := make([]storage.Event, 2)
	for i, val := range gjson.Parse(string(body)).Array() {
		err = json.UnmarshallEvent(val.Raw, &result[i], unmarshalledFields)
		require.NoError(t, err)
		result[i].ID, _ = uuid.Parse(val.Get("ID").String())
		result[i].UserID = userID
	}

	expected := []storage.Event{events[0], events[3]}
	require.Equal(t, expected, result)
}

func TestGetEventUnauthorized(t *testing.T) {
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

	uri := fmt.Sprintf(testUris[testMethodGetEvent], event.ID.String())
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodGet, uri, nil)

	res, err := testClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestGetEventForbidden(t *testing.T) {
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

	uri := fmt.Sprintf(testUris[testMethodGetEvent], event.ID.String())
	req, _ := http.NewRequestWithContext(contextTimeout(), http.MethodGet, uri, nil)
	req.Header.Add(UserIDHeader, uuid.New().String())

	res, err := testClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusForbidden, res.StatusCode)
}
