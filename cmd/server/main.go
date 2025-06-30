package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"the_game_card_game/pkg/server"
	"the_game_card_game/pkg/storage"
	pb "the_game_card_game/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcPort = ":50051"
	httpPort = ":8080"
)

func main() {
	// --- Boilerplate for graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// --- Database and Server Initialization ---
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		postgresDSN = "postgres://user:password@localhost:5432/the_game?sslmode=disable"
	}

	store, err := storage.NewStore(ctx, redisAddr, postgresDSN)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gameServer := server.NewServer(store)

	// --- gRPC Server ---
	go func() {
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterGameServiceServer(s, gameServer)
		log.Printf("gRPC server listening on %s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// --- gRPC-Gateway (HTTP Server) ---
	go func() {
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		grpcEndpoint := "localhost" + grpcPort
		err := pb.RegisterGameServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			log.Fatalf("failed to register gRPC gateway: %v", err)
		}

		log.Printf("HTTP gateway listening on %s", httpPort)
		if err := http.ListenAndServe(httpPort, mux); err != nil {
			log.Fatalf("failed to serve HTTP gateway: %v", err)
		}
	}()

	// --- Wait for shutdown signal ---
	<-sigChan
	log.Println("Shutting down servers...")
	// Add a small delay to allow for in-flight requests to complete
	time.Sleep(2 * time.Second)
} 