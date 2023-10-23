package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/streadway/amqp"
)

type Handler func(context.Context, <-chan amqp.Delivery)

type Consumer interface {
	Connect() error
	Consume(ctx context.Context, handler Handler, threads int) error
	Close() error
}

type C struct {
	uri          string
	exchangeName string
	exchangeType string
	queueName    string
	routingKey   string
	consumerTag  string
	qosCount     int
	conn         *amqp.Connection
	channel      *amqp.Channel
	isConnected  bool
}

func NewConsumer(host string, port int, user, password, exchangeName, exchangeType, routingKey, queueName,
	consumerTag string, qosCount int,
) *C {
	return &C{
		uri:          fmt.Sprintf("amqp://%s:%s@%s:%d/", user, password, host, port),
		exchangeName: exchangeName,
		exchangeType: exchangeType,
		queueName:    queueName,
		routingKey:   routingKey,
		consumerTag:  consumerTag,
		qosCount:     qosCount,
	}
}

func (c *C) Connect() error {
	if c.isConnected {
		return nil
	}

	var err error
	c.conn, err = amqp.Dial(c.uri)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}

	if err = declareExchange(c.channel, c.exchangeName, c.exchangeType); err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	c.isConnected = true

	return nil
}

func (c *C) Consume(ctx context.Context, handler Handler, threads int) error {
	if !c.isConnected {
		return errors.New("not connected")
	}

	messages, err := c.consume()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go func() {
			defer wg.Done()
			handler(ctx, messages)
		}()
	}

	wg.Wait()
	return nil
}

func (c *C) consume() (<-chan amqp.Delivery, error) {
	queue, err := declareQueue(c.channel, c.queueName)
	if err != nil {
		return nil, fmt.Errorf("queue Declare: %w", err)
	}

	if err = c.channel.Qos(c.qosCount, 0, false); err != nil {
		return nil, fmt.Errorf("error setting qos: %w", err)
	}

	if err = bindQueue(c.channel, queue.Name, c.routingKey, c.exchangeName); err != nil {
		return nil, fmt.Errorf("queue Bind: %w", err)
	}

	messages, err := c.channel.Consume(
		queue.Name,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("queue consume: %w", err)
	}

	return messages, nil
}

func (c *C) Close() error {
	if !c.isConnected {
		return nil
	}

	return c.conn.Close()
}
