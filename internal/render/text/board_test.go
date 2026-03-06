package text

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/pieces"
	"github.com/divijg19/Swapchess/view"
)

func TestRenderBoardDimensions(t *testing.T) {
	catalog := pieces.NewCatalog("")
	board := RenderBoard(view.ViewStateFromGameState(engine.NewGame()), catalog, BoardOptions{})

	lines := strings.Split(strings.TrimRight(board, "\n"), "\n")
	if len(lines) != BoardLines {
		t.Fatalf("expected %d lines, got %d", BoardLines, len(lines))
	}

	maxWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			maxWidth = lipgloss.Width(line)
		}
	}
	if maxWidth != BoardWidth {
		t.Fatalf("expected board width %d, got %d", BoardWidth, maxWidth)
	}
}

func TestRenderEngineBoardDimensions(t *testing.T) {
	catalog := pieces.NewCatalog("")
	board := RenderEngineBoard(engine.NewGame(), catalog, BoardOptions{})

	lines := strings.Split(strings.TrimRight(board, "\n"), "\n")
	if len(lines) != BoardLines {
		t.Fatalf("expected %d lines, got %d", BoardLines, len(lines))
	}
}
