package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	pb "the_game_card_game/proto"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// --- TUI MODEL ---

type model struct {
	client         pb.GameServiceClient
	state          *pb.GameState
	playerID       string
	gameID         string
	hand           []int32
	mode           string // "select-card", "select-pile", "confirm-quit"
	selectedCard   int32
	selectedPile   int // 0: up1, 1: up2, 2: down1, 3: down2
	status         string
	err            error
}

var pileIDs = []string{"up1", "up2", "down1", "down2"}

type stateUpdateMsg *pb.GameState
type statusUpdateMsg string
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func newModel(client pb.GameServiceClient, playerID, gameID string) model {
	return model{
		client:   client,
		playerID: playerID,
		gameID:   gameID,
		mode:     "select-card",
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stateUpdateMsg:
		m.state = msg
		if hand, ok := m.state.Hands[m.playerID]; ok {
			m.hand = make([]int32, len(hand.Cards))
			for i, c := range hand.Cards {
				m.hand[i] = c.Value
			}
			sort.Slice(m.hand, func(i, j int) bool { return m.hand[i] < m.hand[j] })
		}
		m.mode = "select-card"
		m.selectedCard = 0
		m.status = m.getTurnStatus()
		return m, nil

	case statusUpdateMsg:
		m.status = string(msg)
		return m, nil

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		return handleKeyPress(m, msg)
	}
	return m, nil
}

func handleKeyPress(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If in confirm-quit mode, only handle y/n/esc.
	if m.mode == "confirm-quit" {
		switch msg.String() {
		case "y", "Y":
			return m, tea.Quit
		case "n", "N", "esc":
			m.mode = "select-card"
			m.status = m.getTurnStatus()
			return m, nil
		}
		return m, nil
	}

	// Global quit
	if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
		m.mode = "confirm-quit"
		m.status = "Are you sure you want to quit? (y/n)"
		return m, nil
	}

	// Only allow actions if it's our turn
	if m.state.CurrentTurnPlayerId != m.playerID {
		return m, nil
	}
	
	switch m.mode {
	case "select-card":
		if msg.String() >= "1" && msg.String() <= "9" {
			cardIndex, _ := strconv.Atoi(msg.String())
			if cardIndex > 0 && cardIndex <= len(m.hand) {
				m.selectedCard = m.hand[cardIndex-1]
				m.mode = "select-pile"
				m.status = fmt.Sprintf("Selected card %d. Use ←/→ to pick a pile, Enter to play.", m.selectedCard)
			}
		} else if msg.String() == "e" {
			if m.state.CardsPlayedThisTurn >= 2 {
				return m, m.endTurnCmd()
			} else {
				m.status = fmt.Sprintf("You must play at least 2 cards to end your turn (played %d).", m.state.CardsPlayedThisTurn)
			}
		}

	case "select-pile":
		switch msg.String() {
		case "left":
			m.selectedPile = (m.selectedPile - 1 + 4) % 4
		case "right":
			m.selectedPile = (m.selectedPile + 1) % 4
		case "enter":
			return m, m.playCardCmd()
		case "esc":
			m.mode = "select-card"
			m.selectedCard = 0
			m.status = m.getTurnStatus()
		}
	}
	return m, nil
}

func (m *model) playCardCmd() tea.Cmd {
	return func() tea.Msg {
		req := &pb.PlayCardRequest{
			GameId:   m.gameID,
			PlayerId: m.playerID,
			Card:     &pb.Card{Value: m.selectedCard},
			PileId:   pileIDs[m.selectedPile],
		}
		res, err := m.client.PlayCard(context.Background(), req)
		if err != nil {
			return statusUpdateMsg(fmt.Sprintf("Error: %v", err))
		}
		if !res.Success {
			return statusUpdateMsg(fmt.Sprintf("Invalid move: %s", res.Message))
		}
		return statusUpdateMsg(fmt.Sprintf("Played %d on %s.", m.selectedCard, pileIDs[m.selectedPile]))
	}
}

func (m *model) endTurnCmd() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.EndTurn(context.Background(), &pb.EndTurnRequest{GameId: m.gameID, PlayerId: m.playerID})
		if err != nil {
			return statusUpdateMsg(fmt.Sprintf("Error ending turn: %v", err))
		}
		return statusUpdateMsg("Turn ended. Waiting for next state...")
	}
}

