package common

import (
	"fmt"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
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

func (r *RabbitMQ) Consume(queueName string, callback func([]byte) bool) error {
	msgs, err := r.ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
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
	_, err := r.ch.QueueDeclare(
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

func (r *RabbitMQ) PublishJSON(exchange string, queue string, json []byte) error {
	return r.ch.Publish(
		exchange, // exchange
		queue,    // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		},
	)
}

func (r *RabbitMQ) GetQueueInfo(queue string) (amqp.Queue, error) {
	return r.ch.QueueInspect(queue)
}

func (r *RabbitMQ) PurgeQueue(queue string) (int, error) {
	return r.ch.QueuePurge(queue, false)
}

func (r *RabbitMQ) Close() {
	r.ch.Close()
	r.conn.Close()
}
