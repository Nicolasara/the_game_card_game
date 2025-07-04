package bot

import (
	"math"

	"the_game_card_game/pkg/game"
	pb "the_game_card_game/proto"
)

type PhasedStrategy struct{}

func NewPhasedStrategy() *PhasedStrategy {
	return &PhasedStrategy{}
}

const (
	earlyGameDeckSize = 65
	midGameDeckSize   = 30
	extremeCardMin    = 20
	extremeCardMax    = 80
	midCardMin        = 40
	midCardMax        = 60
)

type scoredMove struct {
	move      game.Move
	isTenJump bool
	jump      int32
}

func (s *PhasedStrategy) GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error) {
	possibleMoves := game.GetPossibleMoves(playerID, gameState)
	if len(possibleMoves) == 0 {
		return nil, &pb.EndTurnRequest{GameId: gameState.GameId, PlayerId: playerID}, nil
	}

	// Score all possible moves
	var scoredMoves []scoredMove
	for _, move := range possibleMoves {
		pile := gameState.Piles[move.Pile]
		topCard := pile.Cards[len(pile.Cards)-1].Value
		cardValue := move.Card.Value
		isTen := (pile.Ascending && cardValue == topCard-10) || (!pile.Ascending && cardValue == topCard+10)
		scoredMoves = append(scoredMoves, scoredMove{
			move:      move,
			isTenJump: isTen,
			jump:      int32(math.Abs(float64(cardValue - topCard))),
		})
	}

	// 1. Always prioritize "backwards-10" moves
	var tenMoves []scoredMove
	var forwardMoves []scoredMove
	for _, sm := range scoredMoves {
		if sm.isTenJump {
			tenMoves = append(tenMoves, sm)
		} else {
			forwardMoves = append(forwardMoves, sm)
		}
	}

	if len(tenMoves) > 0 {
		bestMove := tenMoves[0]
		maxSpace := int32(0)
		for _, sm := range tenMoves {
			if sm.jump > maxSpace {
				maxSpace = sm.jump
				bestMove = sm
			}
		}
		return &pb.PlayCardRequest{
			GameId: gameState.GameId, PlayerId: playerID, Card: bestMove.move.Card, PileId: bestMove.move.Pile,
		}, nil, nil
	}

	// 2. If no ten-moves, determine game phase and filter moves
	deckSize := gameState.GetDeckSize()
	var phaseFilteredMoves []scoredMove

	switch {
	case deckSize > earlyGameDeckSize: // Early Game
		for _, sm := range forwardMoves {
			if sm.move.Card.Value < extremeCardMin || sm.move.Card.Value > extremeCardMax {
				phaseFilteredMoves = append(phaseFilteredMoves, sm)
			}
		}
	case deckSize > midGameDeckSize: // Mid Game
		for _, sm := range forwardMoves {
			if sm.move.Card.Value >= midCardMin && sm.move.Card.Value <= midCardMax {
				phaseFilteredMoves = append(phaseFilteredMoves, sm)
			}
		}
	}

	movesToConsider := forwardMoves
	if len(phaseFilteredMoves) > 0 {
		movesToConsider = phaseFilteredMoves
	} else if deckSize <= midGameDeckSize { // Late game, consider all forward moves
		movesToConsider = forwardMoves
	}

	// 3. From the considered moves, find the one with the minimal jump
	bestMove := movesToConsider[0]
	minJump := int32(math.MaxInt32)
	for _, sm := range movesToConsider {
		if sm.jump < minJump {
			minJump = sm.jump
			bestMove = sm
		}
	}

	return &pb.PlayCardRequest{
		GameId: gameState.GameId, PlayerId: playerID, Card: bestMove.move.Card, PileId: bestMove.move.Pile,
	}, nil, nil
}
