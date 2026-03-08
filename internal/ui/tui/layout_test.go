package tui

import (
	"testing"

	rendertext "github.com/divijg19/Swapchess/internal/render/text"
)

func TestComputeLayoutRejectsBelowMinimum(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	if _, ok := computeLayout(minWidth-1, minHeight); ok {
		t.Fatalf("expected layout rejection below minimum width")
	}
	if _, ok := computeLayout(minWidth, minHeight-1); ok {
		t.Fatalf("expected layout rejection below minimum height")
	}
}

func TestComputeLayoutAcceptsMinimumViewport(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	if minWidth != 66 || minHeight != 31 {
		t.Fatalf("expected compact minimum viewport 66x31, got %dx%d", minWidth, minHeight)
	}
	layout, ok := computeLayout(minWidth, minHeight)
	if !ok {
		t.Fatalf("expected layout at minimum viewport")
	}
	if layout.UsableWidth <= 0 || layout.LeftWidth <= 0 || layout.RightWidth <= 0 {
		t.Fatalf("expected positive layout dimensions, got %+v", layout)
	}
	if layout.RowWidth != layout.LeftWidth+layoutGapWidth+layout.RightWidth {
		t.Fatalf("expected row width to be board + gap + HUD, got %+v", layout)
	}
	if layout.RightBodyWidth < minimumRightBodyWidth {
		t.Fatalf("expected HUD body width to stay above minimum %d, got %d", minimumRightBodyWidth, layout.RightBodyWidth)
	}
	if layout.BoardBodyLines <= 0 || layout.LogBodyLines <= 0 {
		t.Fatalf("expected positive body line counts, got %+v", layout)
	}
	if layout.BoardCellWidth < 1 || layout.BoardRowHeight < 1 {
		t.Fatalf("expected positive board cell geometry, got %+v", layout)
	}
	rightStackOuterHeight := (layout.GameBodyLines + panelChromeHeight) + (layout.HelpBodyLines + panelChromeHeight) + (layout.LogBodyLines + panelChromeHeight)
	if rightStackOuterHeight > minHeight-headerLines-(layout.InputBodyLines+panelChromeHeight) {
		t.Fatalf("expected HUD stack to fit the main row, got %d lines", rightStackOuterHeight)
	}
	if minHeight > rendertext.BoardLines+24 {
		t.Fatalf("expected compact full-layout height, got %d rows", minHeight)
	}
}

func TestComputeLayoutScalesRightPanelsAtWideViewport(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	narrow, ok := computeLayout(minWidth, minHeight)
	if !ok {
		t.Fatalf("expected layout at minimum viewport")
	}
	wide, ok := computeLayout(minWidth+48, minHeight+16)
	if !ok {
		t.Fatalf("expected layout at wide viewport")
	}
	if wide.RightBodyWidth < narrow.RightBodyWidth {
		t.Fatalf("expected HUD width to grow or hold at larger viewport, got %d -> %d", narrow.RightBodyWidth, wide.RightBodyWidth)
	}
	if wide.BoardOuterWidth < narrow.BoardOuterWidth {
		t.Fatalf("expected board HUD width to grow or hold at larger viewport, got %d -> %d", narrow.BoardOuterWidth, wide.BoardOuterWidth)
	}
	if wide.GameBodyLines > narrow.GameBodyLines || wide.HelpBodyLines > narrow.HelpBodyLines {
		t.Fatalf("expected wider HUD to need no extra wrapping lines, got narrow=%+v wide=%+v", narrow, wide)
	}
	if wide.LogBodyLines < narrow.LogBodyLines {
		t.Fatalf("expected move log height to stay level or grow, got %d -> %d", narrow.LogBodyLines, wide.LogBodyLines)
	}
	if wide.UsableWidth >= largeViewportWidthThreshold {
		boardShare := (wide.BoardBodyWidth * 100) / wide.UsableWidth
		if boardShare > targetBoardColumnPercent+1 {
			t.Fatalf("expected large viewport board width to stay near %d%%, got %d%% (%+v)", targetBoardColumnPercent, boardShare, wide)
		}
	}
	if wide.RowWidth != wide.UsableWidth || wide.MainHeight != rightColumnOuterHeight(wide.GameBodyLines, wide.LogBodyLines, wide.HelpBodyLines) {
		t.Fatalf("expected layout to absorb all available slack, got %+v", wide)
	}
}

