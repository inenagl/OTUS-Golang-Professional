//go:build integration

package integration_test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/json"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/queue"
	grpcapi "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CalendarSenderSuite struct {
	suite.Suite
	mu       sync.Mutex
	ctx      context.Context
	userID   string
	client   grpcapi.CalendarClient
	conn     *grpc.ClientConn
	consumer queue.Consumer
	cancel   context.CancelFunc
	events   []storage.Event
}

func (s *CalendarSenderSuite) SetupSuite() {
	grpcAddr := os.Getenv("GOCLNDR_GRPCADDR")
	if grpcAddr == "" {
		grpcAddr = ":8889"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.client = grpcapi.NewCalendarClient(conn)
	s.conn = conn

	s.userID = uuid.New().String()
	md := metadata.New(nil)
	md.Append(grpcapi.UUIDHeader, s.userID)
	s.ctx = metadata.NewOutgoingContext(context.Background(), md)

	s.setupConsumer()
}

func (s *CalendarSenderSuite) setupConsumer() {
	host := os.Getenv("GOCLNDR_AMQPHOST")
	portS := os.Getenv("GOCLNDR_AMQPPORT")
	var port int
	if portS == "" {
		port = 5672
	} else {
		p, err := strconv.Atoi(portS)
		s.Require().NoError(err)
		port = p
	}
	username := os.Getenv("GOCLNDR_AMQPUSER")
	if username == "" {
		username = "guest"
	}
	password := os.Getenv("GOCLNDR_AMQPPASSWORD")
	if password == "" {
		password = "guest"
	}
	exName := os.Getenv("GOCLNDR_EXCHANGE_NAME")
	if exName == "" {
		exName = "calendar-exchange"
	}
	exKey := os.Getenv("GOCLNDR_ROUTING_KEY")
	if exKey == "" {
		exKey = "sender-key"
	}
	exType := os.Getenv("GOCLNDR_EXCHANGE_TYPE")
	if exType == "" {
		exType = "topic"
	}
	queueName := os.Getenv("GOCLNDR_QUEUE_NAME")
	if queueName == "" {
		queueName = "sender-queue"
	}
	tag := os.Getenv("GOCLNDR_CONSUMER_TAG")
	if tag == "" {
		tag = "tests-tag"
	}

	s.consumer = queue.NewConsumer(host, port, username, password, exName, exType, exKey, queueName, tag, 10)
	err := s.consumer.Connect()
	s.Require().NoError(err)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	_, err = s.consumer.PurgeQueue()
	s.Require().NoError(err)

	go func(ctx context.Context, h queue.Handler) {
		err := s.consumer.Consume(ctx, h, 1)
		s.Require().NoError(err)
	}(ctx, s.HandleQueue)
}

func (s *CalendarSenderSuite) HandleQueue(ctx context.Context, deliveries <-chan amqp.Delivery) {
	var event storage.Event
	var err error

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-deliveries:
			err = json.UnmarshallEvent(string(msg.Body), &event, unmarshalledFields)
			s.Require().NoError(err)

			// Добавляем пришедший event в сьют для последующей проверки в тесте.
			s.mu.Lock()
			s.events = append(s.events, event)
			s.mu.Unlock()

			err = msg.Ack(false)
			s.Require().NoError(err)
		}
	}
}

func (s *CalendarSenderSuite) TeardownSuite() {
	s.cancel()
	err := s.conn.Close()
	s.Require().NoError(err)
	err = s.consumer.Close()
	s.Require().NoError(err)
}

func (s *CalendarSenderSuite) SetupTest() {
}

func (s *CalendarSenderSuite) TeardownTest() {
	s.events = []storage.Event{}
}

func (s *CalendarSenderSuite) TestSendNotification() {
	day := time.Now()

	events := make([]*grpcapi.Event, 10)
	var ev grpcapi.Event
	// Создаём ивенты со стартом сейчас и далее каждый час от текущего времени.
	for i := 0; i < 10; i++ {
		ev = randomGRPCEvent()
		ev.StartDate = timestamppb.New(day.Add(time.Duration(i) * time.Hour))
		ev.EndDate = timestamppb.New(day.Add(time.Duration(i+1) * time.Hour))
		ev.NotifyBefore = durationpb.New(15 * time.Minute)
		res, err := s.client.CreateEvent(s.ctx, &grpcapi.EventRequest{Event: &ev})
		s.Require().NoError(err)
		events[i] = res
	}
	// О первом ивенте должна быть нотификация, т.к. старт сейчас, а нотификация за 15 минут до старта.
	s.Require().Eventually(
		func() bool {
			s.mu.Lock()
			count := len(s.events)
			s.mu.Unlock()
			return count == 1
		},
		15*time.Second,
		time.Second,
	)
	// Проверяем, что уведомление было о нашем ивенте.
	s.Require().Equal(mapFromGRPCEvent(events[0]), mapFromEvent(s.events[0]))

	// Поменяем время старта у пары случайных ивентов на текущее, чтобы инициировать отправку уведомлений.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(4) + 1 // с 1 по 5,
	j := r.Intn(3) + 6 // с 6 по 9.

	events[i].StartDate = timestamppb.New(day)
	events[j].StartDate = timestamppb.New(day)

	_, err := s.client.UpdateEvent(s.ctx, &grpcapi.EventRequest{Event: events[i]})
	s.Require().NoError(err)
	// пауза, чтобы гарантировать порядок отправки уведомлений.
	time.Sleep(5 * time.Second)
	_, err = s.client.UpdateEvent(s.ctx, &grpcapi.EventRequest{Event: events[j]})
	s.Require().NoError(err)

	// Ждём появления нотификаций (теперь у нас должно быть 3 ивента).
	s.Require().Eventually(
		func() bool {
			s.mu.Lock()
			count := len(s.events)
			s.mu.Unlock()
			return count == 3
		},
		15*time.Second,
		time.Second,
	)
	// Сравниваем наши два изменённые ивента с тем, что пришло в уведомлениях.
	s.Require().Equal(mapFromGRPCEvent(events[i]), mapFromEvent(s.events[1]))
	s.Require().Equal(mapFromGRPCEvent(events[j]), mapFromEvent(s.events[2]))
	// Проверим на всякий случай, что число отправленных уведомлений всё ещё 3
	s.Require().Equal(3, len(s.events))
}

func TestCalendarSenderSuite(t *testing.T) {
	suite.Run(t, new(CalendarSenderSuite))
}
