package internalhttp

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/app"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type EventField string

const (
	EventID           = EventField(storage.EventID)
	EventTitle        = EventField(storage.EventTitle)
	EventStartDate    = EventField(storage.EventStartDate)
	EventEndDate      = EventField(storage.EventEndDate)
	EventDescription  = EventField(storage.EventDescription)
	EventNotifyBefore = EventField(storage.EventNotifyBefore)
)

type SearchPeriod int

const (
	Undefined SearchPeriod = iota
	Day
	Week
	Month
)

type fieldParseErr struct {
	err   error
	field EventField
}

func (e fieldParseErr) Error() string {
	return e.err.Error()
}

func marshallEvent(event storage.Event) string {
	return fmt.Sprintf(`{"%s":"%s","%s":"%s","%s":"%s","%s":"%s","%s":"%s","%s":"%s"}`,
		EventID, event.ID.String(),
		EventTitle, event.Title,
		EventStartDate, event.StartDate.Format(time.DateTime),
		EventEndDate, event.EndDate.Format(time.DateTime),
		EventDescription, event.Description,
		EventNotifyBefore, event.NotifyBefore.String(),
	)
}

func unmarshallEvent(source string, target *storage.Event) error {
	fields := []EventField{
		EventTitle,
		EventDescription,
		EventStartDate,
		EventEndDate,
		EventNotifyBefore,
	}

	for _, field := range fields {
		value := gjson.Get(source, string(field))
		if !value.Exists() {
			continue
		}
		switch field { //nolint:exhaustive
		case EventTitle:
			target.Title = value.String()
		case EventDescription:
			target.Description = value.String()
		case EventStartDate:
			tm, err := time.Parse(time.DateTime, value.String())
			if err != nil {
				return fieldParseErr{err, field}
			}
			target.StartDate = tm
		case EventEndDate:
			tm, err := time.Parse(time.DateTime, value.String())
			if err != nil {
				return fieldParseErr{err, field}
			}
			target.EndDate = tm
		case EventNotifyBefore:
			dr, err := time.ParseDuration(value.String())
			if err != nil {
				return fieldParseErr{err, field}
			}
			target.NotifyBefore = dr
		default:
			continue
		}
	}

	return nil
}

func marshallEvents(events []storage.Event) string {
	b := strings.Builder{}
	b.WriteString("[")
	l := len(events) - 1

	for i := 0; i <= l; i++ {
		b.WriteString(marshallEvent(events[i]))
		if i != l {
			b.WriteString(",")
		}
	}
	b.WriteString("]")

	return b.String()
}

func (s Server) write(w http.ResponseWriter, resp string) {
	_, err := w.Write([]byte(resp))
	if err != nil {
		s.logger.Error(err.Error())
	}
}

func (s Server) writeError(w http.ResponseWriter, msg string) {
	s.write(w, fmt.Sprintf(`{"error":"%s"}`, msg))
}

func (s Server) getUserIDFromRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	uid := r.Context().Value(UserIDKey)
	if uid == nil {
		w.WriteHeader(http.StatusUnauthorized)
		s.writeError(w, "eventId is not set")
		return uuid.UUID{}, false
	}

	return uid.(uuid.UUID), true
}

func (s Server) getEventByRequest(w http.ResponseWriter, r *http.Request) (storage.Event, bool) {
	eventID := mux.Vars(r)["eventId"]
	if eventID == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.writeError(w, "eventId is not set")
		return storage.Event{}, false
	}

	eid, err := uuid.Parse(eventID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.writeError(w, err.Error())
		return storage.Event{}, false
	}

	uid, ok := s.getUserIDFromRequest(w, r)
	if !ok {
		return storage.Event{}, false
	}

	event, err := s.app.GetEvent(eid, uid)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, app.ErrAccessDenied):
			w.WriteHeader(http.StatusForbidden)
		default:
			s.logger.Error(err.Error(), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		s.writeError(w, err.Error())
		return storage.Event{}, false
	}

	return event, true
}

func (s Server) getValidJSONFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	scanner := bufio.NewScanner(r.Body)
	b := strings.Builder{}
	for scanner.Scan() {
		b.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.writeError(w, err.Error())
		return "", false
	}

	json := b.String()
	if !gjson.Valid(json) {
		w.WriteHeader(http.StatusBadRequest)
		s.writeError(w, "invalid json")
		return "", false
	}

	return json, true
}

func (s Server) getDateFromRequest(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	t, err := time.Parse(time.DateOnly, mux.Vars(r)["date"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.writeError(w, err.Error())
		return time.Now(), false
	}

	return t, true
}

func (s Server) getEvent(w http.ResponseWriter, r *http.Request) {
	event, ok := s.getEventByRequest(w, r)
	if !ok {
		return
	}

	w.WriteHeader(http.StatusOK)
	s.write(w, marshallEvent(event))
}

func (s Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	json, ok := s.getValidJSONFromReq(w, r)
	if !ok {
		return
	}

	target, ok := s.getEventByRequest(w, r)
	if !ok {
		return
	}

	err := unmarshallEvent(json, &target)
	if err != nil {
		var ie fieldParseErr
		if errors.As(err, &ie) {
			w.WriteHeader(http.StatusBadRequest)
			s.writeError(w, "invalid value in "+string(ie.field))
		} else {
			s.logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			s.writeError(w, err.Error())
		}
		return
	}

	res, err := s.app.UpdateEvent(target.ID, target.UserID, target)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, app.ErrAccessDenied):
			w.WriteHeader(http.StatusForbidden)
		default:
			s.logger.Error(err.Error(), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		s.writeError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	s.write(w, marshallEvent(res))
}

func (s Server) createEvent(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	json, ok := s.getValidJSONFromReq(w, r)
	if !ok {
		return
	}

	target := storage.Event{}

	err := unmarshallEvent(json, &target)
	if err != nil {
		var ie fieldParseErr
		if errors.As(err, &ie) {
			w.WriteHeader(http.StatusBadRequest)
			s.writeError(w, "invalid value in "+string(ie.field))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			s.writeError(w, err.Error())
			s.logger.Error(err.Error())
		}
		return
	}

	res, err := s.app.CreateEvent(uid, target)
	if err != nil {
		s.logger.Error(err.Error(), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		s.writeError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	s.write(w, marshallEvent(res))
}

func (s Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	event, ok := s.getEventByRequest(w, r)
	if !ok {
		return
	}

	if err := s.app.DeleteEvent(event.ID, uid); err != nil {
		s.logger.Error(err.Error(), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		s.writeError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	s.write(w, `{"status":"ok"}`)
}

func (s Server) getForPeriod(w http.ResponseWriter, r *http.Request, period SearchPeriod) {
	uid, ok := s.getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	start, ok := s.getDateFromRequest(w, r)
	if !ok {
		return
	}
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
		s.logger.Error(err.Error(), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		s.writeError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	s.write(w, marshallEvents(events))
}

func (s Server) getForDay(w http.ResponseWriter, r *http.Request) {
	s.getForPeriod(w, r, Day)
}

func (s Server) getForWeek(w http.ResponseWriter, r *http.Request) {
	s.getForPeriod(w, r, Week)
}

func (s Server) getForMonth(w http.ResponseWriter, r *http.Request) {
	s.getForPeriod(w, r, Month)
}
