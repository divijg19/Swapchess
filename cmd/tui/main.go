package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/view"
)

// simple TUI model
type model struct {
	state       *engine.GameState
	vi          view.ViewState
	input       textinput.Model
	msg         string
	history     []*engine.GameState
	useView     bool
	pieceAssets map[string]string
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
	// load piece glyph assets (optional); fallback to letters if missing
	m.pieceAssets = loadPieceAssets(filepath.Join("assets", "pieces"))
	return m
}

func loadPieceAssets(dir string) map[string]string {
	assets := make(map[string]string)
	colors := []struct{ code, name string }{{"w", "white"}, {"b", "black"}}
	kinds := []string{"pawn", "knight", "bishop", "rook", "queen", "king"}
	for _, c := range colors {
		for _, k := range kinds {
			fname := filepath.Join(dir, fmt.Sprintf("%s_%s.txt", c.name, k))
			data, err := os.ReadFile(fname)
			if err != nil {
				continue
			}
			s := strings.TrimSpace(string(data))
			if s == "" {
				continue
			}
			assets[fmt.Sprintf("%s_%s", c.code, k)] = s
		}
	}
	return assets
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func pieceGlyphFromPiece(p *engine.Piece, assets map[string]string) string {
	if p == nil {
		return "."
	}
	kind := "?"
	switch p.Kind {
	case engine.Pawn:
		kind = "pawn"
	case engine.Knight:
		kind = "knight"
	case engine.Bishop:
		kind = "bishop"
	case engine.Rook:
		kind = "rook"
	case engine.Queen:
		kind = "queen"
	case engine.King:
		kind = "king"
	}
	colorCode := "b"
	if p.Color == engine.White {
		colorCode = "w"
	}
	key := fmt.Sprintf("%s_%s", colorCode, kind)
	if g, ok := assets[key]; ok {
		return g
	}
	// fallback single-letter glyph
	ch := '?'
	switch p.Kind {
	case engine.Pawn:
		ch = 'p'
	case engine.Knight:
		ch = 'n'
	case engine.Bishop:
		ch = 'b'
	case engine.Rook:
		ch = 'r'
	case engine.Queen:
		ch = 'q'
	case engine.King:
		ch = 'k'
	}
	if p.Color == engine.White {
		return strings.ToUpper(string(ch))
	}
	return string(ch)
}

func renderBoard(s *engine.GameState, assets map[string]string) string {
	var b [8][8]string
	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := s.Board.Squares[f][r]
			if p == nil {
				b[r][f] = "."
				continue
			}
			b[r][f] = pieceGlyphFromPiece(p, assets)
		}
	}
	out := ""
	for rank := 7; rank >= 0; rank-- {
		out += fmt.Sprintf("%d ", rank+1)
		for file := 0; file < 8; file++ {
			out += b[rank][file] + " "
		}
		out += "\n"
	}
	out += "  a b c d e f g h\n"
	return out
}

func renderBoardFromView(v view.ViewState, assets map[string]string) string {
	var b [8][8]string
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			b[r][f] = "."
		}
	}
	for _, vp := range v.Pieces {
		// map view piece to glyph
		kind := "?"
		switch vp.Kind {
		case engine.Pawn:
			kind = "pawn"
		case engine.Knight:
			kind = "knight"
		case engine.Bishop:
			kind = "bishop"
		case engine.Rook:
			kind = "rook"
		case engine.Queen:
			kind = "queen"
		case engine.King:
			kind = "king"
		}
		colorCode := "b"
		if vp.Color == engine.White {
			colorCode = "w"
		}
		key := fmt.Sprintf("%s_%s", colorCode, kind)
		glyph := "."
		if g, ok := assets[key]; ok {
			glyph = g
		} else {
			// fallback single-letter
			ch := '?'
			switch vp.Kind {
			case engine.Pawn:
				ch = 'p'
			case engine.Knight:
				ch = 'n'
			case engine.Bishop:
				ch = 'b'
			case engine.Rook:
				ch = 'r'
			case engine.Queen:
				ch = 'q'
			case engine.King:
				ch = 'k'
			}
			if vp.Color == engine.White {
				glyph = strings.ToUpper(string(ch))
			} else {
				glyph = string(ch)
			}
		}
		if vp.Y >= 0 && vp.Y < 8 && vp.X >= 0 && vp.X < 8 {
			b[vp.Y][vp.X] = glyph
		}
	}
	out := ""
	for rank := 7; rank >= 0; rank-- {
		out += fmt.Sprintf("%d ", rank+1)
		for file := 0; file < 8; file++ {
			out += b[rank][file] + " "
		}
		out += "\n"
	}
	out += "  a b c d e f g h\n"
	return out
}

func parseMove(s string) (engine.Move, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return engine.Move{}, fmt.Errorf("empty move")
	}
	// normalize common separators and casing
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "->", "")

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
	}
	return mv, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
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
					if mv.Promotion == 0 {
						// prompt user to choose promotion
						m.awaitingPromotion = true
						m.pendingMove = mv
						m.msg = "Promotion: press q (queen), r (rook), b (bishop), n (knight)"
						return m, nil
					}
				}
			}
			// push current state to history
			m.history = append(m.history, m.state.Clone())

			if err := engine.ApplyMove(m.state, mv); err != nil {
				m.msg = "Illegal move: " + err.Error()
				return m, nil
			}
			m.vi = view.ViewStateFromGameState(m.state)
			m.msg = "Move applied"
			return m, nil
		}
		// promotion choice handling
		if m.awaitingPromotion {
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
				// push history and apply
				m.history = append(m.history, m.state.Clone())
				if err := engine.ApplyMove(m.state, m.pendingMove); err != nil {
					m.msg = "Illegal promotion move: " + err.Error()
					m.awaitingPromotion = false
					return m, nil
				}
				m.vi = view.ViewStateFromGameState(m.state)
				m.msg = "Promotion applied"
				m.awaitingPromotion = false
				return m, nil
			case "c", "esc":
				m.awaitingPromotion = false
				m.msg = "Promotion cancelled"
				return m, nil
			default:
				m.msg = "Choose promotion: q/r/b/n"
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	out := "SwapChess TUI\n"
	if m.useView {
		out += renderBoardFromView(m.vi, m.pieceAssets)
	} else {
		out += renderBoard(m.state, m.pieceAssets)
	}
	out += "Turn: " + m.state.Turn.String() + "\n"
	out += "\n"
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
