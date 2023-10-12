package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrUnknownCondition = errors.New("uknown condition")
	ErrIncomparableType = errors.New("incomparable type")
)

type EventField string

const (
	EventID           EventField = "ID"
	EventTitle        EventField = "Title"
	EventStartDate    EventField = "StartDate"
	EventEndDate      EventField = "EndDate"
	EventDescription  EventField = "Description"
	EventUserID       EventField = "UserID"
	EventNotifyBefore EventField = "NotifyBefore"
	EventNotifiedAt   EventField = "NotifiedAt"
)

type ConditionType string

const (
	TypeEq       ConditionType = "="
	TypeNotEq    ConditionType = "!="
	TypeLess     ConditionType = "<"
	TypeLessOrEq ConditionType = "<="
	TypeMore     ConditionType = ">"
	TypeMoreOrEq ConditionType = ">="
	TypeIn       ConditionType = "IN"
	TypeNotIn    ConditionType = "NOT IN"
)

type SortDirection string

const (
	DirectionAsc  SortDirection = "ASC"
	DirectionDesc SortDirection = "DESC"
)

type EventCondition struct {
	Field  EventField
	Type   ConditionType
	Sample interface{}
}

type EventSort struct {
	Field     EventField
	Direction SortDirection
}

type Event struct {
	ID           uuid.UUID     `db:"id"`
	Title        string        `db:"title"`
	StartDate    time.Time     `db:"start_date"`
	EndDate      time.Time     `db:"end_date"`
	Description  string        `db:"description"`
	UserID       uuid.UUID     `db:"user_id"`
	NotifyBefore time.Duration `db:"notify_before"`
	NotifiedAt   time.Time     `db:"notified_at"`
}

func (e Event) GetFieldValue(field EventField) interface{} {
	switch field {
	case EventID:
		return e.ID
	case EventTitle:
		return e.Title
	case EventStartDate:
		return e.StartDate
	case EventEndDate:
		return e.EndDate
	case EventDescription:
		return e.Description
	case EventUserID:
		return e.UserID
	case EventNotifyBefore:
		return e.NotifyBefore
	case EventNotifiedAt:
		return e.NotifiedAt
	}

	return nil
}
