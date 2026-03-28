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

	return chn, queue, nil
}
