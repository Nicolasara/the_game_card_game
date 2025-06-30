package game

import (
	"fmt"
	"math/rand"
	pb "the_game_card_game/proto"
	"time"
)

// NewGame initializes a new game state.
func NewGame(gameID string, playerID string) *pb.GameState {
	// For now, we assume a single-player game. Hand size would vary with more players.
	const handSize = 8

	// Create a deck of cards from 2 to 99
	deck := make([]*pb.Card, 98)
	for i := 0; i < 98; i++ {
		deck[i] = &pb.Card{Value: int32(i + 2)}
	}

	// Shuffle the deck
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	// Deal the initial hand
	hand := deck[:handSize]
	remainingDeck := deck[handSize:]

	return &pb.GameState{
		GameId:    gameID,
		PlayerIds: []string{playerID},
		DeckSize:  int32(len(remainingDeck)),
		Deck:      remainingDeck,
		Piles: map[string]*pb.Pile{
			"up1":   {Ascending: true, Cards: []*pb.Card{{Value: 1}}},
			"up2":   {Ascending: true, Cards: []*pb.Card{{Value: 1}}},
			"down1": {Ascending: false, Cards: []*pb.Card{{Value: 100}}},
			"down2": {Ascending: false, Cards: []*pb.Card{{Value: 100}}},
		},
		Hands: map[string]*pb.Hand{
			playerID: {Cards: hand},
		},
	}
}

// AddPlayer adds a new player to the game state and deals them a hand.
func AddPlayer(state *pb.GameState, playerID string, handSize int) (*pb.GameState, error) {
	// Check if there are enough cards in the deck
	if len(state.Deck) < handSize {
		return nil, fmt.Errorf("not enough cards in deck to deal a new hand")
	}

	// Add player to the list
	state.PlayerIds = append(state.PlayerIds, playerID)

	// Deal hand
	hand := state.Deck[:handSize]
	state.Deck = state.Deck[handSize:]
	state.DeckSize = int32(len(state.Deck))
	state.Hands[playerID] = &pb.Hand{Cards: hand}

	return state, nil
}

// PlayCard validates a move and updates the game state.
func PlayCard(state *pb.GameState, playerID string, cardValue int32, pileID string) (*pb.GameState, error) {
	// Find the pile
	pile, ok := state.Piles[pileID]
	if !ok {
		return nil, fmt.Errorf("pile '%s' not found", pileID)
	}

	// Validate the move
	topCard := pile.Cards[len(pile.Cards)-1]
	isTenBack := false
	if pile.Ascending {
		isTenBack = cardValue == topCard.Value-10
	} else {
		isTenBack = cardValue == topCard.Value+10
	}

	isValid := (pile.Ascending && cardValue > topCard.Value) || (!pile.Ascending && cardValue < topCard.Value) || isTenBack
	if !isValid {
		return nil, fmt.Errorf("invalid move: card %d cannot be played on pile %s (top card: %d)", cardValue, pileID, topCard.Value)
	}

	// Find and remove the card from the player's hand
	playerHand, ok := state.Hands[playerID]
	if !ok {
		return nil, fmt.Errorf("player '%s' not found in game", playerID)
	}

	cardIndex := -1
	for i, c := range playerHand.Cards {
		if c.Value == cardValue {
			cardIndex = i
			break
		}
	}

	if cardIndex == -1 {
		return nil, fmt.Errorf("player does not have card %d", cardValue)
	}
	playedCard := playerHand.Cards[cardIndex]
	playerHand.Cards = append(playerHand.Cards[:cardIndex], playerHand.Cards[cardIndex+1:]...)

	// Add the card to the pile
	pile.Cards = append(pile.Cards, playedCard)

	// TODO: Draw new cards for the player
	// TODO: Check for win/loss conditions

	return state, nil
} 