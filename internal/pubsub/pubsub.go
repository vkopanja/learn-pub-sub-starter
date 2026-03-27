package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	byteVal, err := json.Marshal(val)
	if err != nil {
		fmt.Errorf("Error marshaling value to []byte: %w", err)
		return err
	}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        byteVal,
	})

	return nil
}
