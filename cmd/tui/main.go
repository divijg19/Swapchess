package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/pieces"
	"github.com/divijg19/Swapchess/view"
)

// simple TUI model
type model struct {
	state        *engine.GameState
	vi           view.ViewState
	input        textinput.Model
	msg          string
	history      []*engine.GameState
	useView      bool
	pieceCatalog *pieces.Catalog
	// promotion flow
	awaitingPromotion bool
	pendingMove       engine.Move
}

func initialModel() model {
	// start with a standard starting GameState
	st := engine.NewGame()
	// default castling rights for standard starting position can be set by UI later
	m := model{state: st}
	m.vi = view.ViewStateFromGameState(st)
	m.useView = true
	m.input = textinput.New()
	m.input.Placeholder = "'e2e4' or 'a2 a4'"
	m.input.Focus()
	m.input.CharLimit = 10
	m.input.Width = 20
	m.msg = "Type moves like e2e4 and press Enter"
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.awaitingPromotion {
			if msg.Type == tea.KeyEsc {
				m.awaitingPromotion = false
				m.msg = "Promotion cancelled"
				return m, nil
			}
			k := strings.ToLower(msg.String())
			switch k {
			case "q", "r", "b", "n":
				var pk engine.PieceKind
				switch k {
				case "q":
					pk = engine.Queen
				case "r":
					pk = engine.Rook
				case "b":
					pk = engine.Bishop
				case "n":
					pk = engine.Knight
				}
				m.pendingMove.Promotion = pk
				m.pendingMove.PromotionSet = true
				prev := m.state.Clone()
				if err := engine.ApplyMove(m.state, m.pendingMove); err != nil {
					m.msg = "Illegal promotion move: " + err.Error()
					m.awaitingPromotion = false
					return m, nil
				}
				m.history = append(m.history, prev)
				m.vi = view.ViewStateFromGameState(m.state)
				m.msg = "Promotion applied"
				m.awaitingPromotion = false
				return m, nil
			case "c":
				m.awaitingPromotion = false
				m.msg = "Promotion cancelled"
				return m, nil
			default:
				m.msg = "Choose promotion: q/r/b/n (esc to cancel)"
				return m, nil
			}
		}

		if msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

		if msg.String() == "u" || msg.String() == "U" {
			// undo
			if len(m.history) > 0 {
				last := m.history[len(m.history)-1]
				m.history = m.history[:len(m.history)-1]
				m.state = last
				m.vi = view.ViewStateFromGameState(m.state)
				m.msg = "Undid last move"
			} else {
				m.msg = "No moves to undo"
			}
			return m, nil
		}
		if msg.String() == "r" || msg.String() == "R" {
			m.useView = !m.useView
			if m.useView {
				m.msg = "Rendering: view model"
			} else {
				m.msg = "Rendering: engine board"
			}
			return m, nil
		}
		if msg.Type == tea.KeyEnter {
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				m.msg = "Please enter a move (e2e4)"
				return m, nil
			}
			m.input.SetValue("")
			mv, err := parseMove(val)
			if err != nil {
				m.msg = "Parse error: " + err.Error()
				return m, nil
			}
			// detect pawn promotion that needs a choice
			piece := m.state.Board.Squares[mv.From.File][mv.From.Rank]
			if piece != nil && piece.Kind == engine.Pawn {
				if (piece.Color == engine.White && mv.To.Rank == 7) || (piece.Color == engine.Black && mv.To.Rank == 0) {
					if !mv.HasExplicitPromotion() {
						// prompt user to choose promotion
						m.awaitingPromotion = true
						m.pendingMove = mv
						m.msg = "Promotion: press q (queen), r (rook), b (bishop), n (knight), esc/c to cancel"
						return m, nil
					}
				}
			}
			prev := m.state.Clone()
			if err := engine.ApplyMove(m.state, mv); err != nil {
				m.msg = "Illegal move: " + err.Error()
				return m, nil
			}
			m.history = append(m.history, prev)
			m.vi = view.ViewStateFromGameState(m.state)
			m.msg = "Move applied"
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	out := "SwapChess TUI\n\n"
	if m.useView {
		out += renderBoardFromView(m.vi, m.pieceCatalog)
	} else {
		out += renderBoard(m.state, m.pieceCatalog)
	}

	status := "in play"
	if engine.IsCheckmate(m.state) {
		status = "checkmate"
	} else if engine.IsStalemate(m.state) {
		status = "stalemate"
	} else if engine.IsInCheck(m.state, m.state.Turn) {
		status = "check"
	}

	out += fmt.Sprintf("Turn: %s  |  Status: %s  |  Renderer: %s\n", m.state.Turn.String(), status, map[bool]string{true: "view", false: "engine"}[m.useView])
	out += "Commands: Enter=apply  u=undo  r=toggle renderer  esc=quit\n\n"
	out += m.msg + "\n"
	out += m.input.View()
	return out
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error starting TUI:", err)
		os.Exit(1)
	}
}
