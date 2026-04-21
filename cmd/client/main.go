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

	userGameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(userGameState),
	)
	if err != nil {
		fmt.Printf("there was an issue subscribing JSON: %v", err.Error())
	}

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
				continue
			}
			fmt.Printf("player %v successfully moved units %v to %v\n", cmdMove.Player, cmdMove.Units, cmdMove.ToLocation)
		case "status":
			userGameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			ctx.Done()

			fmt.Println("Shutting down gracefully...")

			gamelogic.PrintQuit()
			breakFromRepl = true
		default:
			fmt.Println("unknown command")
		}

		if breakFromRepl {
			_, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()
			break
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		gs.HandlePause(ps)
		defer fmt.Print("> ")
	}
}
