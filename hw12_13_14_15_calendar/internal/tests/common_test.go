//go:build integration

package integration_test

import (
	"math/rand"
	"time"

	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/json"
	grpcapi "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/jackc/fake"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var marshalledFields = []json.EventField{
	json.EventID,
	json.EventTitle,
	json.EventStartDate,
	json.EventEndDate,
	json.EventDescription,
	json.EventNotifyBefore,
}

var unmarshalledFields = []json.EventField{
	json.EventID,
	json.EventTitle,
	json.EventDescription,
	json.EventStartDate,
	json.EventEndDate,
	json.EventNotifyBefore,
}

func randomDay() time.Time {
	return time.Date(fake.Year(2023, 2025), time.Month((rand.Intn(12) + 1)), fake.Day(), 0, 0, 0, 0, time.UTC)
}

func randomDuration() time.Duration {
	return time.Duration(rand.Intn(24))*time.Hour + time.Duration(rand.Intn(59))*time.Minute
}

func randomGRPCEvent() grpcapi.Event {
	start := randomDay().Add(randomDuration())
	return grpcapi.Event{
		Title:        fake.Sentence(),
		Description:  fake.Paragraph(),
		StartDate:    timestamppb.New(start),
		EndDate:      timestamppb.New(start.Add(randomDuration())),
		NotifyBefore: durationpb.New(randomDuration()),
	}
}

func mapFromGRPCEvent(event *grpcapi.Event) map[string]interface{} {
	return map[string]interface{}{
		"Title":        event.Title,
		"Description":  event.Description,
		"StartDate":    event.StartDate.AsTime().Format(time.DateTime),
		"EndDate":      event.EndDate.AsTime().Format(time.DateTime),
		"NotifyBefore": event.NotifyBefore.AsDuration(),
	}
}

func mapFromEvent(event storage.Event) map[string]interface{} {
	return map[string]interface{}{
		"Title":        event.Title,
		"Description":  event.Description,
		"StartDate":    event.StartDate.Format(time.DateTime),
		"EndDate":      event.EndDate.Format(time.DateTime),
		"NotifyBefore": event.NotifyBefore,
	}
}

func randomEvent() storage.Event {
	start := randomDay().Add(randomDuration())
	return storage.Event{
		Title:        fake.Sentence(),
		Description:  fake.Paragraph(),
		StartDate:    start,
		EndDate:      start.Add(randomDuration()),
		NotifyBefore: randomDuration(),
	}
}
