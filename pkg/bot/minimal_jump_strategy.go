package bot

import (
	"math"

	"the_game_card_game/pkg/game"
	pb "the_game_card_game/proto"
)

// MinimalJumpStrategy implements a "greedy" strategy that plays the card
// that makes the smallest possible jump. It prioritizes "10-back" moves.
type MinimalJumpStrategy struct{}

// NewMinimalJumpStrategy creates a new MinimalJumpStrategy.
func NewMinimalJumpStrategy() *MinimalJumpStrategy {
	return &MinimalJumpStrategy{}
}

// GetNextMove determines the next move for the bot to make.
func (s *MinimalJumpStrategy) GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error) {
	possibleMoves := game.GetPossibleMoves(playerID, gameState)

	if len(possibleMoves) == 0 {
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	var bestMove game.Move
	minDiff := int32(math.MaxInt32)

	for _, move := range possibleMoves {
		pile := gameState.Piles[move.Pile]
		topCard := pile.Cards[len(pile.Cards)-1]
		cardValue := move.Card.Value

		// Prioritize 10-back moves
		if (pile.Ascending && cardValue == topCard.Value-10) || (!pile.Ascending && cardValue == topCard.Value+10) {
			bestMove = move
			break // This is always a good move, so we can take it immediately.
		}

		diff := int32(math.Abs(float64(cardValue - topCard.Value)))
		if diff < minDiff {
			minDiff = diff
			bestMove = move
		}
	}

	return &pb.PlayCardRequest{
		GameId:   gameState.GameId,
		PlayerId: playerID,
		Card:     bestMove.Card,
		PileId:   bestMove.Pile,
	}, nil, nil
}
