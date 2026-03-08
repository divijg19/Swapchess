package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/app"
)

func fullSizedModel(debug string) model {
	current := initialModel(debug)
	minWidth, minHeight := viewportForScale(2)
	current.session.Resize(minWidth, minHeight)
	current.syncInput()
	return current
}

func TestConstraintViewBelowThreshold(t *testing.T) {
	current := initialModel("")
	minWidth, minHeight := minimumViewport()
	for minWidth > 1 && minHeight > 1 {
		if _, ok := computeLayoutForContent(minWidth-1, minHeight-1, current.layoutContent()); !ok {
			break
		}
		minWidth--
		minHeight--
	}
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

func TestCompactViewAtCompactThreshold(t *testing.T) {
	current := initialModel("")
	minWidth, minHeight := minimumViewport()
	current.session.Resize(minWidth, minHeight)
	current.syncInput()

	view := current.View()
	if strings.Contains(view, "Viewport Constraints") {
		t.Fatalf("expected responsive compact layout at compact threshold, got constrained view")
	}
	if !strings.Contains(view, "Board") {
		t.Fatalf("expected board panel in compact layout")
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
	if !strings.Contains(strings.Join(updated.helpLines(80), "\n"), "Alt") {
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
	if strings.Contains(strings.Join(plain.helpLines(80), "\n"), "Debug") {
		t.Fatalf("unexpected debug help in plain TUI")
	}

	debug := fullSizedModel("engine")
	debug.helpExpanded = true
	if !strings.Contains(strings.Join(debug.helpLines(80), "\n"), "Debug") {
		t.Fatalf("expected debug help line when debug renderer is enabled")
	}
}

func TestGameLinesDropSwapSuppressedRow(t *testing.T) {
	current := fullSizedModel("")

	lines := strings.Join(current.gameLines(80), "\n")
	if strings.Contains(lines, "Swap suppressed") {
		t.Fatalf("did not expect swap-suppressed row in game state: %q", lines)
	}
	if !strings.Contains(lines, "Cursor") || !strings.Contains(lines, "Select") {
		t.Fatalf("expected aligned cursor and selected rows, got %q", lines)
	}
}

func TestMoveLogGroupsFullMoves(t *testing.T) {
	current := fullSizedModel("")
	current.session.Submit("e2e4")
	current.session.Submit("e7e5")

	lines := current.moveLogLines(80, 5)
	if len(lines) != 1 {
		t.Fatalf("expected one full-move line, got %d: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "01.") || !strings.Contains(lines[0], "e2e4") || !strings.Contains(lines[0], "e7e5") {
		t.Fatalf("expected grouped white and black moves, got %q", lines[0])
	}
}

func TestMoveLogScrollsThroughOlderEntries(t *testing.T) {
	current := fullSizedModel("")
	current.session.MoveLog = make([]app.MoveRecord, 0, 40)
	for idx := 1; idx <= 40; idx++ {
		current.session.MoveLog = append(current.session.MoveLog, app.MoveRecord{Index: idx, Notation: fmt.Sprintf("m%02d", idx)})
	}
	current.normalizeMoveLogScroll()
	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected layout for move log scrolling")
	}
	bottom := strings.Join(current.moveLogLines(layout.RightBodyWidth, layout.LogBodyLines), "\n")
	if !strings.Contains(bottom, "20.") {
		t.Fatalf("expected latest move group at bottom, got %q", bottom)
	}

	current.scrollMoveLog(1)
	current.normalizeMoveLogScroll()
	scrolled := strings.Join(current.moveLogLines(layout.RightBodyWidth, layout.LogBodyLines), "\n")
	if scrolled == bottom {
		t.Fatalf("expected scrolled move log to change visible entries")
	}
	if !strings.Contains(scrolled, "01.") && !strings.Contains(scrolled, "02.") {
		t.Fatalf("expected scrolled log to reveal older entries, got %q", scrolled)
	}
}

func TestSelectedBoardDoesNotClip(t *testing.T) {
	current := fullSizedModel("")
	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected full layout at minimum viewport")
	}

	current.session.ActivateCursor()
	panel := current.renderBoardPanel(layout)
	if strings.Contains(panel, "...") {
		t.Fatalf("did not expect clipped board panel after selection: %q", panel)
	}
}

func TestBoardPanelUsesBorder(t *testing.T) {
	current := fullSizedModel("")
	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected full layout")
	}

	panel := current.renderBoardPanel(layout)
	if !strings.Contains(panel, "╭") || !strings.Contains(panel, "╯") {
		t.Fatalf("expected board panel border, got %q", panel)
	}
}

func TestBoardPanelResizesInSyncWithLayout(t *testing.T) {
	current := fullSizedModel("")

	for _, widthDelta := range []int{0, 12, 28} {
		current.session.Resize(current.session.Width+widthDelta, current.session.Height+8)
		current.syncInput()

		layout, ok := current.layout()
		if !ok {
			t.Fatalf("expected layout after resize delta %d", widthDelta)
		}

		boardPanel := current.renderBoardPanel(layout)
		if got := lipgloss.Width(boardPanel); got != layout.LeftWidth {
			t.Fatalf("expected board panel width %d after resize delta %d, got %d", layout.LeftWidth, widthDelta, got)
		}
		if layout.RowWidth != layout.LeftWidth+layoutGapWidth+layout.RightWidth {
			t.Fatalf("expected synced row width after resize delta %d, got %+v", widthDelta, layout)
		}
	}
}

func TestBoardPanelFillsMainRowHeight(t *testing.T) {
	current := fullSizedModel("engine")
	current.helpExpanded = true
	current.session.Resize(current.session.Width+20, current.session.Height+8)
	current.syncInput()

	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected layout for expanded main row")
	}

	boardPanel := current.renderBoardPanel(layout)
	if got := lipgloss.Height(boardPanel); got != layout.MainHeight {
		t.Fatalf("expected board panel height %d to fill main row, got %d", layout.MainHeight, got)
	}
}

