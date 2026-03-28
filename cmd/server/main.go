package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connStr := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed RabbitMQ connection: %v", err.Error()))
	}
	defer conn.Close()

	fmt.Println("RabbitMQ connection successful!")

	chn, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("There was an error opening the channel: %v", err))
	}

	playState := routing.PlayingState{
		IsPaused: true,
	}

	pubsub.PublishJSON(chn, routing.ExchangePerilDirect, routing.PauseKey, playState)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()

	fmt.Println("Shutting down gracefully...")

	_, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
}
