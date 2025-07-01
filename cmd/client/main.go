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
	mode           string // "select-card", "select-pile"
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
		status:   "Select a card to play using number keys (1-9).",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

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
		m.mode = "select-card" // Reset mode after any update
		m.status = "Select a card (1-" + strconv.Itoa(len(m.hand)) + "), then select a pile (←/→)."
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
	switch m.mode {
	case "select-card":
		if msg.Type >= tea.KeyRunes && msg.String() >= "1" && msg.String() <= "9" {
			cardIndex, _ := strconv.Atoi(msg.String())
			cardIndex-- // 1-based to 0-based
			if cardIndex < len(m.hand) {
				m.selectedCard = m.hand[cardIndex]
				m.mode = "select-pile"
				m.status = fmt.Sprintf("Selected card %d. Select a pile with ←/→, then press Enter to play.", m.selectedCard)
			} else {
				m.status = fmt.Sprintf("Invalid card selection. Choose 1-%d.", len(m.hand))
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
			m.status = "Selection cancelled. Select a card to play (1-9)."
		}
	}

	if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	return m, nil
}

func (m *model) playCardCmd() tea.Cmd {
	return func() tea.Msg {
		pileID := pileIDs[m.selectedPile]
		req := &pb.PlayCardRequest{
			GameId:   m.gameID,
			PlayerId: m.playerID,
			Card:     &pb.Card{Value: m.selectedCard},
			PileId:   pileID,
		}
		res, err := m.client.PlayCard(context.Background(), req)
		if err != nil {
			return statusUpdateMsg(fmt.Sprintf("Error playing card: %v", err))
		}
		if !res.Success {
			return statusUpdateMsg(fmt.Sprintf("Move rejected: %s", res.Message))
		}
		// On success, we don't need to do anything.
		// The stream will send a stateUpdateMsg which will redraw the board.
		return statusUpdateMsg(fmt.Sprintf("Played %d on %s. Waiting for next state...", m.selectedCard, pileID))
	}
}

// --- TUI STYLES & VIEW ---
var (
	baseStyle = lipgloss.NewStyle().Padding(1, 2)
	pileStyle = baseStyle.Copy().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(12).
			Align(lipgloss.Center)
	selectedPileStyle = pileStyle.Copy().BorderForeground(lipgloss.Color("228"))
	handStyle         = baseStyle.Copy().Border(lipgloss.DoubleBorder(), true).BorderForeground(lipgloss.Color("228"))
)

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("An error occurred: %v\n", m.err)
	}
	if m.state == nil {
		return "Connecting to game stream..."
	}

	// Piles
	var pileViews []string
	for i, id := range pileIDs {
		style := pileStyle
		if m.mode == "select-pile" && i == m.selectedPile {
			style = selectedPileStyle
		}
		initial := "100"
		if strings.HasPrefix(id, "up") {
			initial = "1"
		}
		pileViews = append(pileViews, getPileView(id, initial, m.state.Piles[id], style))
	}
	pilesView := lipgloss.JoinHorizontal(lipgloss.Top, pileViews...)

	// Hand
	var handItems []string
	for i, v := range m.hand {
		cardStr := strconv.Itoa(int(v))
		if m.selectedCard == v {
			cardStr = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Render(cardStr)
		}
		handItems = append(handItems, fmt.Sprintf("%d: %s", i+1, cardStr))
	}
	handStr := fmt.Sprintf("Your Hand (%s):\n%s", m.playerID, strings.Join(handItems, "  "))
	handView := handStyle.Render(handStr)

	// Status & Help
	help := " | Press 'q' or 'ctrl+c' to quit."
	if m.mode == "select-pile" {
		help = " | 'esc' to re-select card" + help
	}
	statusView := lipgloss.NewStyle().Faint(true).Render(m.status + help)

	return lipgloss.JoinVertical(lipgloss.Left, pilesView, handView, statusView)
}

func getPileView(name, initialValue string, pile *pb.Pile, style lipgloss.Style) string {
	topCard := initialValue
	if pile != nil && len(pile.Cards) > 0 {
		topCard = strconv.Itoa(int(pile.Cards[len(pile.Cards)-1].Value))
	}
	displayName := strings.ToUpper(name)
	if strings.HasPrefix(name, "up") {
		displayName = "UP ⬆"
	} else {
		displayName = "DOWN ⬇"
	}
	return style.Render(fmt.Sprintf("%s\n\n%s", displayName, topCard))
}

// --- MAIN LOGIC ---

func main() {
	// --- Command Line Flags ---
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		gameID   = fs.String("game", "", "Game ID to join or stream")
		playerID = fs.String("player", "default-player", "Your Player ID")
		create   = fs.Bool("create", false, "Create a new game")
	)
	fs.Parse(os.Args[1:])

	if !*create && *gameID == "" {
		log.Fatalf("You must either provide a -game ID to join or use the -create flag.")
	}
	if *playerID == "" {
		log.Fatalf("-player flag cannot be empty.")
	}

	// --- gRPC Connection ---
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewGameServiceClient(conn)

	// --- Create game if requested ---
	if *create {
		res, err := client.CreateGame(context.Background(), &pb.CreateGameRequest{PlayerId: *playerID})
		if err != nil {
			log.Fatalf("Failed to create game: %v", err)
		}
		*gameID = res.GetGameState().GetGameId()
		log.Printf("Game created with ID: %s. Starting stream...", *gameID)
		time.Sleep(2 * time.Second) // Give user time to see the ID
	}

	// --- Run TUI ---
	m := newModel(client, *playerID, *gameID)
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	go streamState(p, client, *gameID)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

func streamState(p *tea.Program, client pb.GameServiceClient, gameID string) {
	stream, err := client.StreamGameState(context.Background(), &pb.StreamGameStateRequest{GameId: gameID})
	if err != nil {
		p.Send(errMsg{err})
		return
	}
	for {
		state, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break // Stream closed cleanly
		}
		if err != nil {
			p.Send(errMsg{err})
			return
		}
		p.Send(stateUpdateMsg(state))
	}
} 