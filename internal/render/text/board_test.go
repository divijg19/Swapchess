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

func TestRenderBoardLargeDimensions(t *testing.T) {
	catalog := pieces.NewCatalog("")
	board := RenderBoard(view.ViewStateFromGameState(engine.NewGame()), catalog, BoardOptions{Size: BoardSizeLarge})

	lines := strings.Split(strings.TrimRight(board, "\n"), "\n")
	if len(lines) != LargeBoardLines {
		t.Fatalf("expected %d lines, got %d", LargeBoardLines, len(lines))
	}

	maxWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			maxWidth = lipgloss.Width(line)
		}
	}
	if maxWidth != LargeBoardWidth {
		t.Fatalf("expected large board width %d, got %d", LargeBoardWidth, maxWidth)
	}
}

func TestRenderBoardXLargeDimensions(t *testing.T) {
	catalog := pieces.NewCatalog("")
	board := RenderBoard(view.ViewStateFromGameState(engine.NewGame()), catalog, BoardOptions{Size: BoardSizeXLarge})

	lines := strings.Split(strings.TrimRight(board, "\n"), "\n")
	if len(lines) != XLargeBoardLines {
		t.Fatalf("expected %d lines, got %d", XLargeBoardLines, len(lines))
	}

	maxWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			maxWidth = lipgloss.Width(line)
		}
	}
	if maxWidth != XLargeBoardWidth {
		t.Fatalf("expected xlarge board width %d, got %d", XLargeBoardWidth, maxWidth)
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

func TestRenderBoardCompactDimensions(t *testing.T) {
	catalog := pieces.NewCatalog("")
	board := RenderBoard(view.ViewStateFromGameState(engine.NewGame()), catalog, BoardOptions{Size: BoardSizeCompact})

	lines := strings.Split(strings.TrimRight(board, "\n"), "\n")
	if len(lines) != CompactBoardLines {
		t.Fatalf("expected %d lines, got %d", CompactBoardLines, len(lines))
	}

	maxWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			maxWidth = lipgloss.Width(line)
		}
	}
	if maxWidth != CompactBoardWidth {
		t.Fatalf("expected compact board width %d, got %d", CompactBoardWidth, maxWidth)
	}
}

func TestFitMetricsUsesAdditionalWidthAtFixedHeight(t *testing.T) {
	base := FitMetrics(BoardWidth, BoardLines)
	wider := FitMetrics(BoardWidth+13, BoardLines)
	if wider.Width <= base.Width {
		t.Fatalf("expected fitted board to widen with additional horizontal space, got %d -> %d", base.Width, wider.Width)
	}
	aspect := float64(wider.Width) / float64(wider.Height)
	if aspect > 3.50 {
		t.Fatalf("expected fixed-height fit to stay within bounded aspect, got %.2f (%+v)", aspect, wider)
	}
}

func TestFitMetricsUsesAdditionalWidthAtCompactHeight(t *testing.T) {
	base := FitMetrics(42, BoardLines)
	wider := FitMetrics(50, BoardLines)
	if wider.Width <= base.Width {
		t.Fatalf("expected compact-height fit to widen when extra width is available, got %d -> %d", base.Width, wider.Width)
	}
	aspect := float64(wider.Width) / float64(wider.Height)
	if aspect > 3.50 {
		t.Fatalf("expected compact-height fit to stay within bounded aspect, got %.2f (%+v)", aspect, wider)
	}
}

func TestFitMetricsCanFillWideCompactViewport(t *testing.T) {
	metrics := FitMetrics(66, BoardLines)
	if metrics.Width != 66 {
		t.Fatalf("expected compact-height fit to use full available width 66, got %+v", metrics)
	}
	aspect := float64(metrics.Width) / float64(metrics.Height)
	if aspect > 3.50 {
		t.Fatalf("expected wide compact fit to stay within bounded aspect, got %.2f (%+v)", aspect, metrics)
	}
}