func TestComputeLayoutWidensBoardAsViewportWidens(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	base, ok := computeLayout(minWidth, minHeight)
	if !ok {
		t.Fatalf("expected base layout to fit")
	}
	wideWidth, wideHeight := viewportForScale(2)
	wider, ok := computeLayout(wideWidth, wideHeight)
	if !ok {
		t.Fatalf("expected wider layout to fit")
	}
	if wider.BoardBodyWidth <= base.BoardBodyWidth {
		t.Fatalf("expected board width to grow with larger proportional viewport, got %d -> %d", base.BoardBodyWidth, wider.BoardBodyWidth)
	}
	if wider.LeftWidth < wider.BoardOuterWidth {
		t.Fatalf("expected board column width to contain board HUD width, got %+v", wider)
	}
	if wider.UsableWidth >= largeViewportWidthThreshold {
		panelShare := (wider.LeftWidth * 100) / wider.UsableWidth
		if panelShare > targetBoardColumnPercent+1 {
			t.Fatalf("expected wider layout board panel to stay near %d%%, got %d%% (%+v)", targetBoardColumnPercent, panelShare, wider)
		}
	}
}

func TestComputeLayoutKeepsHudWidthNearFixed(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	layout, ok := computeLayout(minWidth+40, minHeight+10)
	if !ok {
		t.Fatalf("expected layout at expanded viewport")
	}
	if layout.RightBodyWidth < preferredRightBodyWidth {
		t.Fatalf("expected HUD width to flex upward on expanded viewport, got %+v", layout)
	}
	if layout.LeftWidth < layout.BoardOuterWidth {
		t.Fatalf("expected expanded layout board column width to contain board HUD width, got %+v", layout)
	}
	if layout.UsableWidth >= largeViewportWidthThreshold {
		panelShare := (layout.LeftWidth * 100) / layout.UsableWidth
		if panelShare > targetBoardColumnPercent+1 {
			t.Fatalf("expected expanded layout board panel to stay near %d%%, got %d%% (%+v)", targetBoardColumnPercent, panelShare, layout)
		}
	}
	if layout.LogBodyLines < minimumLogBodyLines {
		t.Fatalf("expected move log height to have an increased baseline, got %+v", layout)
	}
	if layout.RowWidth != layout.UsableWidth || layout.MainHeight != rightColumnOuterHeight(layout.GameBodyLines, layout.LogBodyLines, layout.HelpBodyLines) {
		t.Fatalf("expected expanded layout to absorb remaining slack, got %+v", layout)
	}
}

func TestComputeLayoutWidensBoardAtFixedHeight(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	base, ok := computeLayout(minWidth, minHeight)
	if !ok {
		t.Fatalf("expected base layout")
	}
	wide, ok := computeLayout(minWidth+40, minHeight)
	if !ok {
		t.Fatalf("expected wide layout at fixed height")
	}
	if wide.BoardBodyLines != base.BoardBodyLines {
		t.Fatalf("expected fixed-height resize to preserve board height, got %d -> %d", base.BoardBodyLines, wide.BoardBodyLines)
	}
	if wide.BoardBodyWidth <= base.BoardBodyWidth {
		t.Fatalf("expected fixed-height resize to widen the board, got %d -> %d", base.BoardBodyWidth, wide.BoardBodyWidth)
	}
}

func TestComputeLayoutWidensBoardBeyondSeventyThreeColumnsAtCompactHeight(t *testing.T) {
	content := minimumLayoutContent()
	narrow, ok := computeLayoutForContent(73, minimumViewportHeight, content)
	if !ok {
		t.Fatalf("expected layout at 73 columns")
	}
	wide, ok := computeLayoutForContent(95, minimumViewportHeight, content)
	if !ok {
		t.Fatalf("expected layout at wider compact-height viewport")
	}
	if wide.BoardBodyWidth <= narrow.BoardBodyWidth {
		t.Fatalf("expected compact-height board to widen beyond 73 columns, got %d -> %d", narrow.BoardBodyWidth, wide.BoardBodyWidth)
	}
}

func TestComputeLayoutUsesMostOfBoardPanelAtCompactHeight(t *testing.T) {
	layout, ok := computeLayoutForContent(110, minimumViewportHeight, minimumLayoutContent())
	if !ok {
		t.Fatalf("expected layout for wide compact-height viewport")
	}
	innerWidth := layout.LeftWidth - boardPanelChromeWidth
	if innerWidth < 1 {
		t.Fatalf("expected positive board panel inner width, got %+v", layout)
	}
	expected := rendertext.FitMetrics(innerWidth, layout.MainHeight-boardPanelChromeHeight)
	if expected.Width == 0 || expected.Height == 0 {
		t.Fatalf("expected fitted board metrics for compact-height viewport")
	}
	if layout.BoardBodyWidth != expected.Width || layout.BoardBodyLines != expected.Height {
		t.Fatalf("expected compact-height board to use maximum fitted size %+v, got %+v", expected, layout)
	}
}

func TestComputeLayoutTallensBoardAsViewportTallens(t *testing.T) {
	minWidth, minHeight := minimumViewport()
	base, ok := computeLayout(minWidth, minHeight)
	if !ok {
		t.Fatalf("expected base layout to fit")
	}
	tallWidth, tallHeight := viewportForScale(2)
	taller, ok := computeLayout(tallWidth, tallHeight)
	if !ok {
		t.Fatalf("expected taller layout to fit")
	}
	if taller.BoardBodyLines <= base.BoardBodyLines {
		t.Fatalf("expected board height to grow with larger proportional viewport, got %d -> %d", base.BoardBodyLines, taller.BoardBodyLines)
	}
}

