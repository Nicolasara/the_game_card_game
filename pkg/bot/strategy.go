package bot

import (
	pb "the_game_card_game/proto"
)

// Strategy defines the interface for a game-playing bot.
type Strategy interface {
	// GetNextMove determines the next move for the bot to make.
	GetNextMove(playerID string, gameState *pb.GameState) (*pb.PlayCardRequest, *pb.EndTurnRequest, error)
}
