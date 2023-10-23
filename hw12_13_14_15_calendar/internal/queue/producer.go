package queue

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

type Producer interface {
	Connect() error
	Close() error
	Publish(json string) error
}

type P struct {
	uri          string
	exchangeName string
	exchangeType string
	routingKey   string
	queueName    string
	qosCount     int
	conn         *amqp.Connection
	channel      *amqp.Channel
	isConnected  bool
}

func NewProducer(host string, port int, user, password, exchangeName, exchangeType, routingKey, queueName string,
	qosCount int,
) *P {
	return &P{
		uri:          fmt.Sprintf("amqp://%s:%s@%s:%d/", user, password, host, port),
		exchangeName: exchangeName,
		exchangeType: exchangeType,
		routingKey:   routingKey,
		queueName:    queueName,
		qosCount:     qosCount,
		isConnected:  false,
	}
}

func (p *P) Connect() error {
	var err error
	if p.isConnected {
		return nil
	}

	p.conn, err = amqp.Dial(p.uri)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	p.channel, err = p.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}

	if err = declareExchange(p.channel, p.exchangeName, p.exchangeType); err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	queue, err := declareQueue(p.channel, p.queueName)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}

	if err = p.channel.Qos(p.qosCount, 0, false); err != nil {
		return fmt.Errorf("error setting qos: %w", err)
	}

	if err = bindQueue(p.channel, queue.Name, p.routingKey, p.exchangeName); err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	p.isConnected = true

	return nil
}

func (p *P) Close() error {
	if !p.isConnected {
		return nil
	}

	return p.conn.Close()
}

func (p *P) Publish(json string) error {
	if !p.isConnected {
		return errors.New("not connected")
	}

	if err := p.channel.Publish(
		p.exchangeName,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            []byte(json),
			DeliveryMode:    amqp.Persistent,
			Priority:        0,
		},
	); err != nil {
		return fmt.Errorf("exchange publish: %w", err)
	}

	return nil
}
