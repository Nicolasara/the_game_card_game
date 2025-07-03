package main

import (
	"context"
	"flag"
	"log"
	"time"

	"the_game_card_game/pkg/bot"
	pb "the_game_card_game/proto"

	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server", "localhost:50051", "The server address in the format of host:port")
	strategy   = flag.String("strategy", "random", "The bot's strategy (e.g., random, minimal-jump)")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewGameServiceClient(conn)

	var botStrategy bot.Strategy
	switch *strategy {
	case "random":
		botStrategy = bot.NewRandomStrategy()
	case "minimal-jump":
		botStrategy = bot.NewMinimalJumpStrategy()
	case "safe-ten":
		botStrategy = bot.NewSafeTenStrategy()
	default:
		log.Fatalf("Unknown strategy: %s", *strategy)
	}

	// Create a new game
	createGameResp, err := client.CreateGame(context.Background(), &pb.CreateGameRequest{})
	if err != nil {
		log.Fatalf("could not create game: %v", err)
	}
	gameID := createGameResp.GameState.GameId
	playerID := createGameResp.GameState.PlayerIds[0]

	log.Printf("Game created with ID: %s, Player ID: %s", gameID, playerID)

	// Game loop
	for {
		// Get game state
		// In a real bot, you'd likely have a streaming connection, but for this simple one,
		// we'll just get the state at the beginning of our turn.
		stream, err := client.StreamGameState(context.Background(), &pb.StreamGameStateRequest{GameId: gameID})
		if err != nil {
			log.Fatalf("could not open stream: %v", err)
		}
		gameState, err := stream.Recv()
		if err != nil {
			log.Fatalf("could not receive game state: %v", err)
		}

		if gameState.GetGameOver() {
			log.Printf("Game is over: %s", gameState.GetMessage())
			break
		}

		// Check if it's our turn
		if gameState.CurrentTurnPlayerId != playerID {
			// It's not our turn, wait a bit
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// It's our turn, get the next move from the strategy
		playReq, endTurnReq, err := botStrategy.GetNextMove(playerID, gameState)
		if err != nil {
			log.Fatalf("strategy error: %v", err)
		}

		if playReq != nil {
			// Play a card
			_, err := client.PlayCard(context.Background(), playReq)
			if err != nil {
				log.Fatalf("could not play card: %v", err)
			}
			log.Printf("Played card: %v on pile %s", playReq.Card.Value, playReq.PileId)
		} else if endTurnReq != nil {
			// End the turn
			endTurnResp, err := client.EndTurn(context.Background(), endTurnReq)
			if err != nil {
				log.Fatalf("could not end turn: %v", err)
			}
			if !endTurnResp.Success {
				log.Printf("Could not end turn: %s. Stopping.", endTurnResp.Message)
				break
			}
			log.Printf("Ended turn")
		} else {
			log.Println("Strategy returned no move, ending turn by default.")
			_, err := client.EndTurn(context.Background(), &pb.EndTurnRequest{GameId: gameID, PlayerId: playerID})
			if err != nil {
				log.Fatalf("could not end turn: %v", err)
			}
		}

		// Small delay to make the game observable
		time.Sleep(500 * time.Millisecond)
	}
}
