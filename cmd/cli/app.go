package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/pieces"
	"github.com/divijg19/Swapchess/view"
)

type rendererMode int

const (
	renderView rendererMode = iota
	renderEngine
)

func (r rendererMode) String() string {
	if r == renderEngine {
		return "engine"
	}
	return "view"
}

type model struct {
	state        *engine.GameState
	vi           view.ViewState
	input        textinput.Model
	msg          string
	history      []*engine.GameState
	renderer     rendererMode
	pieceCatalog *pieces.Catalog
	width        int
	height       int
	// promotion flow
	awaitingPromotion bool
	pendingMove       engine.Move
}

func initialModel() model {
	st := engine.NewGame()
	m := model{state: st}
	m.vi = view.ViewStateFromGameState(st)
	m.renderer = renderView
	m.input = textinput.New()
	m.input.Placeholder = "move: e2e4 | commands: help, undo, renderer view|engine, quit"
	m.input.Focus()
	m.input.CharLimit = 64
	m.input.Width = 48
	m.msg = "Enter a move like e2e4. Type help for commands."
	m.pieceCatalog = pieces.NewCatalog(filepath.Join("assets", "pieces"))
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func emptySquareGlyph(file, rank int) string {
	if (file+rank)%2 == 0 {
		return "·"
	}
	return "◦"
}

func renderBoardGrid(cells [8][8]string) string {
	out := "    a   b   c   d   e   f   g   h\n"
	out += "  +---+---+---+---+---+---+---+---+\n"
	for rank := 7; rank >= 0; rank-- {
		out += fmt.Sprintf("%d |", rank+1)
		for file := 0; file < 8; file++ {
			out += fmt.Sprintf(" %s |", cells[rank][file])
		}
		out += fmt.Sprintf(" %d\n", rank+1)
		out += "  +---+---+---+---+---+---+---+---+\n"
	}
	out += "    a   b   c   d   e   f   g   h\n"
	return out
}

func renderBoard(s *engine.GameState, catalog *pieces.Catalog) string {
	var b [8][8]string
	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := s.Board.Squares[f][r]
			if p == nil {
				b[r][f] = emptySquareGlyph(f, r)
				continue
			}
			b[r][f] = catalog.Glyph(p)
		}
	}
	return renderBoardGrid(b)
}

func renderBoardFromView(v view.ViewState, catalog *pieces.Catalog) string {
	var b [8][8]string
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			b[r][f] = emptySquareGlyph(f, r)
		}
	}
	for _, vp := range v.Pieces {
		glyph := catalog.GlyphFor(vp.Kind, vp.Color)
		if vp.Y >= 0 && vp.Y < 8 && vp.X >= 0 && vp.X < 8 {
			b[vp.Y][vp.X] = glyph
		}
	}
	return renderBoardGrid(b)
}

func parseMove(s string) (engine.Move, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return engine.Move{}, fmt.Errorf("empty move")
	}
	// normalize common separators and casing
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "->", "")
	s = strings.ReplaceAll(s, "-", "")

	// allow optional promotion suffix (e.g., e7e8q)
	if len(s) != 4 && len(s) != 5 {
		return engine.Move{}, fmt.Errorf("invalid move format; expected like e2e4 or e7e8q")
	}

	fileOf := func(c byte) (int, error) {
		if c < 'a' || c > 'h' {
			return 0, fmt.Errorf("file out of range: %c", c)
		}
		return int(c - 'a'), nil
	}
	rankOf := func(c byte) (int, error) {
		if c < '1' || c > '8' {
			return 0, fmt.Errorf("rank out of range: %c", c)
		}
		return int(c - '1'), nil
	}

	f0, err := fileOf(s[0])
	if err != nil {
		return engine.Move{}, err
	}
	r0, err := rankOf(s[1])
	if err != nil {
		return engine.Move{}, err
	}
	f1, err := fileOf(s[2])
	if err != nil {
		return engine.Move{}, err
	}
	r1, err := rankOf(s[3])
	if err != nil {
		return engine.Move{}, err
	}

	from := engine.Position{File: f0, Rank: r0}
	to := engine.Position{File: f1, Rank: r1}
	mv := engine.Move{From: from, To: to}
	if len(s) == 5 {
		switch s[4] {
		case 'q':
			mv.Promotion = engine.Queen
		case 'r':
			mv.Promotion = engine.Rook
		case 'b':
			mv.Promotion = engine.Bishop
		case 'n':
			mv.Promotion = engine.Knight
		default:
			return engine.Move{}, fmt.Errorf("unknown promotion piece: %c", s[4])
		}
		mv.PromotionSet = true
	}
	return mv, nil
}

func moveString(mv engine.Move) string {
	base := fmt.Sprintf("%c%d%c%d", byte('a'+mv.From.File), mv.From.Rank+1, byte('a'+mv.To.File), mv.To.Rank+1)
	if !mv.HasExplicitPromotion() {
		return base
	}
	switch mv.Promotion {
	case engine.Queen:
		return base + "q"
	case engine.Rook:
		return base + "r"
	case engine.Bishop:
		return base + "b"
	case engine.Knight:
		return base + "n"
	default:
		return base
	}
}

