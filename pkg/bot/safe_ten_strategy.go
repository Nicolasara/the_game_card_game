package bot

import (
	"math"

	"the_game_card_game/pkg/game"
	pb "the_game_card_game/proto"
)

// SafeTenStrategy implements a defensive strategy that prioritizes making the
// special "10-back" move. If not available, it falls back to the minimal jump.
type SafeTenStrategy struct{}

// NewSafeTenStrategy creates a new SafeTenStrategy.
func NewSafeTenStrategy() *SafeTenStrategy {
	return &SafeTenStrategy{}
}

// GetNextMove determines the next move for the bot to make.
func (s *SafeTenStrategy) GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error) {
	possibleMoves := game.GetPossibleMoves(playerID, gameState)

	if len(possibleMoves) == 0 {
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	// Prioritize "10-back" moves.
	for _, move := range possibleMoves {
		pile := gameState.Piles[move.Pile]
		topCard := pile.Cards[len(pile.Cards)-1]
		cardValue := move.Card.Value

		if (pile.Ascending && cardValue == topCard.Value-10) || (!pile.Ascending && cardValue == topCard.Value+10) {
			return &pb.PlayCardRequest{
				GameId:   gameState.GameId,
				PlayerId: playerID,
				PileId:   move.Pile,
				Card:     move.Card,
			}, nil, nil
		}
	}

	// If no 10-back move is available, fall back to the minimal jump strategy.
	var bestMove game.Move
	minDiff := int32(math.MaxInt32)

	for _, move := range possibleMoves {
		pile := gameState.Piles[move.Pile]
		topCard := pile.Cards[len(pile.Cards)-1]
		cardValue := move.Card.Value

		diff := int32(math.Abs(float64(cardValue - topCard.Value)))
		if diff < minDiff {
			minDiff = diff
			bestMove = move
		}
	}

	return &pb.PlayCardRequest{
		GameId:   gameState.GameId,
		PlayerId: playerID,
		PileId:   bestMove.Pile,
		Card:     bestMove.Card,
	}, nil, nil
}