func TestComputeLayoutMaximizesBoardBeforeGrowingHud(t *testing.T) {
	content := layoutContent{
		GameLines:  []string{"state"},
		LogLines:   []string{"01. e2e4  e7e5"},
		HelpLines:  []string{"help"},
		InputLines: []string{"move> "},
	}
	width, height := 118, 39
	layout, ok := computeLayoutForContent(width, height, content)
	if !ok {
		t.Fatalf("expected layout for board-first sizing")
	}
	boardOuterWidthLimit := layout.UsableWidth - (minimumRightBodyWidth + panelChromeWidth) - layoutGapWidth
	if layout.UsableWidth >= largeViewportWidthThreshold {
		boardOuterWidthLimit = minInt(boardOuterWidthLimit, ((layout.UsableWidth*targetBoardColumnPercent)/100)+boardPanelChromeWidth)
	}
	expectedBoard := rendertext.FitMetrics(boardOuterWidthLimit-boardPanelChromeWidth, layout.MainHeight-boardPanelChromeHeight)
	if expectedBoard.Width == 0 || expectedBoard.Height == 0 {
		t.Fatalf("expected fitted board for board-first sizing")
	}
	if layout.LeftWidth != boardOuterWidthLimit {
		t.Fatalf("expected board-first layout to grow board panel to %d, got %+v", boardOuterWidthLimit, layout)
	}
	if layout.BoardOuterWidth != expectedBoard.Width+boardPanelChromeWidth {
		t.Fatalf("expected board-first layout to use widest square-ish board %d, got %+v", expectedBoard.Width+boardPanelChromeWidth, layout)
	}
	if layout.RightBodyWidth < minimumRightBodyWidth {
		t.Fatalf("expected panels to grow only after board saturation, got %+v", layout)
	}
}

func TestComputeLayoutLetsBoardReachSixtyPercentWhenHeightAllows(t *testing.T) {
	layout, ok := computeLayoutForContent(140, 55, layoutContent{
		GameLines:  []string{"state"},
		LogLines:   []string{"01. e2e4  e7e5"},
		HelpLines:  []string{"help"},
		InputLines: []string{"move> "},
	})
	if !ok {
		t.Fatalf("expected layout for tall wide viewport")
	}
	targetBoardWidth := (layout.UsableWidth * targetBoardColumnPercent) / 100
	if layout.BoardBodyWidth != targetBoardWidth {
		t.Fatalf("expected board to use full %d%% target width %d when height allows, got %+v", targetBoardColumnPercent, targetBoardWidth, layout)
	}
}

func TestComputeLayoutTracksDynamicContent(t *testing.T) {
	current := fullSizedModel("engine")
	current.helpExpanded = true
	current.session.Resize(current.session.Width+20, current.session.Height+10)
	current.session.Message = "This is a deliberately longer status line that should be measured into the command panel body height."
	current.session.Hint = "This is a deliberately longer hint line that should also participate in layout measurement."
	current.session.Submit("e2e4")
	current.session.Submit("e7e5")
	current.session.Submit("g1f3")
	current.session.Submit("b8c6")
	current.syncInput()

	layout, ok := current.layout()
	if !ok {
		t.Fatalf("expected layout for dynamic content")
	}
	if layout.InputBodyLines != minimumInputBodyLine {
		t.Fatalf("expected stable command panel minimum lines, got %d", layout.InputBodyLines)
	}
	if layout.RightBodyWidth < minimumRightBodyWidth {
		t.Fatalf("expected HUD body width above minimum with dynamic content, got %d", layout.RightBodyWidth)
	}
	if layout.GameBodyLines < len(current.gameLines(layout.RightBodyWidth)) {
		t.Fatalf("expected game panel to fit all measured lines, got layout=%+v", layout)
	}
	if layout.HelpBodyLines > maximumHelpBodyLines {
		t.Fatalf("expected help panel to stay compact, got layout=%+v", layout)
	}
	if layout.LogBodyLines < minimumLogBodyLines {
		t.Fatalf("expected move log panel to keep expanded baseline height, got layout=%+v", layout)
	}
}

func TestViewportForScaleFitsRequestedLegacyBoardShape(t *testing.T) {
	width, height := viewportForScale(2)
	layout, ok := computeLayout(width, height)
	if !ok {
		t.Fatalf("expected layout for legacy scale viewport")
	}
	requested := rendertext.MetricsForScale(2)
	if layout.BoardBodyWidth < requested.Width || layout.BoardBodyLines < requested.Height {
		t.Fatalf("expected fitted board to satisfy legacy scale-2 viewport, got %+v vs %+v", layout, requested)
	}
}
