package main

import (
	"context"
	"log"
	"time"

	pb "the_game_card_game/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGameServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CreateGame(ctx, &pb.CreateGameRequest{Name: "My New Game"})
	if err != nil {
		log.Fatalf("could not create game: %v", err)
	}
	log.Printf("Game created with ID: %s and Name: %s", r.GetId(), r.GetName())
} 