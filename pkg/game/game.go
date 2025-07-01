package game

import (
	"fmt"
	"math/rand"
	"time"

	pb "the_game_card_game/proto"
)

// NewGame initializes a new game state.
func NewGame(gameID string, playerID string) *pb.GameState {
	const handSize = 8
	deck := createShuffledDeck()

	hand := deck[:handSize]
	remainingDeck := deck[handSize:]

	return &pb.GameState{
		GameId:                 gameID,
		PlayerIds:              []string{playerID},
		DeckSize:               int32(len(remainingDeck)),
		Deck:                   remainingDeck,
		CurrentTurnPlayerId:    playerID, // First player starts
		CardsPlayedThisTurn:    0,
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

func createShuffledDeck() []*pb.Card {
	deck := make([]*pb.Card, 98)
	for i := 0; i < 98; i++ {
		deck[i] = &pb.Card{Value: int32(i + 2)}
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

// AddPlayer adds a new player to the game state and deals them a hand.
func AddPlayer(state *pb.GameState, playerID string, handSize int) (*pb.GameState, error) {
	if len(state.Deck) < handSize {
		return nil, fmt.Errorf("not enough cards in deck to deal a new hand")
	}

	state.PlayerIds = append(state.PlayerIds, playerID)
	hand := state.Deck[:handSize]
	state.Deck = state.Deck[handSize:]
	state.DeckSize = int32(len(state.Deck))
	state.Hands[playerID] = &pb.Hand{Cards: hand}
	return state, nil
}

// PlayCard validates a move, updates the game state, and increments the turn's card counter.
func PlayCard(state *pb.GameState, playerID string, cardValue int32, pileID string) (*pb.GameState, error) {
	pile, ok := state.Piles[pileID]
	if !ok {
		return nil, fmt.Errorf("pile '%s' not found", pileID)
	}

	topCard := pile.Cards[len(pile.Cards)-1]
	isTenBack := (pile.Ascending && cardValue == topCard.Value-10) || (!pile.Ascending && cardValue == topCard.Value+10)
	isValid := (pile.Ascending && cardValue > topCard.Value) || (!pile.Ascending && cardValue < topCard.Value) || isTenBack
	if !isValid {
		return nil, fmt.Errorf("invalid move: card %d on pile %s (top: %d)", cardValue, pileID, topCard.Value)
	}

	playerHand, ok := state.Hands[playerID]
	if !ok {
		return nil, fmt.Errorf("player '%s' not found", playerID)
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
	pile.Cards = append(pile.Cards, playedCard)
	state.CardsPlayedThisTurn++
	return state, nil
}

// EndTurn replenishes the player's hand, resets the turn counter, and advances to the next player.
func EndTurn(state *pb.GameState, playerID string) (*pb.GameState, error) {
	// Replenish hand
	numToDraw := int(state.CardsPlayedThisTurn)
	if numToDraw > 0 {
		hand, ok := state.Hands[playerID]
		if !ok {
			return nil, fmt.Errorf("player '%s' not found", playerID)
		}
		
		drawCount := 0
		if len(state.Deck) < numToDraw {
			drawCount = len(state.Deck)
		} else {
			drawCount = numToDraw
		}

		if drawCount > 0 {
			hand.Cards = append(hand.Cards, state.Deck[:drawCount]...)
			state.Deck = state.Deck[drawCount:]
			state.DeckSize = int32(len(state.Deck))
		}
	}

	// Reset counter
	state.CardsPlayedThisTurn = 0

	// Advance to next player
	currentPlayerIndex := -1
	for i, id := range state.PlayerIds {
		if id == playerID {
			currentPlayerIndex = i
			break
		}
	}
	if currentPlayerIndex == -1 {
		return nil, fmt.Errorf("could not find current player '%s' in player list", playerID)
	}
	nextPlayerIndex := (currentPlayerIndex + 1) % len(state.PlayerIds)
	state.CurrentTurnPlayerId = state.PlayerIds[nextPlayerIndex]

	return state, nil
} 