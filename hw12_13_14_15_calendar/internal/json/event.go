package json

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/tidwall/gjson"
)

type EventField string

const (
	EventID           = EventField(storage.EventID)
	EventTitle        = EventField(storage.EventTitle)
	EventStartDate    = EventField(storage.EventStartDate)
	EventEndDate      = EventField(storage.EventEndDate)
	EventDescription  = EventField(storage.EventDescription)
	EventNotifyBefore = EventField(storage.EventNotifyBefore)
	EventUserID       = EventField(storage.EventUserID)
	EventNotifiedAt   = EventField(storage.EventNotifiedAt)
)

type handler func(event storage.Event) string

var marshallMap = map[EventField]handler{
	EventID:           func(event storage.Event) string { return event.ID.String() },
	EventTitle:        func(event storage.Event) string { return event.Title },
	EventStartDate:    func(event storage.Event) string { return event.StartDate.Format(time.DateTime) },
	EventEndDate:      func(event storage.Event) string { return event.EndDate.Format(time.DateTime) },
	EventDescription:  func(event storage.Event) string { return event.Description },
	EventNotifyBefore: func(event storage.Event) string { return event.NotifyBefore.String() },
	EventUserID:       func(event storage.Event) string { return event.UserID.String() },
	EventNotifiedAt:   func(event storage.Event) string { return event.NotifiedAt.Format(time.DateTime) },
}

type FieldParseErr struct {
	Err   error
	Field EventField
}

func (e FieldParseErr) Error() string {
	return e.Err.Error()
}

func MarshallEvent(event storage.Event, fields []EventField) string {
	if len(fields) == 0 {
		return "{}"
	}
	strArr := make([]string, len(fields))
	for i, field := range fields {
		strArr[i] = fmt.Sprintf(`"%s":"%s"`, string(field), marshallMap[field](event))
	}

	return fmt.Sprintf(
		`{%s}`,
		strings.Join(strArr, ","),
	)
}

func MarshallEvents(events []storage.Event, fields []EventField) string {
	b := strings.Builder{}
	b.WriteString("[")
	l := len(events) - 1

	for i := 0; i <= l; i++ {
		b.WriteString(MarshallEvent(events[i], fields))
		if i != l {
			b.WriteString(",")
		}
	}
	b.WriteString("]")

	return b.String()
}

func UnmarshallEvent(source string, target *storage.Event, fields []EventField) error {
	for _, field := range fields {
		value := gjson.Get(source, string(field))
		if !value.Exists() {
			continue
		}
		switch field {
		case EventID:
			id, err := uuid.Parse(value.String())
			if err != nil {
				return FieldParseErr{err, field}
			}
			target.ID = id
		case EventTitle:
			target.Title = value.String()
		case EventStartDate:
			tm, err := time.Parse(time.DateTime, value.String())
			if err != nil {
				return FieldParseErr{err, field}
			}
			target.StartDate = tm
		case EventEndDate:
			tm, err := time.Parse(time.DateTime, value.String())
			if err != nil {
				return FieldParseErr{err, field}
			}
			target.EndDate = tm
		case EventNotifyBefore:
			dr, err := time.ParseDuration(value.String())
			if err != nil {
				return FieldParseErr{err, field}
			}
			target.NotifyBefore = dr
		case EventDescription:
			target.Description = value.String()
		case EventUserID:
			id, err := uuid.Parse(value.String())
			if err != nil {
				return FieldParseErr{err, field}
			}
			target.UserID = id
		case EventNotifiedAt:
			tm, err := time.Parse(time.DateTime, value.String())
			if err != nil {
				return FieldParseErr{err, field}
			}
			target.NotifiedAt = tm
		default:
			continue
		}
	}

	return nil
}
