package server

import (
	"context"
	"log"

	"the_game_card_game/pkg/game"
	"the_game_card_game/pkg/storage"
	pb "the_game_card_game/proto"

	"github.com/google/uuid"
)

type Server struct {
	pb.UnimplementedGameServiceServer
	store *storage.Store
}

func NewServer(store *storage.Store) *Server {
	return &Server{store: store}
}

func (s *Server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	log.Printf("CreateGame request received for player: %s", req.GetPlayerId())

	gameID := uuid.New().String()
	playerID := req.GetPlayerId()

	// Create the initial game state using the game logic package
	initialState := game.NewGame(gameID, playerID)

	// Persist to PostgreSQL
	if err := s.store.CreateGame(ctx, gameID, playerID); err != nil {
		log.Printf("failed to create game in postgres: %v", err)
		return nil, err
	}

	// Persist to Redis
	if err := s.store.UpdateGameState(ctx, gameID, initialState); err != nil {
		log.Printf("failed to update game state in redis: %v", err)
		return nil, err
	}

	return &pb.CreateGameResponse{
		GameState: initialState,
	}, nil
}

func (s *Server) JoinGame(ctx context.Context, req *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	log.Printf("JoinGame request received for game %s by player %s", req.GetGameId(), req.GetPlayerId())

	// Get current game state from Redis
	state, err := s.store.GetGameState(ctx, req.GetGameId())
	if err != nil {
		log.Printf("failed to get game state: %v", err)
		return &pb.JoinGameResponse{Success: false}, err
	}

	// Add the new player to the game
	// TODO: Make hand size dynamic based on number of players
	const handSize = 7
	newState, err := game.AddPlayer(state, req.GetPlayerId(), handSize)
	if err != nil {
		log.Printf("failed to add player: %v", err)
		return &pb.JoinGameResponse{Success: false}, err
	}

	// Update the game state in Redis
	if err := s.store.UpdateGameState(ctx, req.GetGameId(), newState); err != nil {
		log.Printf("failed to update game state: %v", err)
		return &pb.JoinGameResponse{Success: false}, err
	}

	return &pb.JoinGameResponse{Success: true, GameState: newState}, nil
}

func (s *Server) PlayCard(ctx context.Context, req *pb.PlayCardRequest) (*pb.PlayCardResponse, error) {
	log.Printf("PlayCard request received for game %s by player %s", req.GetGameId(), req.GetPlayerId())

	// Get current game state from Redis
	state, err := s.store.GetGameState(ctx, req.GetGameId())
	if err != nil {
		log.Printf("failed to get game state: %v", err)
		return &pb.PlayCardResponse{Success: false, Message: "Game not found"}, err
	}

	// Apply the move using the game logic
	newState, err := game.PlayCard(state, req.GetPlayerId(), req.GetCard().GetValue(), req.GetPileId())
	if err != nil {
		log.Printf("invalid move: %v", err)
		return &pb.PlayCardResponse{Success: false, Message: err.Error()}, nil
	}

	// Update the game state in Redis
	if err := s.store.UpdateGameState(ctx, req.GetGameId(), newState); err != nil {
		log.Printf("failed to update game state: %v", err)
		return &pb.PlayCardResponse{Success: false, Message: "Failed to save game state"}, err
	}

	// Notify subscribers that the game state has changed.
	if err := s.store.PublishGameUpdate(ctx, req.GetGameId()); err != nil {
		// This is not a critical error, so we just log it.
		// The game state was updated, but clients won't get the real-time push.
		log.Printf("failed to publish game update: %v", err)
	}

	// Asynchronously save the move to PostgreSQL
	go func() {
		err := s.store.SaveMove(context.Background(), req.GetGameId(), req.GetPlayerId(), int(req.GetCard().GetValue()), req.GetPileId())
		if err != nil {
			log.Printf("failed to save move to postgres: %v", err)
		}
	}()

	return &pb.PlayCardResponse{Success: true}, nil
}

func (s *Server) StreamGameState(req *pb.StreamGameStateRequest, stream pb.GameService_StreamGameStateServer) error {
	log.Printf("StreamGameState request received for game %s", req.GetGameId())
	ctx := stream.Context()

	// Immediately send the current state
	initialState, err := s.store.GetGameState(ctx, req.GetGameId())
	if err != nil {
		return err
	}
	if err := stream.Send(initialState); err != nil {
		return err
	}

	// Subscribe to future updates
	pubsub := s.store.SubscribeToGameUpdates(ctx, req.GetGameId())
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Client for game %s disconnected", req.GetGameId())
			return nil
		case <-ch:
			log.Printf("Received update for game %s, sending new state", req.GetGameId())
			newState, err := s.store.GetGameState(ctx, req.GetGameId())
			if err != nil {
				log.Printf("Error getting new state for game %s: %v", req.GetGameId(), err)
				// Decide if we should continue or terminate
				continue
			}
			if err := stream.Send(newState); err != nil {
				log.Printf("Error sending new state for game %s: %v", req.GetGameId(), err)
				return err // Client likely disconnected
			}
		}
	}
}