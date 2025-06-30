//go:build integration

package server

import (
	"context"
	"os"
	"testing"
	"the_game_card_game/pkg/storage"
	pb "the_game_card_game/proto"
	"time"

	"github.com/stretchr/testify/require"
)

var testStore *storage.Store
var testServer *Server

func TestMain(m *testing.M) {
	// Set up environment variables for database connections
	if os.Getenv("POSTGRES_DSN") == "" {
		os.Setenv("POSTGRES_DSN", "postgres://user:password@localhost:5432/the_game")
	}
	if os.Getenv("REDIS_ADDR") == "" {
		os.Setenv("REDIS_ADDR", "localhost:6379")
	}

	// You must have Docker running with `docker-compose up -d` for these tests to pass.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testStore, err = storage.NewStore(ctx, os.Getenv("REDIS_ADDR"), os.Getenv("POSTGRES_DSN"))
	if err != nil {
		panic("failed to connect to databases for testing: " + err.Error())
	}
	defer testStore.Close()

	testServer = NewServer(testStore)

	// Run tests
	os.Exit(m.Run())
}

func TestGameLifecycle_Integration(t *testing.T) {
	ctx := context.Background()
	// 1. Create Game
	createReq := &pb.CreateGameRequest{PlayerId: "player1"}
	createRes, err := testServer.CreateGame(ctx, createReq)
	require.NoError(t, err)
	gameID := createRes.GameState.GameId

	// 2. Join Game
	joinReq := &pb.JoinGameRequest{GameId: gameID, PlayerId: "player2"}
	joinRes, err := testServer.JoinGame(ctx, joinReq)
	require.NoError(t, err)
	require.True(t, joinRes.Success)
	require.Len(t, joinRes.GameState.PlayerIds, 2, "There should be two players")
	require.Len(t, joinRes.GameState.Hands["player2"].Cards, 7, "Player 2 should have 7 cards")

	// 3. Play Card (Valid Move)
	hand := joinRes.GameState.Hands["player1"].Cards
	var cardToPlay *pb.Card
	for _, c := range hand {
		if c.Value > 1 {
			cardToPlay = c
			break
		}
	}
	require.NotNil(t, cardToPlay, "Could not find a valid card to play")

	playReq := &pb.PlayCardRequest{
		GameId:   gameID,
		PlayerId: "player1",
		Card:     cardToPlay,
		PileId:   "up1",
	}
	playRes, err := testServer.PlayCard(ctx, playReq)
	require.NoError(t, err)
	require.True(t, playRes.Success, "Valid card play should succeed")
}

func TestStreamGameState_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Create a game
	createRes, err := testServer.CreateGame(ctx, &pb.CreateGameRequest{PlayerId: "integration_stream_player"})
	require.NoError(t, err)
	gameID := createRes.GameState.GameId

	// 2. Directly subscribe to the Redis channel for this game
	channelName := "game-updates:" + gameID
	pubsub := testStore.Redis.Subscribe(ctx, channelName)

	// Wait for the subscription to be confirmed before proceeding.
	_, err = pubsub.Receive(ctx)
	require.NoError(t, err, "failed to subscribe to redis channel")
	defer pubsub.Close()

	// 3. Play a card, which should trigger a publish
	cardToPlay := createRes.GameState.Hands["integration_stream_player"].Cards[0]
	_, err = testServer.PlayCard(ctx, &pb.PlayCardRequest{
		GameId:   gameID,
		PlayerId: "integration_stream_player",
		Card:     cardToPlay,
		PileId:   "up1",
	})
	require.NoError(t, err)

	// 4. Assert that the message was received on the channel
	select {
	case msg, ok := <-pubsub.Channel():
		require.True(t, ok, "pubsub channel should be open")
		require.Equal(t, "update", msg.Payload)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for game update notification")
	}
}

// A minimal implementation of the server stream interface for testing purposes.
type testStreamGameStateServer struct {
	pb.GameService_StreamGameStateServer
	ctx context.Context
}

func (s *testStreamGameStateServer) Context() context.Context {
	return s.ctx
}

func (s *testStreamGameStateServer) Send(state *pb.GameState) error {
	// In a real test, you might send this to a channel for the main test goroutine to read.
	return nil
} 