package storage

import (
	"context"
	"os"
	"testing"
	pb "the_game_card_game/proto"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// This is a bit of a hack for testing.
	// In a real application, you'd use a proper configuration management system.
	if os.Getenv("POSTGRES_DSN") == "" {
		os.Setenv("POSTGRES_DSN", "postgres://user:password@localhost:5432/the_game")
	}
	if os.Getenv("REDIS_ADDR") == "" {
		os.Setenv("REDIS_ADDR", "localhost:6379")
	}

	// You must have Docker running with `docker-compose up -d` for this test to pass.
	os.Exit(m.Run())
}

func TestDatabaseConnections(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := NewStore(ctx, os.Getenv("REDIS_ADDR"), os.Getenv("POSTGRES_DSN"))
	require.NoError(t, err, "failed to connect to databases")
	defer store.Close()

	gameID := uuid.New().String()
	playerID := "player-test-123"

	// --- Test PostgreSQL ---
	t.Run("PostgreSQL", func(t *testing.T) {
		// Write
		err := store.CreateGame(ctx, gameID, playerID)
		require.NoError(t, err, "failed to create game in postgres")

		// Read
		isActive, err := store.GetGameForTest(ctx, gameID)
		require.NoError(t, err, "failed to get game from postgres")
		require.True(t, isActive, "game should be active")
	})

	// --- Test Redis ---
	t.Run("Redis", func(t *testing.T) {
		// Write
		initialState := &pb.GameState{
			GameId:    gameID,
			PlayerIds: []string{playerID},
			DeckSize:  98,
		}
		err := store.UpdateGameState(ctx, gameID, initialState)
		require.NoError(t, err, "failed to update game state in redis")

		// Read
		retrievedState, err := store.GetGameState(ctx, gameID)
		require.NoError(t, err, "failed to get game state from redis")
		require.Equal(t, gameID, retrievedState.GetGameId())
		require.Equal(t, int32(98), retrievedState.GetDeckSize())
	})
} 