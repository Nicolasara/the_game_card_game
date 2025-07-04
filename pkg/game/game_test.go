package game

import (
	"testing"
	pb "the_game_card_game/proto"

	"github.com/stretchr/testify/require"
)

func TestPlayCard_LosingCondition(t *testing.T) {
	// 1. Setup
	playerID := "stuck-player"
	initialState := &pb.GameState{
		GameId:              "game-with-no-moves",
		PlayerIds:           []string{playerID},
		CurrentTurnPlayerId: playerID,
		Piles: map[string]*pb.Pile{
			// After playing 50, the only remaining card '4' will have no valid pile.
			"up1":   {Ascending: true, Cards: []*pb.Card{{Value: 48}}},
			"up2":   {Ascending: true, Cards: []*pb.Card{{Value: 98}}},
			"down1": {Ascending: false, Cards: []*pb.Card{{Value: 6}}},
			"down2": {Ascending: false, Cards: []*pb.Card{{Value: 7}}},
		},
		Hands: map[string]*pb.Hand{
			playerID: {Cards: []*pb.Card{{Value: 50}, {Value: 8}}},
		},
		CardsPlayedThisTurn: 0,
	}

	// 2. Execute
	// Player plays their first card, which is a valid move.
	newState, err := PlayCard(initialState, playerID, 50, "up1")

	// 3. Assert
	require.NoError(t, err)
	require.NotNil(t, newState)

	// The key assertion: The game should now be over.
	require.True(t, newState.GameOver, "Game should be over because no second move is possible")
	require.Contains(t, newState.Message, "lost: No more valid moves")

	// The hand should have one card left.
	require.Len(t, newState.Hands[playerID].Cards, 1, "Player should have one card left in hand")
}

func TestEndTurn_WinCondition(t *testing.T) {
	// 1. Setup
	playerID := "victorious-player"
	initialState := &pb.GameState{
		GameId:              "winnable-game",
		PlayerIds:           []string{playerID},
		CurrentTurnPlayerId: playerID,
		Deck:                []*pb.Card{}, // Deck is empty
		DeckSize:            0,
		Piles: map[string]*pb.Pile{
			"up1": {Cards: []*pb.Card{{Value: 50}}},
		},
		Hands: map[string]*pb.Hand{
			// The player has played all but their last two cards
			playerID: {Cards: []*pb.Card{}},
		},
		CardsPlayedThisTurn: 2, // Player has met the minimum
	}

	// 2. Execute
	newState, err := EndTurn(initialState, playerID)

	// 3. Assert
	require.NoError(t, err)
	require.NotNil(t, newState)
	require.True(t, newState.GameOver, "Game should be over because all cards are played")
	require.Equal(t, "You won! All cards have been played.", newState.Message)
}

func TestEndTurn_DeckEmptyRule(t *testing.T) {
	// 1. Setup
	playerID := "player-on-empty-deck"
	initialState := &pb.GameState{
		GameId:              "empty-deck-game",
		PlayerIds:           []string{playerID},
		CurrentTurnPlayerId: playerID,
		Deck:                []*pb.Card{}, // Deck is empty
		DeckSize:            0,
		Hands: map[string]*pb.Hand{
			playerID: {Cards: []*pb.Card{{Value: 20}}},
		},
		CardsPlayedThisTurn: 1, // Player has played one card
	}

	// 2. Execute
	// With an empty deck, playing one card should be enough to end the turn.
	_, err := EndTurn(initialState, playerID)

	// 3. Assert
	require.NoError(t, err, "Ending turn with 1 card should be allowed when deck is empty")
}

func TestEndTurn_NextPlayerHasNoMoves(t *testing.T) {
	// 1. Setup
	playerA := "player-a"
	playerB := "player-b-stuck"
	initialState := &pb.GameState{
		GameId:              "no-moves-for-next-player-game",
		PlayerIds:           []string{playerA, playerB},
		CurrentTurnPlayerId: playerA,
		DeckSize:            0, // Deck is empty
		Piles: map[string]*pb.Pile{
			"up1":   {Ascending: true, Cards: []*pb.Card{{Value: 98}}},
			"up2":   {Ascending: true, Cards: []*pb.Card{{Value: 99}}},
			"down1": {Ascending: false, Cards: []*pb.Card{{Value: 2}}},
			"down2": {Ascending: false, Cards: []*pb.Card{{Value: 3}}},
		},
		Hands: map[string]*pb.Hand{
			playerA: {Cards: []*pb.Card{}},            // Player A has finished their cards
			playerB: {Cards: []*pb.Card{{Value: 50}}}, // Player B has a card but no valid move
		},
		CardsPlayedThisTurn: 2, // Player A has met the minimum
	}

	// 2. Execute
	// Player A ends their turn, making it Player B's turn.
	newState, err := EndTurn(initialState, playerA)

	// 3. Assert
	require.NoError(t, err)
	require.True(t, newState.GameOver, "Game should be over because the next player has no valid moves")
	require.Contains(t, newState.Message, "lost: No more valid moves")
}
