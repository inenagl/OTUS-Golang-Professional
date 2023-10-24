//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/json"
	internalhttp "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/http"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type CalendarRESTSuite struct {
	suite.Suite
	addr   string
	userID string
	client http.Client
	ctx    context.Context
}

func (s *CalendarRESTSuite) SetupSuite() {
	httpAddr := os.Getenv("GOCLNDR_HTTPADDR")
	if httpAddr == "" {
		httpAddr = ":8888"
	}
	s.addr = "http://" + httpAddr
	s.client = http.Client{}
	s.ctx, _ = context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet
}

func (s *CalendarRESTSuite) TeardownSuite() {
}

func (s *CalendarRESTSuite) SetupTest() {
	s.userID = uuid.New().String()
}

func (s *CalendarRESTSuite) TeardownTest() {
	s.userID = ""
}

func (s *CalendarRESTSuite) createEventReq(event *storage.Event) *http.Request {
	eventData := json.MarshallEvent(*event, marshalledFields)
	req, err := http.NewRequestWithContext(s.ctx, "POST", s.addr+"/event", bytes.NewReader([]byte(eventData)))
	s.Require().NoError(err)
	req.Header.Add(internalhttp.UserIDHeader, s.userID)

	return req
}

func (s *CalendarRESTSuite) updateEventReq(event *storage.Event) *http.Request {
	eventJSON := json.MarshallEvent(*event, marshalledFields)
	req, err := http.NewRequestWithContext(
		s.ctx,
		"POST",
		s.addr+"/event/"+event.ID.String(),
		bytes.NewReader([]byte(eventJSON)),
	)
	s.Require().NoError(err)
	req.Header.Add(internalhttp.UserIDHeader, s.userID)

	return req
}

func (s *CalendarRESTSuite) getEventReq(id uuid.UUID) *http.Request {
	req, err := http.NewRequestWithContext(s.ctx, "GET", s.addr+"/event/"+id.String(), nil)
	s.Require().NoError(err)
	req.Header.Add(internalhttp.UserIDHeader, s.userID)

	return req
}

func (s *CalendarRESTSuite) deleteEventReq(id uuid.UUID) *http.Request {
	req, err := http.NewRequestWithContext(s.ctx, "DELETE", s.addr+"/event/"+id.String(), nil)
	s.Require().NoError(err)
	req.Header.Add(internalhttp.UserIDHeader, s.userID)

	return req
}

func (s *CalendarRESTSuite) getForPeriodReq(day time.Time, period internalhttp.SearchPeriod) *http.Request {
	address := ""
	switch period { //nolint: exhaustive
	case internalhttp.Day:
		address = s.addr + "/events/day/" + day.Format(time.DateOnly)
	case internalhttp.Week:
		address = s.addr + "/events/week/" + day.Format(time.DateOnly)
	case internalhttp.Month:
		address = s.addr + "/events/month/" + day.Format(time.DateOnly)
	default:
		s.T().Fail()
	}
	req, err := http.NewRequestWithContext(s.ctx, "GET", address, nil)
	s.Require().NoError(err)
	req.Header.Add(internalhttp.UserIDHeader, s.userID)

	return req
}

func (s *CalendarRESTSuite) getForDayReq(day time.Time) *http.Request {
	return s.getForPeriodReq(day, internalhttp.Day)
}

func (s *CalendarRESTSuite) getForWeekReq(day time.Time) *http.Request {
	return s.getForPeriodReq(day, internalhttp.Week)
}

func (s *CalendarRESTSuite) getForMonthReq(day time.Time) *http.Request {
	return s.getForPeriodReq(day, internalhttp.Month)
}

func (s *CalendarRESTSuite) eventFromHTTPBody(body io.Reader) storage.Event {
	eventData, err := io.ReadAll(body)
	s.Require().NoError(err)
	var resEvent storage.Event
	err = json.UnmarshallEvent(string(eventData), &resEvent, unmarshalledFields)
	s.Require().NoError(err)

	return resEvent
}

