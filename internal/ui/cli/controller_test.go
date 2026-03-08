package cli

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/app"
	"github.com/divijg19/Swapchess/view"
)

type fakeTerminal struct {
	width   int
	height  int
	events  []KeyEvent
	renders []Frame
	closed  bool
	index   int
}

func (f *fakeTerminal) Size() (int, int, error) {
	return f.width, f.height, nil
}

func (f *fakeTerminal) NextEvent() (KeyEvent, error) {
	if f.index >= len(f.events) {
		return KeyEvent{}, io.EOF
	}
	event := f.events[f.index]
	f.index++
	return event, nil
}

func (f *fakeTerminal) Render(frame Frame) error {
	f.renders = append(f.renders, frame)
	return nil
}

func (f *fakeTerminal) Close() error {
	f.closed = true
	return nil
}

func TestControllerTypingUpdatesHintsLive(t *testing.T) {
	width, height := fullMinimumSize()
	terminal := &fakeTerminal{
		width:  width,
		height: height,
		events: []KeyEvent{
			{Kind: KeyText, Text: "e2e4"},
			{Kind: KeyQuit},
		},
	}
	controller := newController(terminal, "")

	if err := controller.Run(); err != nil {
		t.Fatalf("controller run returned error: %v", err)
	}
	if !terminal.closed {
		t.Fatalf("expected terminal to be closed")
	}
	if controller.session.Hint != "Move syntax and context look valid. Press Enter to apply." {
		t.Fatalf("unexpected hint after typing: %q", controller.session.Hint)
	}
	if controller.editor.String() != "e2e4" {
		t.Fatalf("expected editor to preserve typed input, got %q", controller.editor.String())
	}
}

func TestControllerInvalidSubmitPreservesInput(t *testing.T) {
	width, height := fullMinimumSize()
	terminal := &fakeTerminal{
		width:  width,
		height: height,
		events: []KeyEvent{
			{Kind: KeyText, Text: "e2e5"},
			{Kind: KeySubmit},
			{Kind: KeyQuit},
		},
	}
	controller := newController(terminal, "")

	if err := controller.Run(); err != nil {
		t.Fatalf("controller run returned error: %v", err)
	}
	if controller.editor.String() != "e2e5" {
		t.Fatalf("expected invalid input to remain editable, got %q", controller.editor.String())
	}
	if !strings.Contains(controller.session.Message, "Illegal move:") {
		t.Fatalf("expected illegal move message, got %q", controller.session.Message)
	}
}

func TestControllerLegalSubmitClearsInputAndAdvancesGame(t *testing.T) {
	width, height := fullMinimumSize()
	terminal := &fakeTerminal{
		width:  width,
		height: height,
		events: []KeyEvent{
			{Kind: KeyText, Text: "e2e4"},
			{Kind: KeySubmit},
			{Kind: KeyQuit},
		},
	}
	controller := newController(terminal, "")

	if err := controller.Run(); err != nil {
		t.Fatalf("controller run returned error: %v", err)
	}
	if controller.session.Game.Turn != engine.Black {
		t.Fatalf("expected black to move after legal submit, got %s", controller.session.Game.Turn)
	}
	if controller.editor.String() != "" {
		t.Fatalf("expected editor to clear after legal submit, got %q", controller.editor.String())
	}
}

func TestControllerBackspaceEditThenSubmitAppliesMove(t *testing.T) {
	width, height := fullMinimumSize()
	terminal := &fakeTerminal{
		width:  width,
		height: height,
		events: []KeyEvent{
			{Kind: KeyText, Text: "e2e5"},
			{Kind: KeyBackspace},
			{Kind: KeyText, Text: "4"},
			{Kind: KeySubmit},
			{Kind: KeyQuit},
		},
	}
	controller := newController(terminal, "")

	if err := controller.Run(); err != nil {
		t.Fatalf("controller run returned error: %v", err)
	}
	if controller.session.Game.Turn != engine.Black {
		t.Fatalf("expected black to move after corrected submit, got %s", controller.session.Game.Turn)
	}
	if len(controller.session.MoveLog) != 1 {
		t.Fatalf("expected one applied move after edit+submit, got %d", len(controller.session.MoveLog))
	}
	if controller.editor.String() != "" {
		t.Fatalf("expected editor to clear after corrected legal submit, got %q", controller.editor.String())
	}
	if controller.session.LastMoveNotation() != "e2e4" {
		t.Fatalf("expected corrected move e2e4, got %q", controller.session.LastMoveNotation())
	}
}

func TestControllerPromotionEscCancelsAndClearsInput(t *testing.T) {
	width, height := fullMinimumSize()
	terminal := &fakeTerminal{
		width:  width,
		height: height,
		events: []KeyEvent{
			{Kind: KeyText, Text: "e7e8"},
			{Kind: KeySubmit},
			{Kind: KeyCancel},
			{Kind: KeyQuit},
		},
	}
	controller := newController(terminal, "")
	controller.session.Game = &engine.GameState{Turn: engine.White}
	controller.session.Game.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	controller.session.Game.Board.Squares[7][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	controller.session.Game.Board.Squares[4][6] = &engine.Piece{Kind: engine.Pawn, Color: engine.White}
	controller.session.View = view.ViewStateFromGameState(controller.session.Game)
	controller.session.Preview("")

	if err := controller.Run(); err != nil {
		t.Fatalf("controller run returned error: %v", err)
	}
	if controller.session.InputMode != app.InputModeCommand {
		t.Fatalf("expected command mode after cancel, got %s", controller.session.InputMode)
	}
	if controller.editor.String() != "" {
		t.Fatalf("expected editor to clear after cancel, got %q", controller.editor.String())
	}
}

func TestControllerCtrlCQuitsImmediately(t *testing.T) {
	width, height := fullMinimumSize()
	terminal := &fakeTerminal{
		width:  width,
		height: height,
		events: []KeyEvent{
			{Kind: KeyQuit},
		},
	}
	controller := newController(terminal, "")

	if err := controller.Run(); err != nil {
		t.Fatalf("controller run returned error: %v", err)
	}
	if len(terminal.renders) != 1 {
		t.Fatalf("expected only the initial render before quit, got %d renders", len(terminal.renders))
	}
}

func TestRunReturnsTerminalRequirementError(t *testing.T) {
	err := run("", func() (Terminal, error) {
		return nil, ErrTerminalRequired
	})
	if !errors.Is(err, ErrTerminalRequired) {
		t.Fatalf("expected terminal requirement error, got %v", err)
	}
}
