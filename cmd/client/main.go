package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "the_game_card_game/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// --- gRPC Connection ---
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewGameServiceClient(conn)

	// --- Command Line Flags ---
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go <command> [arguments]")
	}
	command := os.Args[1]
	fs := flag.NewFlagSet(command, flag.ExitOnError)

	var (
		gameID   = fs.String("game", "", "Game ID")
		playerID = fs.String("player", "default-player", "Player ID")
		cardVal  = fs.Int("card", 0, "Card value to play")
		pileID   = fs.String("pile", "", "Pile ID to play on (up1, up2, down1, down2)")
	)

	fs.Parse(os.Args[2:])

	// --- Command Dispatch ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch command {
	case "create":
		res, err := client.CreateGame(ctx, &pb.CreateGameRequest{PlayerId: *playerID})
		if err != nil {
			log.Fatalf("could not create game: %v", err)
		}
		log.Printf("Game created with ID: %s", res.GetGameState().GetGameId())
		fmt.Println("GameState:", res.GetGameState().String())

	case "join":
		if *gameID == "" {
			log.Fatal("-game flag is required for join command")
		}
		res, err := client.JoinGame(ctx, &pb.JoinGameRequest{GameId: *gameID, PlayerId: *playerID})
		if err != nil {
			log.Fatalf("could not join game: %v", err)
		}
		if res.Success {
			log.Printf("Successfully joined game %s", *gameID)
			fmt.Println("GameState:", res.GetGameState().String())
		} else {
			log.Printf("Failed to join game %s", *gameID)
		}

	case "play":
		if *gameID == "" || *playerID == "" || *cardVal == 0 || *pileID == "" {
			log.Fatal("-game, -player, -card, and -pile flags are required for play command")
		}
		res, err := client.PlayCard(ctx, &pb.PlayCardRequest{
			GameId:   *gameID,
			PlayerId: *playerID,
			Card:     &pb.Card{Value: int32(*cardVal)},
			PileId:   *pileID,
		})
		if err != nil {
			log.Fatalf("could not play card: %v", err)
		}
		if res.Success {
			log.Printf("Card %d played successfully on pile %s", *cardVal, *pileID)
		} else {
			log.Printf("Failed to play card: %s", res.Message)
		}

	case "stream":
		if *gameID == "" {
			log.Fatal("-game flag is required for stream command")
		}
		stream, err := client.StreamGameState(context.Background(), &pb.StreamGameStateRequest{GameId: *gameID})
		if err != nil {
			log.Fatalf("could not stream game state: %v", err)
		}
		log.Printf("Streaming game state for game %s...", *gameID)
		for {
			state, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error receiving stream: %v", err)
			}
			log.Println("--- Game State Update ---")
			log.Println(state.String())
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
} 