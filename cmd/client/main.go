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
	fmt.Println("Starting Peril client...")

	connStr := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed RabbitMQ connection: %v", err.Error()))
	}
	defer conn.Close()

	fmt.Println("RabbitMQ connection successful!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		panic(fmt.Sprintf("There was an issue with fetching username: %v", err.Error()))
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Printf("there was an issue declaring and binding queue: %s", err.Error())
	}

	userGameState := gamelogic.NewGameState(username)
	for {
		input := gamelogic.GetInput()
		breakFromRepl := false

		switch input[0] {
		case "spawn":
			fmt.Println("sending spawn message")
			err := userGameState.CommandSpawn(input)
			if err != nil {
				fmt.Printf("error spawning %v\n", input[1])
			}
		case "move":
			fmt.Println("sending move message")
			cmdMove, err := userGameState.CommandMove(input)
			if err != nil {
				fmt.Printf("error moving unit %v to %v\n", input[2], input[1])
			}
			fmt.Printf("player %v successfully moved units %v to %v\n", cmdMove.Player, cmdMove.Units, cmdMove.ToLocation)
		case "status":
			userGameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			fmt.Println("exiting...")
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			ctx.Done()

			gamelogic.PrintQuit()

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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()

	fmt.Println("Shutting down gracefully...")

	_, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
}
