package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/app"
)

func fullSizedModel(debug string) model {
	current := initialModel(debug)
	minWidth, minHeight := minimumViewport()
	current.session.Resize(minWidth, minHeight)
	current.syncInput()
	return current
}

func TestConstraintViewBelowThreshold(t *testing.T) {
	current := initialModel("")
	minWidth, minHeight := minimumViewport()
	current.session.Resize(minWidth-1, minHeight-1)
	current.syncInput()

	view := current.View()
	if !strings.Contains(view, "Viewport Constraints") {
		t.Fatalf("expected constrained layout, got %q", view)
	}
	if !strings.Contains(view, "Command Line") {
		t.Fatalf("expected constrained layout to keep input panel, got %q", view)
	}
}

func TestFullViewAtThreshold(t *testing.T) {
	current := fullSizedModel("")

	view := current.View()
	if strings.Contains(view, "Viewport Constraints") {
		t.Fatalf("expected full layout at threshold, got constrained view")
	}
	if !strings.Contains(view, "Board") {
		t.Fatalf("expected board panel in full layout")
	}
}

func TestResizePreservesPromptState(t *testing.T) {
	current := fullSizedModel("")
	minWidth, minHeight := minimumViewport()
	current.focus = focusPrompt
	current.syncInput()
	current.input.SetValue("e2e4")
	current.session.Preview(current.input.Value())

	current.session.Resize(minWidth-1, minHeight-1)
	current.syncInput()
	view := current.View()
	if !strings.Contains(view, "e2e4") {
		t.Fatalf("expected constrained view to preserve prompt contents, got %q", view)
	}
}

func TestEscCancelsTransientState(t *testing.T) {
	current := initialModel("")
	selected := current.session.Cursor
	current.session.Selected = &selected
	current.session.InputMode = app.InputModeBoardSelect
	current.syncInput()

	next, cmd := current.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := next.(model)
	if cmd != nil {
		t.Fatalf("expected esc to cancel transient state without quitting")
	}
	if updated.session.InputMode != app.InputModeCommand {
		t.Fatalf("expected command input mode after esc, got %s", updated.session.InputMode)
	}
}

func TestCtrlCQuits(t *testing.T) {
	current := initialModel("")
	_, cmd := current.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected ctrl+c to return a quit command")
	}
}

func TestTabTogglesFocus(t *testing.T) {
	current := fullSizedModel("")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := next.(model)
	if updated.focus != focusPrompt {
		t.Fatalf("expected tab to focus prompt, got %s", updated.focus)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	if updated.focus != focusBoard {
		t.Fatalf("expected second tab to return focus to board, got %s", updated.focus)
	}
}

func TestColonFocusesPrompt(t *testing.T) {
	current := fullSizedModel("")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	updated := next.(model)
	if updated.focus != focusPrompt {
		t.Fatalf("expected colon to focus prompt, got %s", updated.focus)
	}
}

func TestQuestionMarkExpandsHelp(t *testing.T) {
	current := fullSizedModel("")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	updated := next.(model)
	if !updated.helpExpanded {
		t.Fatalf("expected question mark to expand help")
	}
	if !strings.Contains(strings.Join(updated.helpLines(), "\n"), "Arrows or h/j/k/l: move board cursor") {
		t.Fatalf("expected expanded help line in help model")
	}
}

func TestBoardNavigationMovesCursor(t *testing.T) {
	current := fullSizedModel("")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updated := next.(model)
	if updated.session.Cursor != (engine.Position{File: 3, Rank: 1}) {
		t.Fatalf("expected cursor to move left to d2, got %+v", updated.session.Cursor)
	}
}

func TestBoardSelectionAndMoveApply(t *testing.T) {
	current := fullSizedModel("")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if updated.session.InputMode != app.InputModeBoardSelect {
		t.Fatalf("expected board selection mode after first enter, got %s", updated.session.InputMode)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(model)

	if updated.session.InputMode != app.InputModeCommand {
		t.Fatalf("expected command mode after move, got %s", updated.session.InputMode)
	}
	if len(updated.session.MoveLog) != 1 {
		t.Fatalf("expected one move after board interaction, got %d", len(updated.session.MoveLog))
	}
}

func TestEscLeavesPromptFocus(t *testing.T) {
	current := fullSizedModel("")
	current.focus = focusPrompt
	current.syncInput()

	next, cmd := current.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := next.(model)
	if cmd != nil {
		t.Fatalf("expected esc to leave prompt focus without quitting")
	}
	if updated.focus != focusBoard {
		t.Fatalf("expected esc to return focus to board, got %s", updated.focus)
	}
}

func TestBoardUndoShortcut(t *testing.T) {
	current := fullSizedModel("")
	current.session.Submit("e2e4")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	updated := next.(model)
	if len(updated.session.MoveLog) != 0 {
		t.Fatalf("expected board undo shortcut to remove last move, got %d entries", len(updated.session.MoveLog))
	}
}

func TestDebugHelpShowsRendererLineOnlyWhenEnabled(t *testing.T) {
	plain := fullSizedModel("")
	plain.helpExpanded = true
	if strings.Contains(strings.Join(plain.helpLines(), "\n"), "Debug: renderer view|engine|toggle") {
		t.Fatalf("unexpected debug help in plain TUI")
	}

	debug := fullSizedModel("engine")
	debug.helpExpanded = true
	if !strings.Contains(strings.Join(debug.helpLines(), "\n"), "Debug: renderer view|engine|toggle") {
		t.Fatalf("expected debug help line when debug renderer is enabled")
	}
}
