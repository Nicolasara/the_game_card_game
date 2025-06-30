package server

import (
	"context"
	"log"

	pb "the_game_card_game/proto"
)

type Server struct {
	pb.UnimplementedGameServiceServer
}

func (s *Server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	log.Printf("CreateGame request received for player: %s", req.GetPlayerId())

	// TODO: Implement actual game creation logic here.
	// For now, return a placeholder response.
	return &pb.CreateGameResponse{
		GameState: &pb.GameState{
			GameId:    "new-game-123",
			PlayerIds: []string{req.GetPlayerId()},
		},
	}, nil
}