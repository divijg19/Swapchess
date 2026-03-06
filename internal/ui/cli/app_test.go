package cli

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/app"
)

func sizedModel(debug string, width, height int) model {
	current := initialModel(debug)
	current.session.Resize(width, height)
	current.syncInput()
	return current
}

func TestCompactViewPreservesPrompt(t *testing.T) {
	model := sizedModel("", 30, 6)

	view := model.View()
	if !strings.Contains(view, "move>") {
		t.Fatalf("expected compact view to preserve prompt, got %q", view)
	}
}

func TestTypingUpdatesHint(t *testing.T) {
	current := initialModel("")
	typed := []rune("e2e4")
	for _, r := range typed {
		next, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		current = next.(model)
	}

	if current.session.Hint != "Move syntax and context look valid. Press Enter to apply." {
		t.Fatalf("unexpected hint after typing move: %q", current.session.Hint)
	}
}

func TestPromotionModeUpdatesPlaceholder(t *testing.T) {
	current := initialModel("")
	current.session.Game = &engine.GameState{Turn: engine.White}
	current.session.Game.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	current.session.Game.Board.Squares[7][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	current.session.Game.Board.Squares[4][6] = &engine.Piece{Kind: engine.Pawn, Color: engine.White}
	current.session.Preview("")

	current.input.SetValue("e7e8")
	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	current = next.(model)

	if current.session.InputMode != "promotion" {
		t.Fatalf("expected promotion mode, got %s", current.session.InputMode)
	}
	if current.input.Placeholder != "promotion: q/r/b/n" {
		t.Fatalf("expected promotion placeholder, got %q", current.input.Placeholder)
	}
}

func TestCtrlCQuits(t *testing.T) {
	current := initialModel("")
	_, cmd := current.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected ctrl+c to return quit command")
	}
}

func TestEscCancelsPromotionAndClearsInput(t *testing.T) {
	current := promotionModel()
	current.input.SetValue("x")

	next, cmd := current.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := next.(model)
	if cmd != nil {
		t.Fatalf("expected esc to cancel transient state without quitting")
	}
	if updated.session.InputMode != app.InputModeCommand {
		t.Fatalf("expected command mode after esc, got %s", updated.session.InputMode)
	}
	if updated.input.Value() != "" {
		t.Fatalf("expected esc to clear prompt contents, got %q", updated.input.Value())
	}
}

func TestEnterQuitReturnsQuitCommand(t *testing.T) {
	current := initialModel("")
	current.input.SetValue("quit")

	_, cmd := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected quit command when submitting quit")
	}
}

func TestCompactViewTriggersForSmallHeight(t *testing.T) {
	current := sizedModel("", 120, 5)

	view := current.View()
	if !strings.Contains(view, "Viewport too small for full board mode.") {
		t.Fatalf("expected small-height compact reason, got %q", view)
	}
	if !strings.Contains(view, "move>") {
		t.Fatalf("expected prompt line in compact view, got %q", view)
	}
}

func TestCompactViewTriggersForSmallWidth(t *testing.T) {
	current := sizedModel("", 20, 20)

	view := current.View()
	if !strings.Contains(view, "Viewport too small for full board mode.") {
		t.Fatalf("expected small-width compact reason, got %q", view)
	}
}

func promotionModel() model {
	current := initialModel("")
	current.session.Game = &engine.GameState{Turn: engine.White, RandSeed: 1}
	current.session.Game.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	current.session.Game.Board.Squares[7][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	current.session.Game.Board.Squares[4][6] = &engine.Piece{Kind: engine.Pawn, Color: engine.White}
	current.session.Preview("")
	current.input.SetValue("e7e8")
	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(model)
}
