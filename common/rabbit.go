package common

import (
	"fmt"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQ struct {
	conn *amqp.Connection
	Ch   *amqp.Channel
}

var logger *zap.SugaredLogger

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	logger = GetLogger()

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{conn, ch}, nil
}

func (r *RabbitMQ) Consume(callback func([]byte) bool) error {
	// Consume messages from the queue
	msgs, err := r.Ch.Consume(
		"cowboy-queue", // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		if callback(msg.Body) {
			msg.Ack(false)
		} else {
			msg.Nack(false, true)
		}
	}

	return nil
}

func (r *RabbitMQ) DeclareQueue(name string) error {
	_, err := r.Ch.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}
	return nil
}

func (r *RabbitMQ) Close() {
	r.Ch.Close()
	r.conn.Close()
}
