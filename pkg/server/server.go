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
	log.Printf("CreateGame request received: %v", req)
	return &pb.CreateGameResponse{
		Id:   "some-uuid",
		Name: req.GetName(),
	}, nil
}