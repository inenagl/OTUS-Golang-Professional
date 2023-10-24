package sender

import (
	"context"
	"time"

	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/json"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/queue"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var unmarshalledFields = []json.EventField{
	json.EventID,
	json.EventUserID,
	json.EventTitle,
	json.EventDescription,
	json.EventStartDate,
	json.EventEndDate,
	json.EventNotifyBefore,
}

type Sender interface {
	Start(ctx context.Context) error
	Stop() error
}

type S struct {
	threads  int
	logger   *zap.Logger
	consumer queue.Consumer
	cancel   context.CancelFunc
	producer queue.Producer
}

func New(threads int, logg *zap.Logger, consumer queue.Consumer, producer queue.Producer) *S {
	return &S{
		threads:  threads,
		logger:   logg,
		consumer: consumer,
		producer: producer,
	}
}

func (s *S) Start(ctx context.Context) error {
	s.logger.Debug("consumer connect...")
	if err := s.consumer.Connect(); err != nil {
		return err
	}
	if err := s.producer.Connect(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	return s.consumer.Consume(ctx, s.Handle, s.threads)
}

func (s *S) Stop() error {
	s.logger.Debug("cancel in stop")
	s.cancel()
	err := s.consumer.Close()
	if err != nil {
		return err
	}
	err = s.producer.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *S) Handle(ctx context.Context, deliveries <-chan amqp.Delivery) {
	var event storage.Event
	var err error

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("done in handle")
			return
		case msg := <-deliveries:
			if err = json.UnmarshallEvent(string(msg.Body), &event, unmarshalledFields); err != nil {
				s.logger.Error("invalid message: "+err.Error(), zap.String("json", string(msg.Body)))
				continue
			}
			if err = s.send(event); err != nil {
				s.logger.Error("failed to send: "+err.Error(), zap.String("EventID", event.ID.String()))
				continue
			}
			if err = msg.Ack(false); err != nil {
				s.logger.Error("failed to ack: "+err.Error(), zap.String("EventID", event.ID.String()))
				continue
			}
		}
	}
}

func (s *S) send(event storage.Event) error {
	if err := s.producer.Publish(json.MarshallEvent(event, unmarshalledFields)); err != nil {
		s.logger.Error("send to queue: " + err.Error())
		return err
	}

	s.logger.Info(
		"notification is sent",
		zap.String("EventID", event.ID.String()),
		zap.String("UserID", event.UserID.String()),
		zap.String("Title", event.Title),
		zap.String("Description", event.Description),
		zap.String("StartDate", event.StartDate.Format(time.DateTime)),
		zap.String("EndDate", event.EndDate.Format(time.DateTime)),
		zap.String("NotifyBefore", event.NotifyBefore.String()),
	)

	return nil
}
