package game

import (
	"fmt"
	"math/rand"
	"time"

	pb "the_game_card_game/proto"

	"google.golang.org/protobuf/proto"
)

// NewGame initializes a new game state.
func NewGame(gameID string, playerID string) *pb.GameState {
	const handSize = 8
	deck := createShuffledDeck()

	hand := deck[:handSize]
	remainingDeck := deck[handSize:]

	return &pb.GameState{
		GameId:              gameID,
		PlayerIds:           []string{playerID},
		DeckSize:            int32(len(remainingDeck)),
		Deck:                remainingDeck,
		CurrentTurnPlayerId: playerID, // First player starts
		CardsPlayedThisTurn: 0,
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
	newState := proto.Clone(state).(*pb.GameState)

	pile, ok := newState.Piles[pileID]
	if !ok {
		return nil, fmt.Errorf("pile '%s' not found", pileID)
	}

	topCard := pile.Cards[len(pile.Cards)-1]
	isTenBack := (pile.Ascending && cardValue == topCard.Value-10) || (!pile.Ascending && cardValue == topCard.Value+10)
	isValid := (pile.Ascending && cardValue > topCard.Value) || (!pile.Ascending && cardValue < topCard.Value) || isTenBack
	if !isValid {
		return nil, fmt.Errorf("invalid move: card %d on pile %s (top: %d)", cardValue, pileID, topCard.Value)
	}

	playerHand, ok := newState.Hands[playerID]
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
	newState.CardsPlayedThisTurn++

	// After playing, check if the game is lost
	if newState.CardsPlayedThisTurn < 2 && len(playerHand.Cards) > 0 {
		if !isMovePossible(playerHand, newState.Piles) {
			newState.GameOver = true
			newState.Message = fmt.Sprintf("Player %s lost: No more valid moves.", playerID)
		}
	}

	return newState, nil
}

// isMovePossible checks if any card in the hand can be legally played on any pile.
func isMovePossible(hand *pb.Hand, piles map[string]*pb.Pile) bool {
	for _, card := range hand.Cards {
		for _, pile := range piles {
			topCard := pile.Cards[len(pile.Cards)-1]

			// Check normal play rule
			if pile.Ascending && card.Value > topCard.Value {
				return true
			}
			if !pile.Ascending && card.Value < topCard.Value {
				return true
			}

			// Check "10-rule"
			if pile.Ascending && card.Value == topCard.Value-10 {
				return true
			}
			if !pile.Ascending && card.Value == topCard.Value+10 {
				return true
			}
		}
	}
	return false
}

// EndTurn replenishes the player's hand, resets the turn counter, and advances to the next player.
func EndTurn(state *pb.GameState, playerID string) (*pb.GameState, error) {
	// Validate that the player has played enough cards.
	// The minimum is 2, unless the deck is empty, in which case it's 1.
	minCards := 2
	if state.DeckSize == 0 {
		minCards = 1
	}

	if state.CardsPlayedThisTurn < int32(minCards) {
		return nil, fmt.Errorf("must play at least %d card(s) to end turn (played %d)", minCards, state.CardsPlayedThisTurn)
	}

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

	// Check for win condition after advancing the turn
	allHandsEmpty := true
	for _, hand := range state.Hands {
		if len(hand.Cards) > 0 {
			allHandsEmpty = false
			break
		}
	}

	if state.DeckSize == 0 && allHandsEmpty {
		state.GameOver = true
		state.Message = "You won! All cards have been played."
	}

	return state, nil
} 