package server

import (
	"context"
	"testing"
	"the_game_card_game/pkg/storage/mocks"
	pb "the_game_card_game/proto"

	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock Stream for testing
type mockStream struct {
	pb.GameService_StreamGameStateServer
	ctx     context.Context
	recv    chan *pb.GameState
	sent    chan struct{} // Signal that a message was sent
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) Send(state *pb.GameState) error {
	m.recv <- state
	if m.sent != nil {
		m.sent <- struct{}{}
	}
	return nil
}

func TestCreateGame_Unit(t *testing.T) {
	// 1. Setup
	mockStore := new(mocks.Storer)
	server := NewServer(mockStore)
	ctx := context.Background()
	req := &pb.CreateGameRequest{PlayerId: "unit-tester"}

	// 2. Define Mock Expectations
	mockStore.On("CreateGame", mock.Anything, mock.AnythingOfType("string"), "unit-tester").Return(nil)
	mockStore.On("UpdateGameState", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*proto.GameState")).Return(nil)

	// 3. Execute
	res, err := server.CreateGame(ctx, req)

	// 4. Assert
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "unit-tester", res.GameState.PlayerIds[0])
	require.Len(t, res.GameState.Hands["unit-tester"].Cards, 8)
	mockStore.AssertExpectations(t)
}

func TestJoinGame_Unit(t *testing.T) {
	// 1. Setup
	mockStore := new(mocks.Storer)
	server := NewServer(mockStore)
	ctx := context.Background()
	gameID := "game-to-join"
	req := &pb.JoinGameRequest{GameId: gameID, PlayerId: "player2"}

	// Original state that the mock will return
	originalState := &pb.GameState{
		GameId:    gameID,
		PlayerIds: []string{"player1"},
		Deck:      make([]*pb.Card, 20), // A deck with 20 cards
		Hands:     make(map[string]*pb.Hand),
	}

	// 2. Define Mock Expectations
	mockStore.On("GetGameState", mock.Anything, gameID).Return(originalState, nil)
	mockStore.On("UpdateGameState", mock.Anything, gameID, mock.AnythingOfType("*proto.GameState")).Return(nil)

	// 3. Execute
	res, err := server.JoinGame(ctx, req)

	// 4. Assert
	require.NoError(t, err)
	require.True(t, res.Success)
	require.Len(t, res.GameState.PlayerIds, 2, "Should now have two players")
	require.Equal(t, "player2", res.GameState.PlayerIds[1])
	require.Len(t, res.GameState.Hands["player2"].Cards, 7, "Player 2 should have been dealt 7 cards")

	mockStore.AssertExpectations(t)
}

func TestPlayCard_Unit_Valid(t *testing.T) {
	// 1. Setup
	mockStore := new(mocks.Storer)
	server := NewServer(mockStore)
	ctx := context.Background()
	gameID := "game-to-play-in"
	playerID := "player1"
	cardToPlay := &pb.Card{Value: 15}

	// Initial state that the mock will return
	initialState := &pb.GameState{
		GameId:              gameID,
		PlayerIds:           []string{playerID},
		CurrentTurnPlayerId: playerID,
		Piles: map[string]*pb.Pile{
			"up1": {Ascending: true, Cards: []*pb.Card{{Value: 10}}},
		},
		Hands: map[string]*pb.Hand{
			playerID: {Cards: []*pb.Card{cardToPlay, {Value: 25}}},
		},
	}

	req := &pb.PlayCardRequest{GameId: gameID, PlayerId: playerID, Card: cardToPlay, PileId: "up1"}

	// 2. Define Mock Expectations
	mockStore.On("GetGameState", mock.Anything, gameID).Return(initialState, nil)
	mockStore.On("UpdateGameState", mock.Anything, gameID, mock.AnythingOfType("*proto.GameState")).Return(nil)
	mockStore.On("PublishGameUpdate", mock.Anything, gameID).Return(nil)
	mockStore.On("SaveMove", mock.Anything, gameID, playerID, 15, "up1").Return(nil)

	// 3. Execute
	res, err := server.PlayCard(ctx, req)

	// 4. Assert
	require.NoError(t, err)
	require.True(t, res.Success)

	// Allow the SaveMove goroutine to execute
	time.Sleep(50 * time.Millisecond)

	mockStore.AssertExpectations(t)
}

func TestStreamGameState_Unit(t *testing.T) {
	mockStore := mocks.NewStorer(t)
	testServer := NewServer(mockStore)
	gameID := "stream-test-game"

	// 1. Setup mock stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockSrv := &mockStream{
		ctx:     ctx,
		recv:    make(chan *pb.GameState, 1),
		sent:    make(chan struct{}, 1),
	}

	// 2. Setup mock pubsub channel
	mockPubSubChan := make(chan *redis.Message, 1)
	cleanupFunc := func() {
		close(mockPubSubChan)
	}

	// 3. Mock expectations
	initialState := &pb.GameState{GameId: gameID}
	mockStore.On("SubscribeToGameUpdates", mock.Anything, gameID).Return((<-chan *redis.Message)(mockPubSubChan), cleanupFunc, nil)
	mockStore.On("GetGameState", mock.Anything, gameID).Return(initialState, nil).Once() // For initial send

	// 4. Run StreamGameState in a goroutine
	streamErrChan := make(chan error, 1)
	go func() {
		streamErrChan <- testServer.StreamGameState(&pb.StreamGameStateRequest{GameId: gameID}, mockSrv)
	}()

	// 5. Verify initial state is sent
	select {
	case <-mockSrv.sent:
		state := <-mockSrv.recv
		require.Equal(t, gameID, state.GameId)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for initial state")
	}

	// 6. Simulate a game update
	updatedState := &pb.GameState{GameId: gameID, DeckSize: 50}
	mockStore.On("GetGameState", mock.Anything, gameID).Return(updatedState, nil).Once() // For update send
	mockPubSubChan <- &redis.Message{Channel: "game-updates:" + gameID, Payload: "update"}

	// 7. Verify the update is sent
	select {
	case <-mockSrv.sent:
		state := <-mockSrv.recv
		require.Equal(t, int32(50), state.DeckSize)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for update")
	}

	// 8. Cancel the context to close the stream
	cancel()
	select {
	case err := <-streamErrChan:
		require.NoError(t, err, "Stream should close gracefully")
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for stream to close")
	}

	mockStore.AssertExpectations(t)
} 