// --- TUI VIEW & STYLES ---
var (
	baseStyle         = lipgloss.NewStyle().Padding(1, 2)
	pileStyle         = baseStyle.Copy().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Width(12).Align(lipgloss.Center)
	selectedPileStyle = pileStyle.Copy().BorderForeground(lipgloss.Color("228"))
	handStyle         = baseStyle.Copy().Border(lipgloss.DoubleBorder(), true).BorderForeground(lipgloss.Color("228"))
	faintStyle        = lipgloss.NewStyle().Faint(true)
)

func (m model) View() string {
	if m.err != nil { return fmt.Sprintf("Error: %v\n", m.err) }
	if m.state == nil { return "Connecting to game..." }

	// Piles View
	var pileViews []string
	isMyTurn := m.state.CurrentTurnPlayerId == m.playerID
	for i, id := range pileIDs {
		style := pileStyle
		if isMyTurn && m.mode == "select-pile" && i == m.selectedPile {
			style = selectedPileStyle
		}
		pileViews = append(pileViews, getPileView(id, m.state.Piles[id], style))
	}
	pilesView := lipgloss.JoinHorizontal(lipgloss.Top, pileViews...)

	// Hand View
	var handItems []string
	for i, v := range m.hand {
		cardStr := strconv.Itoa(int(v))
		if isMyTurn && m.selectedCard == v {
			cardStr = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Render(cardStr)
		}
		handItems = append(handItems, fmt.Sprintf("%d:%s", i+1, cardStr))
	}
	handView := handStyle.Render(fmt.Sprintf("Your Hand (%s):\n%s", m.playerID, strings.Join(handItems, "  ")))

	// Status & Help
	help := " | 'q': quit"
	if m.mode == "confirm-quit" {
		help = "" // No extra help in this mode
	} else if isMyTurn {
		if m.mode == "select-pile" {
			help += " | 'esc': cancel selection"
		} else if m.state.CardsPlayedThisTurn >= 2 {
			help += " | 'e': end turn"
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, pilesView, handView, faintStyle.Render(m.status+help))
}

func getPileView(name string, pile *pb.Pile, style lipgloss.Style) string {
	initial, displayName := "1", "UP ⬆"
	if strings.HasPrefix(name, "down") {
		initial, displayName = "100", "DOWN ⬇"
	}
	topCard := initial
	if pile != nil && len(pile.Cards) > 0 {
		topCard = strconv.Itoa(int(pile.Cards[len(pile.Cards)-1].Value))
	}
	return style.Render(fmt.Sprintf("%s\n\n%s", displayName, topCard))
}

func (m *model) getTurnStatus() string {
	if m.state.CurrentTurnPlayerId == m.playerID {
		return fmt.Sprintf("Your turn! Select a card (1-%d). %d card(s) played.", len(m.hand), m.state.CardsPlayedThisTurn)
	}
	return fmt.Sprintf("Waiting for %s's turn...", m.state.CurrentTurnPlayerId)
}

// --- MAIN & GRPC ---

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	gameID := fs.String("game", "", "Game ID to join")
	playerID := fs.String("player", "default-player", "Your Player ID")
	create := fs.Bool("create", false, "Create a new game")
	fs.Parse(os.Args[1:])

	if !*create && *gameID == "" { log.Fatal("Use -create or provide a -game ID.") }
	if *playerID == "" { log.Fatal("-player is required.") }

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("did not connect: %v", err) }
	defer conn.Close()
	client := pb.NewGameServiceClient(conn)

	if *create {
		res, err := client.CreateGame(context.Background(), &pb.CreateGameRequest{PlayerId: *playerID})
		if err != nil { log.Fatalf("Failed to create game: %v", err) }
		*gameID = res.GetGameState().GetGameId()
		log.Printf("Game created: %s. Starting TUI...", *gameID)
		time.Sleep(1 * time.Second)
	}

	m := newModel(client, *playerID, *gameID)
	p := tea.NewProgram(m, tea.WithAltScreen())
	go streamState(p, client, *gameID)

	if _, err := p.Run(); err != nil { log.Fatalf("Error running TUI: %v", err) }
}

func streamState(p *tea.Program, client pb.GameServiceClient, gameID string) {
	stream, err := client.StreamGameState(context.Background(), &pb.StreamGameStateRequest{GameId: gameID})
	if err != nil {
		p.Send(errMsg{err})
		return
	}
	for {
		state, err := stream.Recv()
		if errors.Is(err, io.EOF) { break }
		if err != nil {
			p.Send(errMsg{err})
			return
		}
		p.Send(stateUpdateMsg(state))
	}
} 