//go:build integration
// +build integration

package integration_test

import (
	"math/rand"
	"time"

	grpcapi "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/jackc/fake"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
		"StartDate":    event.StartDate.AsTime(),
		"EndDate":      event.EndDate.AsTime(),
		"NotifyBefore": event.NotifyBefore.AsDuration(),
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
