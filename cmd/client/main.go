package main

import (
	"context"
	"log"
	"time"

	pb "the_game_card_game/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGameServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CreateGame(ctx, &pb.CreateGameRequest{PlayerId: "Nico"})
	if err != nil {
		log.Fatalf("could not create game: %v", err)
	}
	log.Printf("Game created with ID: %s", r.GetGameState().GetGameId())
} 