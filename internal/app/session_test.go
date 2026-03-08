package app

import (
	"strings"
	"testing"

	"github.com/divijg19/Swapchess/engine"
)

func TestParseMoveNormalizesInput(t *testing.T) {
	move, err := parseMove(" E2 -> E4 ")
	if err != nil {
		t.Fatalf("parseMove returned error: %v", err)
	}

	if move.From != (engine.Position{File: 4, Rank: 1}) || move.To != (engine.Position{File: 4, Rank: 3}) {
		t.Fatalf("unexpected move parsed: %+v", move)
	}
}

func TestParseMovePromotionAndErrors(t *testing.T) {
	move, err := parseMove("e7-e8n")
	if err != nil {
		t.Fatalf("parseMove returned error: %v", err)
	}
	if !move.HasExplicitPromotion() || move.Promotion != engine.Knight {
		t.Fatalf("expected explicit knight promotion, got %+v", move)
	}

	if _, err := parseMove("e2e9"); err == nil {
		t.Fatalf("expected parse error for out-of-range rank")
	}
	if _, err := parseMove("e7e8x"); err == nil {
		t.Fatalf("expected parse error for unknown promotion piece")
	}
}

func TestNormalizeCommandAndRecognition(t *testing.T) {
	if got := normalizeCommand(":  Undo "); got != "undo" {
		t.Fatalf("expected normalized undo command, got %q", got)
	}
	if recognizedCommand("renderer toggle", false) {
		t.Fatalf("renderer commands should be hidden without debug flag")
	}
	if !recognizedCommand("renderer toggle", true) {
		t.Fatalf("renderer commands should be available with debug flag")
	}
}

func TestPreviewHintsCoverCommonStates(t *testing.T) {
	session := NewSession("")

	if got := session.Preview(""); got == "" {
		t.Fatalf("expected non-empty hint for empty input")
	}
	if got := session.Preview("a7a6"); got == "" || got == session.Preview("") {
		t.Fatalf("expected specific hint for wrong-side move, got %q", got)
	}
	if got := session.Preview("e2e4"); got != "Move syntax and context look valid. Press Enter to apply." {
		t.Fatalf("unexpected hint for valid move: %q", got)
	}
	if got := session.Preview("renderer toggle"); got != "Input not recognized. Examples: e2e4, e7e8q, undo, clear." {
		t.Fatalf("expected debug command to stay hidden without debug flag, got %q", got)
	}
}

func TestPreviewPromotionStates(t *testing.T) {
	session := promotionReadySession()
	session.Submit("e7e8")

	if got := session.Preview(""); got != "Promotion pending. Enter q, r, b, or n and press Enter." {
		t.Fatalf("unexpected empty promotion hint: %q", got)
	}
	if got := session.Preview("rook"); got != "Press Enter to promote to rook." {
		t.Fatalf("unexpected valid promotion hint: %q", got)
	}
	if got := session.Preview("x"); got != "Invalid promotion piece. Valid values: q, r, b, n." {
		t.Fatalf("unexpected invalid promotion hint: %q", got)
	}
}

func TestSubmitLegalMoveAndUndo(t *testing.T) {
	session := NewSession("")

	result := session.Submit("e2e4")
	if result.Quit {
		t.Fatalf("did not expect quit result")
	}
	if session.Game.Turn != engine.Black {
		t.Fatalf("expected black to move after e2e4")
	}
	if len(session.MoveLog) != 1 {
		t.Fatalf("expected move log length 1, got %d", len(session.MoveLog))
	}

	session.Submit("undo")
	if session.Game.Turn != engine.White {
		t.Fatalf("expected white to move after undo")
	}
	if len(session.MoveLog) != 0 {
		t.Fatalf("expected move log to be empty after undo")
	}
}

