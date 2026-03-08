package app

import (
	"fmt"
	"strings"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/view"
)

type Mode string

const (
	ModeTUI Mode = "tui"
	ModeCLI Mode = "cli"
)

type InputMode string

const (
	InputModeCommand     InputMode = "command"
	InputModePromotion   InputMode = "promotion"
	InputModeBoardSelect InputMode = "board_select"
)

type RendererMode string

const (
	RendererView   RendererMode = "view"
	RendererEngine RendererMode = "engine"
)

type ActionResult struct {
	Quit       bool
	Message    string
	Hint       string
	InputMode  InputMode
	ClearInput bool
}

type MoveRecord struct {
	Index     int
	Player    engine.Color
	Move      engine.Move
	Notation  string
	SwapEvent *view.SwapEvent
}

type Session struct {
	Game      *engine.GameState
	View      view.ViewState
	Message   string
	Hint      string
	MoveLog   []MoveRecord
	InputMode InputMode
	Renderer  RendererMode
	Cursor    engine.Position
	Selected  *engine.Position
	Width     int
	Height    int

	DebugRendererEnabled bool

	history        []*engine.GameState
	pendingMove    engine.Move
	hasPendingMove bool
	lastMove       *engine.Move
	lastSwap       *view.SwapEvent
}

const (
	defaultPromptPlaceholder = "move: e2e4 | commands: help, undo, clear, quit"
	promotionPlaceholder     = "promotion: q/r/b/n"
)

func NewSession(debugRenderer string) *Session {
	session := &Session{
		Game:      engine.NewGame(),
		InputMode: InputModeCommand,
		Renderer:  RendererView,
		Cursor:    engine.Position{File: 4, Rank: 1},
		Message:   "Enter a move like e2e4. Type help for commands.",
	}

	switch RendererMode(strings.ToLower(strings.TrimSpace(debugRenderer))) {
	case RendererEngine:
		session.Renderer = RendererEngine
		session.DebugRendererEnabled = true
	case RendererView:
		if strings.TrimSpace(debugRenderer) != "" {
			session.DebugRendererEnabled = true
		}
	}

	session.refreshView()
	session.Hint = session.Preview("")
	return session
}

func (s *Session) Resize(width, height int) {
	s.Width = width
	s.Height = height
}

func (s *Session) Preview(raw string) string {
	s.Hint = s.preview(raw)
	return s.Hint
}

