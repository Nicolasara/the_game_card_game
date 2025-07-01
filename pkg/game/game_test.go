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