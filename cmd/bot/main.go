package main

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcPort = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	log.Println("Bot client connected successfully.")

	// The main bot loop will go here.
}