func promotionChoice(s string) (engine.PieceKind, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "q", "queen":
		return engine.Queen, true
	case "r", "rook":
		return engine.Rook, true
	case "b", "bishop":
		return engine.Bishop, true
	case "n", "knight":
		return engine.Knight, true
	default:
		return engine.Pawn, false
	}
}

func (m model) applyMove(mv engine.Move) model {
	prev := m.state.Clone()
	if err := engine.ApplyMove(m.state, mv); err != nil {
		m.msg = "Illegal move: " + err.Error()
		return m
	}
	m.history = append(m.history, prev)
	m.vi = view.ViewStateFromGameState(m.state)
	m.msg = "Move applied: " + moveString(mv)
	return m
}

func (m model) submitInput() (model, tea.Cmd) {
	val := strings.TrimSpace(m.input.Value())
	if val == "" {
		m.msg = "Empty input. Enter a move like e2e4 or a command (help)."
		return m, nil
	}
	m.input.SetValue("")

	if m.awaitingPromotion {
		pk, ok := promotionChoice(val)
		if !ok {
			m.msg = "Promotion required: enter q, r, b, or n."
			return m, nil
		}
		mv := m.pendingMove
		mv.Promotion = pk
		mv.PromotionSet = true
		m.awaitingPromotion = false
		m = m.applyMove(mv)
		return m, nil
	}

	lower := strings.ToLower(val)
	switch lower {
	case "help", "?":
		m.msg = "Commands: undo|u, renderer view|engine, renderer toggle, quit|exit. Moves: e2e4 or e7e8q."
		return m, nil
	case "undo", "u":
		if len(m.history) == 0 {
			m.msg = "No moves to undo."
			return m, nil
		}
		last := m.history[len(m.history)-1]
		m.history = m.history[:len(m.history)-1]
		m.state = last
		m.vi = view.ViewStateFromGameState(m.state)
		m.awaitingPromotion = false
		m.msg = "Undid last move."
		return m, nil
	case "quit", "exit":
		return m, tea.Quit
	case "renderer toggle", "render toggle":
		if m.renderer == renderView {
			m.renderer = renderEngine
		} else {
			m.renderer = renderView
		}
		m.msg = "Renderer: " + m.renderer.String()
		return m, nil
	case "renderer view", "render view", "view":
		m.renderer = renderView
		m.msg = "Renderer: view"
		return m, nil
	case "renderer engine", "render engine", "engine":
		m.renderer = renderEngine
		m.msg = "Renderer: engine"
		return m, nil
	}

	mv, err := parseMove(val)
	if err != nil {
		m.msg = "Parse error: " + err.Error()
		return m, nil
	}
	piece := m.state.Board.Squares[mv.From.File][mv.From.Rank]
	if piece != nil && piece.Kind == engine.Pawn {
		if (piece.Color == engine.White && mv.To.Rank == 7) || (piece.Color == engine.Black && mv.To.Rank == 0) {
			if !mv.HasExplicitPromotion() {
				m.awaitingPromotion = true
				m.pendingMove = mv
				m.msg = "Promotion required for " + moveString(mv) + ". Enter q/r/b/n."
				return m, nil
			}
		}
	}

	m = m.applyMove(mv)
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputWidth := msg.Width - 10
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.input.Width = inputWidth
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.awaitingPromotion {
				m.awaitingPromotion = false
				m.msg = "Promotion cancelled."
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyEnter:
			return m.submitInput()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func statusLine(state *engine.GameState) string {
	status := "in play"
	if engine.IsCheckmate(state) {
		status = "checkmate"
	} else if engine.IsStalemate(state) {
		status = "stalemate"
	} else if engine.IsInCheck(state, state.Turn) {
		status = "check"
	}
	return status
}

func (m model) renderCLI() string {
	var out strings.Builder
	out.WriteString("SwapChess\n\n")
	if m.renderer == renderView {
		out.WriteString(renderBoardFromView(m.vi, m.pieceCatalog))
	} else {
		out.WriteString(renderBoard(m.state, m.pieceCatalog))
	}
	out.WriteString(fmt.Sprintf("Turn: %s  |  Status: %s  |  Renderer: %s\n", m.state.Turn.String(), statusLine(m.state), m.renderer.String()))
	if m.awaitingPromotion {
		out.WriteString("Input mode: promotion choice pending (q/r/b/n)\n")
	} else {
		out.WriteString("Input mode: move/command\n")
	}
	out.WriteString("Type help for commands.\n\n")
	out.WriteString(m.msg)
	out.WriteString("\n\n")
	out.WriteString("> ")
	out.WriteString(m.input.View())
	return out.String()
}

func (m model) View() string {
	if m.width > 0 && m.width < 42 {
		return fmt.Sprintf("Terminal too narrow (%d cols). Need at least 42.\n\n> %s", m.width, m.input.View())
	}
	frame := m.renderCLI()
	if m.height <= 0 {
		return frame
	}
	lines := strings.Split(frame, "\n")
	if len(lines) >= m.height {
		return strings.Join(lines[:m.height], "\n")
	}
	pad := make([]string, m.height-len(lines))
	return strings.Join(append(lines, pad...), "\n")
}

func Run(withAltScreen bool) error {
	opts := []tea.ProgramOption{}
	if withAltScreen {
		opts = append(opts, tea.WithAltScreen())
	}
	p := tea.NewProgram(initialModel(), opts...)
	return p.Start()
}