func (s *CalendarRESTSuite) eventsFromHTTPBody(body io.Reader) []storage.Event {
	eventsData, err := io.ReadAll(body)
	s.Require().NoError(err)

	var resEvent storage.Event
	parsed := gjson.ParseBytes(eventsData)
	result := make([]storage.Event, len(parsed.Array()))
	for i, data := range parsed.Array() {
		err = json.UnmarshallEvent(data.Raw, &resEvent, unmarshalledFields)
		s.Require().NoError(err)
		result[i] = resEvent
	}

	return result
}

func (s *CalendarRESTSuite) TestUnauthorizedRequest() {
	event := randomEvent()
	id := uuid.New()
	s.userID = ""

	s.T().Run(
		"create", func(t *testing.T) {
			res, err := s.client.Do(s.createEventReq(&event))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)

	s.T().Run(
		"get", func(t *testing.T) {
			res, err := s.client.Do(s.getEventReq(id))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)

	s.T().Run(
		"update", func(t *testing.T) {
			res, err := s.client.Do(s.updateEventReq(&event))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)

	s.T().Run(
		"delete", func(t *testing.T) {
			res, err := s.client.Do(s.deleteEventReq(id))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)

	s.T().Run(
		"get for day", func(t *testing.T) {
			res, err := s.client.Do(s.getForDayReq(randomDay()))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)

	s.T().Run(
		"get for week", func(t *testing.T) {
			res, err := s.client.Do(s.getForWeekReq(randomDay()))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)

	s.T().Run(
		"get for month", func(t *testing.T) {
			res, err := s.client.Do(s.getForMonthReq(randomDay()))
			s.Require().NoError(err)
			defer res.Body.Close()
			s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
		},
	)
}

func (s *CalendarRESTSuite) TestCreateEvent() {
	event := randomEvent()

	res, err := s.client.Do(s.createEventReq(&event))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusOK, res.StatusCode)

	resEvent := s.eventFromHTTPBody(res.Body)
	event.ID = resEvent.ID

	// Проверяем ответ метода создания.
	s.Require().Equal(event, resEvent)

	// Проверяем, что будут возвращаться правильные данные.
	res, err = s.client.Do(s.getEventReq(event.ID))
	s.Require().NoError(err)
	defer res.Body.Close()
	resEvent = s.eventFromHTTPBody(res.Body)
	s.Require().Equal(event, resEvent)
}

func (s *CalendarRESTSuite) TestUpdateEvent() {
	event := randomEvent()

	res, err := s.client.Do(s.createEventReq(&event))
	s.Require().NoError(err)
	defer res.Body.Close()
	resEvent := s.eventFromHTTPBody(res.Body)
	eventID := resEvent.ID

	// Обновим созданное новыми данными.
	event = randomEvent()
	event.ID = eventID
	res, err = s.client.Do(s.updateEventReq(&event))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusOK, res.StatusCode)
	resEvent = s.eventFromHTTPBody(res.Body)

	// Проверяем ответ метода обновления.
	s.Require().Equal(event, resEvent)

	// Проверяем, что возвращаются обновленные данные.
	res, err = s.client.Do(s.getEventReq(eventID))
	s.Require().NoError(err)
	defer res.Body.Close()
	resEvent = s.eventFromHTTPBody(res.Body)
	s.Require().Equal(event, resEvent)
}

func (s *CalendarRESTSuite) TestDeleteEvent() {
	event := randomEvent()

	res, err := s.client.Do(s.createEventReq(&event))
	s.Require().NoError(err)
	defer res.Body.Close()
	resEvent := s.eventFromHTTPBody(res.Body)
	eventID := resEvent.ID

	res, err = s.client.Do(s.deleteEventReq(eventID))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusOK, res.StatusCode)

	// Проверяем, что после удаления ничего не возвращается.
	res, err = s.client.Do(s.getEventReq(eventID))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusNotFound, res.StatusCode)
}

func (s *CalendarRESTSuite) TestGetForDay() {
	day := randomDay()
	var anotherDay time.Time
	for anotherDay = randomDay(); anotherDay.Unix() == day.Unix(); {
		continue
	}
	events := make([]storage.Event, 5)
	var ev storage.Event
	var res *http.Response
	var err error
	for i := 0; i < 5; i++ {
		ev = randomEvent()
		// первые три в текущем дне, остальные - в другом дне
		if i < 3 {
			ev.StartDate = day.Add(time.Duration(i) * time.Hour)
			ev.EndDate = day.Add(time.Duration(i+1) * time.Hour)
		} else {
			ev.StartDate = anotherDay.Add(time.Duration(i) * time.Hour)
			ev.EndDate = anotherDay.Add(time.Duration(i+1) * time.Hour)
		}
		res, err = s.client.Do(s.createEventReq(&ev))
		s.Require().NoError(err)
		defer res.Body.Close()
		resEvent := s.eventFromHTTPBody(res.Body)
		events[i] = resEvent
	}

	res, err = s.client.Do(s.getForDayReq(day))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusOK, res.StatusCode)

	expected := []storage.Event{
		events[0],
		events[1],
		events[2],
	}

	actual := s.eventsFromHTTPBody(res.Body)
	s.Require().EqualValues(expected, actual)
}

func (s *CalendarRESTSuite) TestGetForWeek() {
	firstDay := randomDay()
	// Делаем 10 подряд идущих дней.
	days := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		days[i] = firstDay.AddDate(0, 0, i)
	}

	events := make([]storage.Event, 10)
	var ev storage.Event
	var start time.Time
	var res *http.Response
	var err error
	// Делаем по событию в каждом дне.
	for i := 0; i < 10; i++ {
		ev = randomEvent()
		start = days[i].Add(randomDuration())
		ev.StartDate = start
		ev.EndDate = start.Add(time.Hour)

		res, err = s.client.Do(s.createEventReq(&ev))
		s.Require().NoError(err)
		defer res.Body.Close()
		resEvent := s.eventFromHTTPBody(res.Body)
		events[i] = resEvent
	}

	res, err = s.client.Do(s.getForWeekReq(firstDay))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusOK, res.StatusCode)

	// Ожидаем в ответе первые 7 событий (в течение недели).
	expected := make([]storage.Event, 7)
	for i := 0; i < 7; i++ {
		expected[i] = events[i]
	}
	actual := s.eventsFromHTTPBody(res.Body)
	s.Require().EqualValues(expected, actual)
}

func (s *CalendarRESTSuite) TestGetForMonth() {
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

	events := make([]storage.Event, 10)
	var ev storage.Event
	var start time.Time
	var res *http.Response
	var err error
	// Делаем по событию в каждом дне.
	for i := 0; i < 10; i++ {
		ev = randomEvent()
		start = days[i].Add(randomDuration())
		ev.StartDate = start
		ev.EndDate = start.Add(time.Hour)

		res, err = s.client.Do(s.createEventReq(&ev))
		s.Require().NoError(err)
		defer res.Body.Close()
		resEvent := s.eventFromHTTPBody(res.Body)
		events[i] = resEvent
	}

	res, err = s.client.Do(s.getForMonthReq(firstDay))
	s.Require().NoError(err)
	defer res.Body.Close()
	s.Require().Equal(http.StatusOK, res.StatusCode)

	expected := []storage.Event{
		events[0],
		events[1],
		events[3],
		events[6],
		events[8],
		events[9],
	}
	actual := s.eventsFromHTTPBody(res.Body)
	s.Require().EqualValues(expected, actual)
}

func TestCalendarRESTSuite(t *testing.T) {
	suite.Run(t, new(CalendarRESTSuite))
}
