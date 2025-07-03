package main

import (
	"context"
	"flag"
	"log"
	"time"

	"the_game_card_game/pkg/bot"
	pb "the_game_card_game/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	playerID := flag.String("player", "bot-player", "The ID of the player")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGameServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 1. Create a game
	createRes, err := client.CreateGame(ctx, &pb.CreateGameRequest{PlayerId: *playerID})
	if err != nil {
		log.Fatalf("could not create game: %v", err)
	}
	log.Printf("Game created with ID: %s", createRes.GameState.GameId)
	gameState := createRes.GameState

	strategy := bot.NewRandomStrategy()

	// 2. Play two cards
	for i := 0; i < 2; i++ {
		playReq, _, err := strategy.GetNextMove(*playerID, gameState)
		if err != nil {
			log.Fatalf("strategy error: %v", err)
		}
		if playReq == nil {
			log.Println("No more moves to play, ending turn early.")
			break
		}

		log.Printf("Playing card %d on pile %s", playReq.Card.Value, playReq.PileId)
		_, err = client.PlayCard(ctx, playReq)
		if err != nil {
			log.Fatalf("could not play card: %v", err)
		}

		// We need to get the updated game state to make the next move.
		// For this simple test, we'll just stream once. A real bot would stream continuously.
		stream, err := client.StreamGameState(ctx, &pb.StreamGameStateRequest{GameId: gameState.GameId})
		if err != nil {
			log.Fatalf("could not stream game state: %v", err)
		}
		gameState, err = stream.Recv()
		if err != nil {
			log.Fatalf("could not receive game state: %v", err)
		}
	}

	// 3. End the turn
	log.Println("Ending turn.")
	_, err = client.EndTurn(ctx, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: *playerID})
	if err != nil {
		log.Fatalf("could not end turn: %v", err)
	}

	log.Println("Bot has finished its turn.")
}
