package queue

import (
	"github.com/streadway/amqp"
)

func declareExchange(ch *amqp.Channel, eName, eType string) error {
	return ch.ExchangeDeclare(
		eName,
		eType,
		true,
		false,
		false,
		false,
		nil,
	)
}

func declareQueue(ch *amqp.Channel, qName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		qName,
		true,
		false,
		false,
		false,
		nil,
	)
}

func bindQueue(ch *amqp.Channel, qName, eKey, eName string) error {
	return ch.QueueBind(
		qName,
		eKey,
		eName,
		false,
		nil,
	)
}
