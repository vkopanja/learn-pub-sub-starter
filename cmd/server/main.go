package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug,
		"", pubsub.Durable)
	if err != nil {
		fmt.Println("there was an issue binding to peril topic")
	}

	gamelogic.PrintClientHelp()
	for {
		input := gamelogic.GetInput()
		breakFromRepl := false

		switch input[0] {
		case "pause":
			fmt.Println("sending pause message")
			sendMessage(conn, true)
		case "resume":
			fmt.Println("sending resume message")
			sendMessage(conn, false)
		case "quit":
			fmt.Println("exiting...")
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			ctx.Done()

			fmt.Println("Shutting down gracefully...")

			_, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()
			breakFromRepl = true
		default:
			fmt.Println("unknown command")
		}

		if breakFromRepl {
			break
		}
	}
}

func sendMessage(conn *amqp.Connection, isPaused bool) {
	chn, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("There was an error opening the channel: %v", err))
	}

	playState := routing.PlayingState{
		IsPaused: isPaused,
	}

	pubsub.PublishJSON(chn, routing.ExchangePerilDirect, routing.PauseKey, playState)
}
