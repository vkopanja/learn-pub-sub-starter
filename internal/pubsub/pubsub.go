// Package pubsub is used for pub-sub operations with RabbitMQ
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	byteVal, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("error marshaling value to []byte: %w", err)
	}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        byteVal,
	})

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	chn, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("there was an issue creating channel: %w", err)
	}

	isDurable := queueType == Durable
	isTransient := queueType == Transient
	queue, err := chn.QueueDeclare(queueName, isDurable, isTransient, isTransient, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("issue with queue declaration: %w", err)
	}

	err = chn.QueueBind(queueName, key, exchange, true, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed binding queue to exchange: %w", err)
	}

	return chn, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, Durable)
	if err != nil {
		return fmt.Errorf("there was an issues with subscribing JSON: %v", err.Error())
	}
	if queue.Name == "" {
		return fmt.Errorf("queue name was nil, queue not bound")
	}

	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("there was an issue consuming channel from queue %v, err: %v",
			queueName, err.Error())
	}

	go func() {
		for msg := range msgs {
			var body T
			json.Unmarshal(msg.Body, &body)
			handler(body)
			msg.Ack(false)
		}
	}()

	return nil
}