func TestFitMetricsUsesAdditionalHeightContinuously(t *testing.T) {
	base := FitMetrics(BoardWidth+16, BoardLines)
	taller := FitMetrics(BoardWidth+16, BoardLines+8)
	if taller.RowHeight <= base.RowHeight {
		t.Fatalf("expected fitted board to grow taller with additional vertical space, got %d -> %d", base.RowHeight, taller.RowHeight)
	}
}

func TestFitMetricsPrefersSquareishAspect(t *testing.T) {
	metrics := FitMetrics(90, 45)
	if metrics.Width == 0 || metrics.Height == 0 {
		t.Fatalf("expected fitted metrics for 90x45")
	}
	aspect := float64(metrics.Width) / float64(metrics.Height)
	if aspect > 3.50 {
		t.Fatalf("expected square-ish fitted aspect <= 3.50, got %.2f (%+v)", aspect, metrics)
	}
}

func TestRenderBoardUsesDirectCellGeometry(t *testing.T) {
	catalog := pieces.NewCatalog("")
	metrics := MetricsForCellGeometry(2, 1)
	board := RenderBoard(view.ViewStateFromGameState(engine.NewGame()), catalog, BoardOptions{CellWidth: 2, RowHeight: 1})

	lines := strings.Split(strings.TrimRight(board, "\n"), "\n")
	if len(lines) != metrics.Height {
		t.Fatalf("expected %d lines, got %d", metrics.Height, len(lines))
	}
	maxWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			maxWidth = lipgloss.Width(line)
		}
	}
	if maxWidth != metrics.Width {
		t.Fatalf("expected fitted board width %d, got %d", metrics.Width, maxWidth)
	}
	if lipgloss.Width(lines[0]) != metrics.Width {
		t.Fatalf("expected top coordinate header width %d, got %d", metrics.Width, lipgloss.Width(lines[0]))
	}
	if lipgloss.Width(lines[len(lines)-1]) != metrics.Width {
		t.Fatalf("expected bottom coordinate header width %d, got %d", metrics.Width, lipgloss.Width(lines[len(lines)-1]))
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "8│") || strings.HasPrefix(line, "7│") || strings.HasPrefix(line, "6│") || strings.HasPrefix(line, "5│") || strings.HasPrefix(line, "4│") || strings.HasPrefix(line, "3│") || strings.HasPrefix(line, "2│") || strings.HasPrefix(line, "1│") ||
			strings.HasPrefix(line, "8 │") || strings.HasPrefix(line, "7 │") || strings.HasPrefix(line, "6 │") || strings.HasPrefix(line, "5 │") || strings.HasPrefix(line, "4 │") || strings.HasPrefix(line, "3 │") || strings.HasPrefix(line, "2 │") || strings.HasPrefix(line, "1 │") {
			t.Fatalf("expected rank labels to be separated from the board border, got %q", line)
		}
	}
}

func TestStyleBoardCellKeepsSingleColumnWidth(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		square   view.ViewSquare
		selected bool
		cursor   bool
		focused  bool
	}{
		{
			name:    "white piece",
			content: "♔",
			square: view.ViewSquare{
				Occupied: true,
				Kind:     engine.King,
				Color:    engine.White,
			},
		},
		{
			name:    "black piece selected",
			content: "♚",
			square: view.ViewSquare{
				Occupied: true,
				Kind:     engine.King,
				Color:    engine.Black,
			},
			selected: true,
		},
		{
			name:    "empty cursor",
			content: ".",
			square:  view.ViewSquare{},
			cursor:  true,
			focused: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			styled := StyleBoardCell(tc.content, 0, 0, tc.square, tc.selected, tc.cursor, tc.focused)
			if lipgloss.Width(styled) != 1 {
				t.Fatalf("expected styled cell width 1, got %d", lipgloss.Width(styled))
			}
		})
	}
}

func TestPiecePalettesDifferBySide(t *testing.T) {
	if whitePieceColor == blackPieceColor {
		t.Fatalf("expected white and black piece palettes to differ")
	}
}