func TestSelectedPieceStillRendersAtMinimumViewport(t *testing.T) {
	current := fullSizedModel("")

	next, _ := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	view := updated.View()
	if strings.Contains(view, "Viewport Constraints") {
		t.Fatalf("expected full layout after selecting a piece, got %q", view)
	}
}

func TestFullViewUsesLeftAlignedHudPanels(t *testing.T) {
	current := fullSizedModel("")
	current.session.Resize(current.session.Width+18, current.session.Height)
	current.syncInput()

	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected layout at expanded width")
	}
	view := current.renderFull(layout)
	lines := strings.Split(view, "\n")
	maxWidth := 0
	for _, line := range lines {
		if lineWidth := lipgloss.Width(line); lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	if maxWidth != layout.UsableWidth {
		t.Fatalf("expected rendered full view to span usable width %d, got %d", layout.UsableWidth, maxWidth)
	}
	if layout.RightBodyWidth < minimumRightBodyWidth {
		t.Fatalf("expected HUD body width to stay above minimum %d, got %d", minimumRightBodyWidth, layout.RightBodyWidth)
	}
	if layout.RowWidth != layout.LeftWidth+layoutGapWidth+layout.RightWidth {
		t.Fatalf("expected board and fixed HUD to remain adjacent, got %+v", layout)
	}
	if layout.RightBodyWidth < preferredRightBodyWidth {
		t.Fatalf("expected HUD width to stay above preferred width floor, got %+v", layout)
	}
	if layoutGapWidth != 0 {
		t.Fatalf("expected board and HUD to be flush with no explicit gutter, got gap=%d", layoutGapWidth)
	}
	gamePanel := current.renderRightPreparedPanel("Game State", current.gameLines(layout.RightBodyWidth), layout.RightWidth, layout.GameBodyLines, false)
	panelLines := strings.Split(strings.TrimRight(gamePanel, "\n"), "\n")
	if len(panelLines) < 2 {
		t.Fatalf("expected rendered game panel lines")
	}
	trimmed := strings.TrimLeft(panelLines[1], " │╭╰├└─")
	if !strings.HasPrefix(trimmed, "Game State") {
		t.Fatalf("expected left-aligned game panel title, got %q", panelLines[1])
	}
}

func TestExpandedViewFillsViewportWidth(t *testing.T) {
	current := fullSizedModel("engine")
	current.session.Resize(current.session.Width+24, current.session.Height+6)
	current.syncInput()

	view := current.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	maxWidth := 0
	for _, line := range lines {
		if lineWidth := lipgloss.Width(line); lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	if maxWidth != current.session.Width {
		t.Fatalf("expected expanded view to fill viewport width %d, got %d", current.session.Width, maxWidth)
	}
}

func TestExpandedHudRemainsCompactAtLargeViewport(t *testing.T) {
	current := fullSizedModel("engine")
	current.helpExpanded = true
	current.session.Resize(current.session.Width+24, current.session.Height+8)
	current.syncInput()

	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected layout at large viewport")
	}
	if layout.RightBodyWidth < preferredRightBodyWidth {
		t.Fatalf("expected large viewport HUD width to grow or hold above preferred width, got %+v", layout)
	}
	if layout.UsableWidth >= largeViewportWidthThreshold {
		leftShare := (layout.LeftWidth * 100) / layout.UsableWidth
		if leftShare > targetBoardColumnPercent+1 {
			t.Fatalf("expected large viewport board column to stay near %d%%, got %d%% (%+v)", targetBoardColumnPercent, leftShare, layout)
		}
	}
}

func TestExpandedHelpStaysCompactInRightPanel(t *testing.T) {
	current := fullSizedModel("engine")
	current.helpExpanded = true
	current.syncInput()

	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected layout for expanded help")
	}
	view := current.renderFull(layout)
	if !strings.Contains(view, "PgUp/PgDn") {
		t.Fatalf("expected scroll help entry in expanded help, got %q", view)
	}
	if layout.HelpBodyLines > maximumHelpBodyLines {
		t.Fatalf("expected compact help height, got %+v", layout)
	}
}

func TestMinimumViewportRenderFitsViewport(t *testing.T) {
	current := initialModel("")
	minWidth, minHeight := minimumViewport()
	current.session.Resize(minWidth, minHeight)
	current.syncInput()

	view := current.View()
	assertViewFitsViewport(t, view, minWidth, minHeight)
	if strings.Contains(view, "Viewport Constraints") {
		t.Fatalf("expected full layout at minimum viewport, got constrained view")
	}
}

func TestMinimumViewportDebugRenderFitsViewport(t *testing.T) {
	current := initialModel("engine")
	minWidth, minHeight := minimumViewport()
	current.session.Resize(minWidth, minHeight)
	current.syncInput()

	view := current.View()
	assertViewFitsViewport(t, view, minWidth, minHeight)
	if strings.Contains(view, "Viewport Constraints") {
		t.Fatalf("expected full layout at minimum viewport with debug renderer, got constrained view")
	}
}

func assertViewFitsViewport(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) > height {
		t.Fatalf("expected rendered height <= %d, got %d", height, len(lines))
	}
	maxWidth := 0
	for _, line := range lines {
		if lineWidth := lipgloss.Width(line); lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	if maxWidth > width {
		t.Fatalf("expected rendered width <= %d, got %d", width, maxWidth)
	}
}
