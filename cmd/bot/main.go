package main

import (
	"context"
	"flag"
	"log"

	"the_game_card_game/pkg/bot"
	pb "the_game_card_game/proto"

	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server", "localhost:50051", "The server address in the format of host:port")
	strategy   = flag.String("strategy", "random", "The bot's strategy (e.g., random, minimal-jump)")
	numGames   = flag.Int("num_games", 1, "The number of games the bot should play")
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

	for i := 0; i < *numGames; i++ {
		log.Printf("--- Starting Game %d of %d ---", i+1, *numGames)
		playGame(client, botStrategy, *strategy)
	}
}

func playGame(client pb.GameServiceClient, botStrategy bot.Strategy, strategyName string) {
	// Create a new game
	createGameResp, err := client.CreateGame(context.Background(), &pb.CreateGameRequest{})
	if err != nil {
		log.Printf("could not create game: %v", err)
		return
	}
	gameID := createGameResp.GameState.GameId
	playerID := createGameResp.GameState.PlayerIds[0]

	log.Printf("Game created with ID: %s, Player ID: %s", gameID, playerID)

	// Join the game
	joinRes, err := client.JoinGame(context.Background(), &pb.JoinGameRequest{
		GameId:   gameID,
		PlayerId: playerID,
		Strategy: strategyName,
	})
	if err != nil {
		log.Printf("could not join game: %v", err)
		return
	}
	log.Printf("Successfully joined game: %v", joinRes.GetSuccess())

	// Game loop
	for {
		// Get game state
		// In a real bot, you'd likely have a streaming connection, but for this simple one,
		// we'll just get the state at the beginning of our turn.
		stream, err := client.StreamGameState(context.Background(), &pb.StreamGameStateRequest{GameId: gameID})
		if err != nil {
			log.Printf("could not open stream: %v", err)
			return
		}
		gameState, err := stream.Recv()
		if err != nil {
			log.Printf("could not receive game state: %v", err)
			return
		}

		if gameState.GetGameOver() {
			log.Printf("Game is over: %s", gameState.GetMessage())
			break
		}

		// Check if it's our turn
		if gameState.CurrentTurnPlayerId != playerID {
			continue
		}

		// It's our turn, get the next move from the strategy
		playReq, endTurnReq, err := botStrategy.GetNextMove(playerID, gameState)
		if err != nil {
			log.Printf("strategy error: %v", err)
			return
		}

		if playReq != nil {
			// Play a card
			_, err := client.PlayCard(context.Background(), playReq)
			if err != nil {
				log.Printf("could not play card: %v", err)
				return
			}
			log.Printf("Played card: %v on pile %s", playReq.Card.Value, playReq.PileId)
		} else if endTurnReq != nil {
			// End the turn
			endTurnResp, err := client.EndTurn(context.Background(), endTurnReq)
			if err != nil {
				log.Printf("could not end turn: %v", err)
				return
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
				log.Printf("could not end turn: %v", err)
				return
			}
		}
	}
}
