package bot

import (
	"math"

	"the_game_card_game/pkg/game"
	pb "the_game_card_game/proto"
)

// TwoCardGreedyStrategy is a greedy strategy that aims to play exactly two cards
// before ending its turn.
type TwoCardGreedyStrategy struct{}

// NewTwoCardGreedyStrategy creates a new TwoCardGreedyStrategy.
func NewTwoCardGreedyStrategy() *TwoCardGreedyStrategy {
	return &TwoCardGreedyStrategy{}
}

func (s *TwoCardGreedyStrategy) GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error) {
	// If we've already played 2 or more cards, end the turn.
	// The server requires a minimum of 2 cards to be played (or 1 if the deck is empty),
	// so this strategy will always satisfy the minimum requirement before ending the turn.
	if gameState.GetCardsPlayedThisTurn() >= 2 {
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	possibleMoves := game.GetPossibleMoves(playerID, gameState)

	if len(possibleMoves) == 0 {
		// No moves left, must end turn regardless of cards played.
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	// Find the best "minimal jump" move.
	bestMove := possibleMoves[0]
	minJump := int32(math.MaxInt32)

	for _, move := range possibleMoves {
		pile := gameState.Piles[move.Pile]
		topCard := pile.Cards[len(pile.Cards)-1]
		jump := int32(math.Abs(float64(move.Card.Value - topCard.Value)))

		if jump < minJump {
			minJump = jump
			bestMove = move
		}
	}

	return &pb.PlayCardRequest{
		GameId: gameState.GameId, PlayerId: playerID, Card: bestMove.Card, PileId: bestMove.Pile,
	}, nil, nil
}