func TestSubmitPromotionFlow(t *testing.T) {
	session := NewSession("")
	session.Game = &engine.GameState{Turn: engine.White}
	session.Game.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	session.Game.Board.Squares[7][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	session.Game.Board.Squares[4][6] = &engine.Piece{Kind: engine.Pawn, Color: engine.White}
	session.refreshView()

	session.Submit("e7e8")
	if session.InputMode != InputModePromotion {
		t.Fatalf("expected promotion input mode, got %s", session.InputMode)
	}
	if session.PromptLabel() != "promo" {
		t.Fatalf("expected promo prompt label during promotion mode")
	}

	session.Submit("q")
	if session.InputMode != InputModeCommand {
		t.Fatalf("expected input mode to return to command, got %s", session.InputMode)
	}
	piece := session.Game.Board.Squares[4][7]
	if piece == nil || piece.Kind != engine.Queen || piece.Color != engine.White {
		t.Fatalf("expected promoted queen on e8, got %+v", piece)
	}
}

func TestSubmitCommands(t *testing.T) {
	session := NewSession("engine")

	session.Submit("renderer view")
	if session.Renderer != RendererView {
		t.Fatalf("expected renderer to switch to view")
	}

	session.Submit("clear")
	if len(session.MoveLog) != 0 {
		t.Fatalf("expected move log to remain empty after clear")
	}

	result := session.Submit("quit")
	if !result.Quit {
		t.Fatalf("expected quit action result")
	}
}

func TestSubmitRendererCommandsRespectDebugFlag(t *testing.T) {
	disabled := NewSession("")
	disabled.Submit("renderer engine")
	if disabled.Renderer != RendererView {
		t.Fatalf("renderer should remain view when debug flag is disabled")
	}
	if disabled.Message != "Renderer commands are disabled unless launched with --debug-renderer." {
		t.Fatalf("unexpected disabled renderer message: %q", disabled.Message)
	}

	enabled := NewSession("view")
	enabled.Submit("renderer toggle")
	if enabled.Renderer != RendererEngine {
		t.Fatalf("expected renderer toggle to switch to engine, got %s", enabled.Renderer)
	}
}

func TestCancelTransientResetsModes(t *testing.T) {
	session := promotionReadySession()
	session.Submit("e7e8")
	session.CancelTransient()
	if session.InputMode != InputModeCommand {
		t.Fatalf("expected command mode after cancelling promotion, got %s", session.InputMode)
	}
	if session.hasPendingMove {
		t.Fatalf("expected pending move to be cleared")
	}

	selected := engine.Position{File: 4, Rank: 1}
	session.Selected = &selected
	session.InputMode = InputModeBoardSelect
	session.CancelTransient()
	if session.Selected != nil {
		t.Fatalf("expected selection to be cleared")
	}
	if session.Message != "Selection cleared." {
		t.Fatalf("unexpected board cancel message: %q", session.Message)
	}
}

func TestActivateCursorRejectsEmptyAndOpponentSquares(t *testing.T) {
	session := NewSession("")

	session.Cursor = engine.Position{File: 4, Rank: 3}
	session.ActivateCursor()
	if session.Message != "No piece at e4." {
		t.Fatalf("unexpected empty-square message: %q", session.Message)
	}

	session.Cursor = engine.Position{File: 4, Rank: 6}
	session.ActivateCursor()
	if session.Message != "Select one of your own pieces." {
		t.Fatalf("unexpected opponent-square message: %q", session.Message)
	}
}

func TestActivateCursorSelectionAndMove(t *testing.T) {
	session := NewSession("")

	result := session.ActivateCursor()
	if result.InputMode != InputModeBoardSelect {
		t.Fatalf("expected board selection mode after selecting piece, got %s", result.InputMode)
	}
	if selected := session.SelectedSquare(); selected == nil || *selected != (engine.Position{File: 4, Rank: 1}) {
		t.Fatalf("expected e2 to be selected, got %+v", selected)
	}
	if !strings.Contains(session.Message, "Selected pawn e2.") {
		t.Fatalf("unexpected selection message: %q", session.Message)
	}

	session.MoveCursor(0, 2)
	result = session.ActivateCursor()
	if result.Quit {
		t.Fatalf("did not expect quit while applying move")
	}
	if session.InputMode != InputModeCommand {
		t.Fatalf("expected command mode after applying move, got %s", session.InputMode)
	}
	if session.Selected != nil {
		t.Fatalf("expected selection to be cleared after move")
	}
	if len(session.MoveLog) != 1 {
		t.Fatalf("expected one move in log, got %d", len(session.MoveLog))
	}
}

func TestClearResetsMoveMetadata(t *testing.T) {
	session := swapReadySession()
	session.Submit("a2a3")
	if len(session.MoveLog) != 1 {
		t.Fatalf("expected move log entry before clear")
	}
	if session.View.LastMove == nil {
		t.Fatalf("expected last move metadata before clear")
	}

	session.Submit("clear")
	if len(session.MoveLog) != 0 {
		t.Fatalf("expected move log to be empty after clear")
	}
	if session.LastMoveNotation() != "-" {
		t.Fatalf("expected cleared last move notation, got %q", session.LastMoveNotation())
	}
	if session.View.LastMove != nil {
		t.Fatalf("expected view last move metadata to be cleared")
	}
	if session.View.SwapEvent != nil {
		t.Fatalf("expected view swap metadata to be cleared")
	}
}

func TestMoveRecordCapturesSwapEvent(t *testing.T) {
	session := swapReadySession()
	session.Submit("a2a3")

	if len(session.MoveLog) != 1 {
		t.Fatalf("expected one move record, got %d", len(session.MoveLog))
	}
	if session.MoveLog[0].SwapEvent == nil {
		t.Fatalf("expected swap metadata in move log")
	}
	if session.View.SwapEvent == nil {
		t.Fatalf("expected swap metadata in view snapshot")
	}
	if !strings.Contains(session.Message, "(swap ") {
		t.Fatalf("expected swap message, got %q", session.Message)
	}
}

func TestUndoWithoutHistory(t *testing.T) {
	session := NewSession("")
	result := session.Submit("undo")

	if result.Quit {
		t.Fatalf("did not expect quit on empty undo")
	}
	if session.Message != "No moves to undo." {
		t.Fatalf("unexpected undo message: %q", session.Message)
	}
}

func TestHelpLinesExposeDebugOnlyWhenEnabled(t *testing.T) {
	plain := strings.Join(NewSession("").HelpLines(), " | ")
	if strings.Contains(plain, "debug:") {
		t.Fatalf("unexpected debug help in plain session: %q", plain)
	}

	debug := strings.Join(NewSession("engine").HelpLines(), " | ")
	if !strings.Contains(debug, "debug: renderer view|engine|toggle") {
		t.Fatalf("expected debug help in debug session, got %q", debug)
	}
}

func promotionReadySession() *Session {
	session := NewSession("")
	session.Game = &engine.GameState{Turn: engine.White, RandSeed: 1}
	session.Game.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	session.Game.Board.Squares[7][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	session.Game.Board.Squares[4][6] = &engine.Piece{Kind: engine.Pawn, Color: engine.White}
	session.refreshView()
	session.Hint = session.Preview("")
	return session
}

func swapReadySession() *Session {
	session := NewSession("")
	session.Game = &engine.GameState{Turn: engine.White, RandSeed: 1}
	session.Game.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	session.Game.Board.Squares[0][1] = &engine.Piece{Kind: engine.Rook, Color: engine.White}
	session.Game.Board.Squares[2][2] = &engine.Piece{Kind: engine.Knight, Color: engine.White}
	session.Game.Board.Squares[7][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	session.Cursor = engine.Position{File: 0, Rank: 1}
	session.refreshView()
	session.Hint = session.Preview("")
	return session
}
