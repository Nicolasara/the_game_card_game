package bot

import (
	"math/rand"
	"time"

	"the_game_card_game/pkg/game"
	pb "the_game_card_game/proto"
)

// RandomStrategy plays a random valid card.
type RandomStrategy struct {
	r *rand.Rand
}

// NewRandomStrategy creates a new RandomStrategy.
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetNextMove determines the next move for the bot to make.
func (s *RandomStrategy) GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error) {
	possibleMoves := game.GetPossibleMoves(playerID, gameState)

	if len(possibleMoves) == 0 {
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	// Pick a random move
	move := possibleMoves[s.r.Intn(len(possibleMoves))]

	return &pb.PlayCardRequest{
		GameId:   gameState.GameId,
		PlayerId: playerID,
		Card:     move.Card,
		PileId:   move.Pile,
	}, nil, nil
}
