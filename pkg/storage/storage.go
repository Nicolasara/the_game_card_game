package storage

import (
	"context"
	"fmt"
	pb "the_game_card_game/proto"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/protobuf/proto"
)

// Storer defines the interface for all database operations.
type Storer interface {
	CreateGame(ctx context.Context, gameID string, playerID string) error
	GetGameForTest(ctx context.Context, gameID string) (bool, error)
	GetGameState(ctx context.Context, gameID string) (*pb.GameState, error)
	UpdateGameState(ctx context.Context, gameID string, state *pb.GameState) error
	SaveMove(ctx context.Context, gameID string, playerID string, card int, pileID string) error
	PublishGameUpdate(ctx context.Context, gameID string) error
	SubscribeToGameUpdates(ctx context.Context, gameID string) (<-chan *redis.Message, func(), error)
	Close()
}

// Store holds the clients for our databases. It implements the Storer interface.
type Store struct {
	Redis *redis.Client
	DB    *pgxpool.Pool
}

// NewStore creates a new Store with connections to Redis and PostgreSQL.
func NewStore(ctx context.Context, redisURL, pgConnStr string) (*Store, error) {
	// Connect to Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Connect to PostgreSQL
	dbpool, err := pgxpool.Connect(ctx, pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Store{
		Redis: rdb,
		DB:    dbpool,
	}, nil
}

// Close closes the database connections.
func (s *Store) Close() {
	s.DB.Close()
	s.Redis.Close()
}

// --- Game Logic ---

// CreateGame inserts a new game into PostgreSQL.
func (s *Store) CreateGame(ctx context.Context, gameID string, playerID string) error {
	_, err := s.DB.Exec(ctx, "INSERT INTO games (game_id, is_active) VALUES ($1, $2)", gameID, true)
	if err != nil {
		return fmt.Errorf("failed to insert game: %w", err)
	}
	// We would also add the player to the players table here
	return nil
}

// GetGameForTest retrieves a game's active status from PostgreSQL for testing.
func (s *Store) GetGameForTest(ctx context.Context, gameID string) (bool, error) {
	var isActive bool
	err := s.DB.QueryRow(ctx, "SELECT is_active FROM games WHERE game_id = $1", gameID).Scan(&isActive)
	if err != nil {
		return false, fmt.Errorf("failed to get game for test: %w", err)
	}
	return isActive, nil
}

// GetGameState retrieves a game's state from Redis.
func (s *Store) GetGameState(ctx context.Context, gameID string) (*pb.GameState, error) {
	val, err := s.Redis.Get(ctx, fmt.Sprintf("game:%s", gameID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get game state from redis: %w", err)
	}

	state := &pb.GameState{}
	if err := proto.Unmarshal([]byte(val), state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game state: %w", err)
	}
	return state, nil
}

// UpdateGameState marshals a GameState to protobuf and saves it in Redis.
func (s *Store) UpdateGameState(ctx context.Context, gameID string, state *pb.GameState) error {
	data, err := proto.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	err = s.Redis.Set(ctx, fmt.Sprintf("game:%s", gameID), data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set game state in redis: %w", err)
	}
	return nil
}

func (s *Store) SaveMove(ctx context.Context, gameID string, playerID string, card int, pileID string) error {
	// TODO: Implement logic to save a move to PostgreSQL
	return nil
}

// --- Pub/Sub ---

// PublishGameUpdate sends a notification to a game's channel that it has been updated.
func (s *Store) PublishGameUpdate(ctx context.Context, gameID string) error {
	channel := fmt.Sprintf("game-updates:%s", gameID)
	// The message can be simple, its presence is the notification.
	return s.Redis.Publish(ctx, channel, "update").Err()
}

// SubscribeToGameUpdates subscribes to a game's update channel.
// It returns a channel for messages, a function to close the subscription, and an error.
func (s *Store) SubscribeToGameUpdates(ctx context.Context, gameID string) (<-chan *redis.Message, func(), error) {
	channel := fmt.Sprintf("game-updates:%s", gameID)
	pubsub := s.Redis.Subscribe(ctx, channel)

	// Wait for subscription to be confirmed.
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to game updates: %w", err)
	}

	ch := pubsub.Channel()
	closeFunc := func() {
		pubsub.Close()
	}

	return ch, closeFunc, nil
}
