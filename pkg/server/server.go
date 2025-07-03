package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"the_game_card_game/pkg/game"
	"the_game_card_game/pkg/logger"
	"the_game_card_game/pkg/storage"
	pb "the_game_card_game/proto"

	"github.com/google/uuid"
)

type Server struct {
	pb.UnimplementedGameServiceServer
	store  storage.Storer
	logger *logger.Logger
}

func NewServer(store storage.Storer, logger *logger.Logger) *Server {
	return &Server{store: store, logger: logger}
}

// GameEvent is a generic struct for all game log events.
type GameEvent struct {
	Timestamp time.Time   `json:"timestamp"`
	GameID    string      `json:"game_id"`
	EventType string      `json:"event_type"`
	Payload   interface{} `json:"payload"`
}

// GameStartEventPayload contains the data for a 'game_start' event.
type GameStartEventPayload struct {
	PlayerID string `json:"player_id"`
}

// PlayCardEventPayload contains the data for a 'play_card' event.
type PlayCardEventPayload struct {
	PlayerID  string `json:"player_id"`
	CardValue int32  `json:"card_value"`
	PileID    string `json:"pile_id"`
}

// EndTurnEventPayload contains the data for an 'end_turn' event.
type EndTurnEventPayload struct {
	PlayerID string `json:"player_id"`
}

// GameOverEventPayload contains the data for a 'game_over' event.
type GameOverEventPayload struct {
	Winner  string `json:"winner,omitempty"`
	Message string `json:"message"`
}

func (s *Server) logEvent(gameID, eventType string, payload interface{}) {
	event := GameEvent{
		Timestamp: time.Now(),
		GameID:    gameID,
		EventType: eventType,
		Payload:   payload,
	}
	if err := s.logger.Log(event); err != nil {
		log.Printf("failed to log event: %v", err)
	}
}

func (s *Server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	log.Printf("CreateGame request received for player: %s", req.GetPlayerId())

	gameID := uuid.New().String()
	playerID := req.GetPlayerId()
	if playerID == "" {
		playerID = uuid.New().String()
		log.Printf("Player ID was not provided, generated a new one: %s", playerID)
	}

	s.logEvent(gameID, "game_start", GameStartEventPayload{
		PlayerID: playerID,
	})

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

	// It must be the player's turn to play.
	if state.CurrentTurnPlayerId != req.GetPlayerId() {
		msg := fmt.Sprintf("it is not your turn (current turn: %s)", state.CurrentTurnPlayerId)
		log.Printf("%s", msg)
		return &pb.PlayCardResponse{Success: false, Message: msg}, nil
	}

	// Apply the move using the game logic
	newState, err := game.PlayCard(state, req.GetPlayerId(), req.GetCard().GetValue(), req.GetPileId())
	if err != nil {
		log.Printf("invalid move: %v", err)
		return &pb.PlayCardResponse{Success: false, Message: err.Error()}, nil
	}

	s.logEvent(req.GetGameId(), "play_card", PlayCardEventPayload{
		PlayerID:  req.GetPlayerId(),
		CardValue: req.GetCard().GetValue(),
		PileID:    req.GetPileId(),
	})

	if newState.GetGameOver() {
		s.logEvent(req.GetGameId(), "game_over", GameOverEventPayload{
			Message: newState.GetMessage(),
		})
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

func (s *Server) EndTurn(ctx context.Context, req *pb.EndTurnRequest) (*pb.EndTurnResponse, error) {
	log.Printf("EndTurn request received for game %s by player %s", req.GetGameId(), req.GetPlayerId())

	state, err := s.store.GetGameState(ctx, req.GetGameId())
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Validate that it's the correct player's turn.
	if state.CurrentTurnPlayerId != req.GetPlayerId() {
		msg := fmt.Sprintf("it is not your turn (current turn: %s)", state.CurrentTurnPlayerId)
		return &pb.EndTurnResponse{Success: false, Message: msg}, nil
	}

	newState, err := game.EndTurn(state, req.GetPlayerId())
	if err != nil {
		log.Printf("invalid end turn for player %s: %v", req.GetPlayerId(), err)
		return &pb.EndTurnResponse{Success: false, Message: err.Error()}, nil
	}

	s.logEvent(req.GetGameId(), "end_turn", EndTurnEventPayload{
		PlayerID: req.GetPlayerId(),
	})

	if newState.GetGameOver() {
		s.logEvent(req.GetGameId(), "game_over", GameOverEventPayload{
			Winner:  newState.GetWinner(),
			Message: newState.GetMessage(),
		})
	}

	// Update state in Redis
	if err := s.store.UpdateGameState(ctx, req.GetGameId(), newState); err != nil {
		return nil, fmt.Errorf("failed to update game state: %w", err)
	}

	// Notify subscribers
	if err := s.store.PublishGameUpdate(ctx, req.GetGameId()); err != nil {
		log.Printf("failed to publish game update: %v", err) // Non-critical
	}

	return &pb.EndTurnResponse{Success: true, GameState: newState}, nil
}

func (s *Server) StreamGameState(req *pb.StreamGameStateRequest, stream pb.GameService_StreamGameStateServer) error {
	log.Printf("StreamGameState request received for game %s", req.GetGameId())
	ctx := stream.Context()

	// Subscribe to updates first to avoid race conditions.
	ch, closeSub, err := s.store.SubscribeToGameUpdates(ctx, req.GetGameId())
	if err != nil {
		log.Printf("failed to subscribe to game updates: %v", err)
		return err
	}
	defer closeSub()

	// After subscribing, get and send the current state.
	// This ensures we don't miss an update that happens between the initial get and the subscribe.
	initialState, err := s.store.GetGameState(ctx, req.GetGameId())
	if err != nil {
		return err
	}
	if err := stream.Send(initialState); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Client for game %s disconnected", req.GetGameId())
			return nil
		case msg := <-ch:
			// The message content doesn't matter, its arrival is the signal.
			log.Printf("Received update for game %s via channel %s, sending new state", req.GetGameId(), msg.Channel)
			newState, err := s.store.GetGameState(ctx, req.GetGameId())
			if err != nil {
				log.Printf("Error getting new state for game %s: %v", req.GetGameId(), err)
				continue
			}
			if err := stream.Send(newState); err != nil {
				log.Printf("Error sending new state for game %s: %v", req.GetGameId(), err)
				return err // Client likely disconnected
			}
		}
	}
}
