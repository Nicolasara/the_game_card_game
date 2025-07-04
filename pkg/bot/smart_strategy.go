package bot

import (
	"math"

	"the_game_card_game/pkg/game"
	pb "the_game_card_game/proto"
)

// SmartStrategy combines defensive and offensive tactics. It prioritizes
// "backwards-10" moves that create the most space, and falls back to
// "minimal-jump" moves otherwise.
type SmartStrategy struct{}

// NewSmartStrategy creates a new SmartStrategy.
func NewSmartStrategy() *SmartStrategy {
	return &SmartStrategy{}
}

// GetNextMove implements the Strategy interface for SmartStrategy.
func (s *SmartStrategy) GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error) {
	possibleMoves := game.GetPossibleMoves(playerID, gameState)

	if len(possibleMoves) == 0 {
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	var tenMoves []game.Move
	var forwardMoves []game.Move
	var bestMove *game.Move

	for _, move := range possibleMoves {
		pile := gameState.Piles[move.Pile]
		topCard := pile.Cards[len(pile.Cards)-1]
		cardValue := move.Card.Value

		isTenRule := (pile.Ascending && cardValue == topCard.Value-10) || (!pile.Ascending && cardValue == topCard.Value+10)

		if isTenRule {
			tenMoves = append(tenMoves, move)
		} else {
			forwardMoves = append(forwardMoves, move)
		}
	}

	if len(tenMoves) > 0 {
		maxSpace := int32(-1)
		for _, move := range tenMoves {
			pile := gameState.Piles[move.Pile]
			topCard := pile.Cards[len(pile.Cards)-1]
			space := int32(math.Abs(float64(move.Card.Value - topCard.Value)))
			if space > maxSpace {
				maxSpace = space
				bestMove = &move
			}
		}
	} else if len(forwardMoves) > 0 {
		minJump := int32(math.MaxInt32)
		for _, move := range forwardMoves {
			pile := gameState.Piles[move.Pile]
			topCard := pile.Cards[len(pile.Cards)-1]
			jump := int32(math.Abs(float64(move.Card.Value - topCard.Value)))
			if jump < minJump {
				minJump = jump
				bestMove = &move
			}
		}
	}

	if bestMove == nil {
		// This should not happen if there are possible moves, but as a fallback, end the turn.
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	return &pb.PlayCardRequest{
		GameId:   gameState.GameId,
		PlayerId: playerID,
		Card:     bestMove.Card,
		PileId:   bestMove.Pile,
	}, nil, nil
}