func (s *Session) Submit(raw string) ActionResult {
	value := strings.TrimSpace(raw)
	if value == "" {
		switch s.InputMode {
		case InputModePromotion:
			s.Message = "Promotion required. Enter q, r, b, or n."
		case InputModeBoardSelect:
			s.Message = "Selection active. Pick destination or Esc."
		default:
			s.Message = "Empty input. Enter a move like e2e4 or a command."
		}
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	if s.InputMode == InputModePromotion {
		pk, ok := promotionChoice(value)
		if !ok {
			s.Message = "Promotion required: enter q, r, b, or n."
			s.Hint = s.Preview(value)
			return s.result(false, false)
		}
		move := s.pendingMove
		move.Promotion = pk
		move.PromotionSet = true
		s.InputMode = InputModeCommand
		s.hasPendingMove = false
		s.pendingMove = engine.Move{}
		return s.applyMove(move)
	}

	command := normalizeCommand(value)
	switch command {
	case "help", "?":
		s.Message = "Commands: " + strings.Join(s.HelpLines(), " | ")
		s.Hint = s.Preview("")
		return s.result(false, true)
	case "undo", "u":
		return s.undo()
	case "clear":
		s.MoveLog = nil
		s.lastMove = nil
		s.lastSwap = nil
		s.refreshView()
		s.Message = "Move log cleared."
		s.Hint = s.Preview("")
		return s.result(false, true)
	case "quit", "exit":
		s.Message = "Quitting."
		s.Hint = s.Preview("")
		return s.result(true, true)
	case "renderer view", "render view", "view":
		return s.setRenderer(RendererView)
	case "renderer engine", "render engine", "engine":
		return s.setRenderer(RendererEngine)
	case "renderer toggle", "render toggle":
		return s.toggleRenderer()
	}

	move, err := parseMove(value)
	if err != nil {
		s.Message = "Parse error: " + err.Error()
		s.Hint = s.Preview(value)
		return s.result(false, false)
	}
	return s.submitMove(move)
}

func (s *Session) CancelTransient() ActionResult {
	clearInput := false
	switch s.InputMode {
	case InputModePromotion:
		s.InputMode = InputModeCommand
		s.hasPendingMove = false
		s.pendingMove = engine.Move{}
		s.Message = "Promotion cancelled."
		clearInput = true
	case InputModeBoardSelect:
		s.InputMode = InputModeCommand
		s.Selected = nil
		s.Message = "Selection cleared."
		clearInput = true
	default:
		s.Message = "Nothing to cancel."
	}
	s.Hint = s.Preview("")
	return s.result(false, clearInput)
}

func (s *Session) MoveCursor(df, dr int) {
	s.Cursor.File = clamp(s.Cursor.File+df, 0, 7)
	s.Cursor.Rank = clamp(s.Cursor.Rank+dr, 0, 7)
}

func (s *Session) ActivateCursor() ActionResult {
	if s.InputMode == InputModePromotion {
		s.Message = "Promotion is pending. Enter q, r, b, or n in the prompt."
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	cursor := s.Cursor
	piece := s.Game.Board.Squares[cursor.File][cursor.Rank]
	if s.InputMode != InputModeBoardSelect || s.Selected == nil {
		if piece == nil {
			s.Message = "No piece at " + PositionString(cursor) + "."
			s.Hint = s.Preview("")
			return s.result(false, false)
		}
		if piece.Color != s.Game.Turn {
			s.Message = "Select one of your own pieces."
			s.Hint = s.Preview("")
			return s.result(false, false)
		}
		selected := cursor
		s.Selected = &selected
		s.InputMode = InputModeBoardSelect
		s.Message = fmt.Sprintf("Selected %s %s. Choose destination or Esc.", strings.ToLower(piece.Kind.String()), PositionString(cursor))
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	from := *s.Selected
	if from == cursor {
		s.Selected = nil
		s.InputMode = InputModeCommand
		s.Message = "Selection cleared."
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	move := engine.Move{From: from, To: cursor}
	return s.submitMove(move)
}

func (s *Session) PromptLabel() string {
	if s.InputMode == InputModePromotion {
		return "promo"
	}
	return "move"
}

func (s *Session) PromptPlaceholder() string {
	if s.InputMode == InputModePromotion {
		return promotionPlaceholder
	}
	if s.DebugRendererEnabled {
		return "move: e2e4 | commands: help, undo, clear, renderer toggle, quit"
	}
	return defaultPromptPlaceholder
}

func (s *Session) ModeLabel() string {
	switch s.InputMode {
	case InputModePromotion:
		return "promotion choice"
	case InputModeBoardSelect:
		return "board selection"
	default:
		return "move/command"
	}
}

func (s *Session) HelpLines() []string {
	lines := []string{
		"help",
		"undo",
		"clear",
		"quit",
		"move e2e4",
		"promotion e7e8q",
	}
	if s.DebugRendererEnabled {
		lines = append(lines, "debug: renderer view|engine|toggle")
	}
	return lines
}

func (s *Session) LastMoveNotation() string {
	if s.lastMove == nil {
		return "-"
	}
	return MoveString(*s.lastMove)
}

func (s *Session) SelectedSquare() *engine.Position {
	if s.Selected == nil {
		return nil
	}
	selected := *s.Selected
	return &selected
}

func (s *Session) preview(raw string) string {
	value := strings.TrimSpace(raw)
	switch s.InputMode {
	case InputModePromotion:
		if value == "" {
			return "Promotion pending. Enter q, r, b, or n and press Enter."
		}
		if pk, ok := promotionChoice(value); ok {
			return "Press Enter to promote to " + strings.ToLower(pk.String()) + "."
		}
		return "Invalid promotion piece. Valid values: q, r, b, n."
	case InputModeBoardSelect:
		if value == "" {
			return "Select destination, type move, or Esc."
		}
	}

	if value == "" {
		return "Enter move (e2e4 / e7e8q) or command (help / undo / clear / quit)."
	}

	command := normalizeCommand(value)
	if recognizedCommand(command, s.DebugRendererEnabled) {
		if strings.HasPrefix(command, "renderer") || strings.HasPrefix(command, "render") || command == "view" || command == "engine" {
			if !s.DebugRendererEnabled {
				return "Renderer commands are disabled unless launched with --debug-renderer."
			}
		}
		return "Press Enter to run command: " + command
	}

	move, err := parseMove(value)
	if err != nil {
		return "Input not recognized. Examples: e2e4, e7e8q, undo, clear."
	}
	if err := validateMoveContext(s.Game, move); err != nil {
		return "Move issue: " + err.Error()
	}

	piece := s.Game.Board.Squares[move.From.File][move.From.Rank]
	if piece != nil && piece.Kind == engine.Pawn {
		if (piece.Color == engine.White && move.To.Rank == 7) || (piece.Color == engine.Black && move.To.Rank == 0) {
			if !move.HasExplicitPromotion() {
				return "Promotion is required for this move. Enter it as e7e8q or submit the move and choose q/r/b/n."
			}
		}
	}

	return "Move syntax and context look valid. Press Enter to apply."
}

func (s *Session) submitMove(move engine.Move) ActionResult {
	if err := validateMoveContext(s.Game, move); err != nil {
		s.Message = "Invalid move context: " + err.Error()
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	piece := s.Game.Board.Squares[move.From.File][move.From.Rank]
	if piece != nil && piece.Kind == engine.Pawn {
		if (piece.Color == engine.White && move.To.Rank == 7) || (piece.Color == engine.Black && move.To.Rank == 0) {
			if !move.HasExplicitPromotion() {
				s.InputMode = InputModePromotion
				s.pendingMove = move
				s.hasPendingMove = true
				s.Message = "Promotion required for " + MoveString(move) + ". Enter q/r/b/n."
				s.Hint = s.Preview("")
				return s.result(false, true)
			}
		}
	}

	return s.applyMove(move)
}

func (s *Session) applyMove(move engine.Move) ActionResult {
	previous := s.Game.Clone()
	movedPiece := s.Game.Board.Squares[move.From.File][move.From.Rank]
	mover := s.Game.Turn

	if err := engine.ApplyMove(s.Game, move); err != nil {
		s.Message = "Illegal move: " + err.Error()
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	s.history = append(s.history, previous)

	swapEvent := detectSwapEvent(s.Game, movedPiece, move.To)
	record := MoveRecord{
		Index:     len(s.MoveLog) + 1,
		Player:    mover,
		Move:      move,
		Notation:  MoveString(move),
		SwapEvent: cloneSwapEvent(swapEvent),
	}
	s.MoveLog = append(s.MoveLog, record)

	moveCopy := move
	s.lastMove = &moveCopy
	s.lastSwap = cloneSwapEvent(swapEvent)
	s.InputMode = InputModeCommand
	s.Selected = nil
	s.hasPendingMove = false
	s.pendingMove = engine.Move{}
	s.refreshView()

	if swapEvent != nil {
		s.Message = fmt.Sprintf("Move applied: %s (swap %s <-> %s)", record.Notation, PositionString(swapEvent.A), PositionString(swapEvent.B))
	} else {
		s.Message = "Move applied: " + record.Notation
	}
	s.Hint = s.Preview("")
	return s.result(false, true)
}

func (s *Session) undo() ActionResult {
	if len(s.history) == 0 {
		s.Message = "No moves to undo."
		s.Hint = s.Preview("")
		return s.result(false, false)
	}

	last := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.Game = last
	if len(s.MoveLog) > 0 {
		s.MoveLog = s.MoveLog[:len(s.MoveLog)-1]
	}

	if len(s.MoveLog) > 0 {
		record := s.MoveLog[len(s.MoveLog)-1]
		moveCopy := record.Move
		s.lastMove = &moveCopy
		s.lastSwap = cloneSwapEvent(record.SwapEvent)
	} else {
		s.lastMove = nil
		s.lastSwap = nil
	}

	s.InputMode = InputModeCommand
	s.Selected = nil
	s.hasPendingMove = false
	s.pendingMove = engine.Move{}
	s.refreshView()
	s.Message = "Undid last move."
	s.Hint = s.Preview("")
	return s.result(false, true)
}

func (s *Session) setRenderer(mode RendererMode) ActionResult {
	if !s.DebugRendererEnabled {
		s.Message = "Renderer commands are disabled unless launched with --debug-renderer."
		s.Hint = s.Preview("")
		return s.result(false, false)
	}
	s.Renderer = mode
	s.Message = "Renderer: " + string(mode)
	s.Hint = s.Preview("")
	return s.result(false, true)
}

func (s *Session) toggleRenderer() ActionResult {
	if !s.DebugRendererEnabled {
		s.Message = "Renderer commands are disabled unless launched with --debug-renderer."
		s.Hint = s.Preview("")
		return s.result(false, false)
	}
	if s.Renderer == RendererView {
		return s.setRenderer(RendererEngine)
	}
	return s.setRenderer(RendererView)
}

func (s *Session) refreshView() {
	s.View = view.ViewStateFromGameStateWithMeta(s.Game, view.SnapshotMeta{
		LastMove:  s.lastMove,
		SwapEvent: s.lastSwap,
	})
}

func (s *Session) result(quit, clearInput bool) ActionResult {
	return ActionResult{
		Quit:       quit,
		Message:    s.Message,
		Hint:       s.Hint,
		InputMode:  s.InputMode,
		ClearInput: clearInput,
	}
}

func MoveString(move engine.Move) string {
	base := fmt.Sprintf("%c%d%c%d", byte('a'+move.From.File), move.From.Rank+1, byte('a'+move.To.File), move.To.Rank+1)
	if !move.HasExplicitPromotion() {
		return base
	}

	switch move.Promotion {
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

func PositionString(pos engine.Position) string {
	return fmt.Sprintf("%c%d", byte('a'+pos.File), pos.Rank+1)
}

func parseMove(raw string) (engine.Move, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return engine.Move{}, fmt.Errorf("empty move")
	}

	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "->", "")
	value = strings.ReplaceAll(value, "-", "")

	if len(value) != 4 && len(value) != 5 {
		return engine.Move{}, fmt.Errorf("invalid move format; expected e2e4 or e7e8q")
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

	fromFile, err := fileOf(value[0])
	if err != nil {
		return engine.Move{}, err
	}
	fromRank, err := rankOf(value[1])
	if err != nil {
		return engine.Move{}, err
	}
	toFile, err := fileOf(value[2])
	if err != nil {
		return engine.Move{}, err
	}
	toRank, err := rankOf(value[3])
	if err != nil {
		return engine.Move{}, err
	}

	move := engine.Move{
		From: engine.Position{File: fromFile, Rank: fromRank},
		To:   engine.Position{File: toFile, Rank: toRank},
	}

	if len(value) == 5 {
		switch value[4] {
		case 'q':
			move.Promotion = engine.Queen
		case 'r':
			move.Promotion = engine.Rook
		case 'b':
			move.Promotion = engine.Bishop
		case 'n':
			move.Promotion = engine.Knight
		default:
			return engine.Move{}, fmt.Errorf("unknown promotion piece: %c", value[4])
		}
		move.PromotionSet = true
	}

	return move, nil
}

func promotionChoice(raw string) (engine.PieceKind, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
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

func validateMoveContext(state *engine.GameState, move engine.Move) error {
	if move.From == move.To {
		return fmt.Errorf("source and destination are the same square")
	}

	piece := state.Board.Squares[move.From.File][move.From.Rank]
	if piece == nil {
		return fmt.Errorf("no piece at %s", PositionString(move.From))
	}
	if piece.Color != state.Turn {
		return fmt.Errorf("it is %s to move; %s has %s piece", state.Turn.String(), PositionString(move.From), piece.Color.String())
	}

	dest := state.Board.Squares[move.To.File][move.To.Rank]
	if dest != nil && dest.Color == piece.Color {
		return fmt.Errorf("destination %s contains your own %s", PositionString(move.To), strings.ToLower(dest.Kind.String()))
	}

	return nil
}

func normalizeCommand(raw string) string {
	value := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
	if after, ok := strings.CutPrefix(value, ":"); ok {
		value = strings.TrimSpace(after)
	}
	return value
}

func recognizedCommand(command string, debugEnabled bool) bool {
	switch command {
	case "help", "?", "undo", "u", "clear", "quit", "exit":
		return true
	case "renderer view", "render view", "view",
		"renderer engine", "render engine", "engine",
		"renderer toggle", "render toggle":
		return debugEnabled
	default:
		return false
	}
}

func detectSwapEvent(state *engine.GameState, movedPiece *engine.Piece, destination engine.Position) *view.SwapEvent {
	if movedPiece == nil {
		return nil
	}

	for file := 0; file < 8; file++ {
		for rank := 0; rank < 8; rank++ {
			if state.Board.Squares[file][rank] == movedPiece {
				finalPosition := engine.Position{File: file, Rank: rank}
				if finalPosition == destination {
					return nil
				}
				return &view.SwapEvent{A: destination, B: finalPosition}
			}
		}
	}

	return nil
}

func cloneSwapEvent(event *view.SwapEvent) *view.SwapEvent {
	if event == nil {
		return nil
	}
	copy := *event
	return &copy
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
