package tui

import rendertext "github.com/divijg19/Swapchess/internal/render/text"

const (
	appPaddingX             = 1
	layoutGapWidth          = 0
	headerLines             = 1
	boardPanelChromeWidth   = 2
	boardPanelChromeHeight  = 3
	boardWidthPercent       = 100
	targetBoardColumnPercent = 60
	largeViewportWidthThreshold = 96
	minimumViewportWidth    = 66
	minimumViewportHeight   = 31
	minimumInputBodyLine    = 3
	minimumRightBodyWidth   = 22
	preferredRightBodyWidth = 26
	minimumGameBodyLines    = 4
	minimumHelpBodyLines    = 3
	maximumHelpBodyLines    = 5
	minimumLogBodyLines     = 6
	panelChromeWidth        = 4
	panelBorderWidth        = 2
	panelChromeHeight       = 3
)

type layoutContent struct {
	GameLines  []string
	LogLines   []string
	HelpLines  []string
	InputLines []string
}

type layoutSpec struct {
	ViewportWidth   int
	ViewportHeight  int
	UsableWidth     int
	RowWidth        int
	MainHeight      int
	LeftWidth       int
	RightWidth      int
	RightBodyWidth  int
	BoardBodyWidth  int
	BoardBodyLines  int
	BoardOuterWidth int
	BoardOuterLines int
	BoardCellWidth  int
	BoardRowHeight  int
	LogBodyLines    int
	GameBodyLines   int
	HelpBodyLines   int
	InputBodyLines  int
}

func minimumViewport() (int, int) {
	return minimumViewportWidth, minimumViewportHeight
}

func computeLayout(width, height int) (layoutSpec, bool) {
	if width < minimumViewportWidth || height < minimumViewportHeight {
		return layoutSpec{}, false
	}
	return computeLayoutForContent(width, height, minimumLayoutContent())
}

func computeLayoutForContent(width, height int, content layoutContent) (layoutSpec, bool) {
	if width <= 0 || height <= 0 {
		return layoutSpec{}, false
	}
	content = normalizeLayoutContent(content)

	minimumBoardWidth := rendertext.MetricsForScale(rendertext.BaseBoardScale).Width
	minimumBoardOuterWidth := minimumBoardWidth + boardPanelChromeWidth
	usableWidth := width - (appPaddingX * 2)
	if usableWidth <= 0 {
		return layoutSpec{}, false
	}
	maxRightBodyWidth := usableWidth - minimumBoardOuterWidth - layoutGapWidth - panelChromeWidth
	if maxRightBodyWidth < minimumRightBodyWidth {
		return layoutSpec{}, false
	}

	inputBodyLines := minimumInputBodyLine
	mainHeight := height - headerLines - (inputBodyLines + panelChromeHeight)
	if mainHeight <= boardPanelChromeHeight {
		return layoutSpec{}, false
	}

	maximumBoardOuterWidth := usableWidth - (minimumRightBodyWidth + panelChromeWidth) - layoutGapWidth
	if maximumBoardOuterWidth < minimumBoardOuterWidth {
		return layoutSpec{}, false
	}
	boardOuterWidthLimit := maximumBoardOuterWidth
	if usableWidth >= largeViewportWidthThreshold {
		targetBoardWidth := (usableWidth * targetBoardColumnPercent) / 100
		targetBoardOuterWidth := targetBoardWidth + boardPanelChromeWidth
		if targetBoardOuterWidth >= minimumBoardOuterWidth {
			boardOuterWidthLimit = minInt(boardOuterWidthLimit, targetBoardOuterWidth)
		}
	}

	return maximizeBoardLayout(width, height, usableWidth, mainHeight, inputBodyLines, boardOuterWidthLimit, minimumBoardWidth, minimumBoardOuterWidth, content)
}

func normalizeLayoutContent(content layoutContent) layoutContent {
	if len(content.GameLines) == 0 {
		content.GameLines = []string{"No game state."}
	}
	if len(content.LogLines) == 0 {
		content.LogLines = []string{"No moves yet."}
	}
	if len(content.HelpLines) == 0 {
		content.HelpLines = []string{"No help entries."}
	}
	if len(content.InputLines) == 0 {
		content.InputLines = []string{"move> "}
	}
	return content
}

func wrappedLineCount(lines []string, width int) int {
	wrapped := wrapPlainLines(lines, width)
	if len(wrapped) == 0 {
		return 1
	}
	return len(wrapped)
}

func fixedInfoPanelBodyLines(content layoutContent, rightBodyWidth int) (int, int) {
	gameBodyLines := maxInt(wrappedLineCount(content.GameLines, rightBodyWidth), minimumGameBodyLines)
	helpBodyLines := maxInt(wrappedLineCount(content.HelpLines, rightBodyWidth), minimumHelpBodyLines)
	if helpBodyLines > maximumHelpBodyLines {
		helpBodyLines = maximumHelpBodyLines
	}
	return gameBodyLines, helpBodyLines
}

func fixedMoveLogBodyLines(mainHeight, gameBodyLines, helpBodyLines int) int {
	remaining := mainHeight - ((gameBodyLines + panelChromeHeight) + (helpBodyLines + panelChromeHeight) + panelChromeHeight)
	if remaining < minimumLogBodyLines {
		return minimumLogBodyLines
	}
	return remaining
}

func rightColumnOuterHeight(gameBodyLines, logBodyLines, helpBodyLines int) int {
	return (gameBodyLines + panelChromeHeight) + (logBodyLines + panelChromeHeight) + (helpBodyLines + panelChromeHeight)
}

func maximizeBoardLayout(width, height, usableWidth, mainHeight, inputBodyLines, boardOuterWidthLimit, minimumBoardWidth, minimumBoardOuterWidth int, content layoutContent) (layoutSpec, bool) {
	for currentBoardOuterLimit := boardOuterWidthLimit; currentBoardOuterLimit >= minimumBoardOuterWidth; currentBoardOuterLimit-- {
		availableBoardWidth := reducedBoardWidth(maxInt(currentBoardOuterLimit-boardPanelChromeWidth, minimumBoardWidth), minimumBoardWidth)
		boardMetrics := rendertext.FitMetrics(availableBoardWidth, mainHeight-boardPanelChromeHeight)
		if boardMetrics.Width == 0 || boardMetrics.Height == 0 {
			continue
		}

		boardOuterWidth := boardMetrics.Width + boardPanelChromeWidth
		boardOuterLines := boardMetrics.Height + boardPanelChromeHeight
		if boardOuterWidth > currentBoardOuterLimit || boardOuterLines > mainHeight {
			continue
		}

		leftWidth := currentBoardOuterLimit
		rightWidth := usableWidth - layoutGapWidth - leftWidth
		rightBodyWidth := rightWidth - panelChromeWidth
		if rightBodyWidth < minimumRightBodyWidth {
			continue
		}

		gameBodyLines, helpBodyLines := fixedInfoPanelBodyLines(content, rightBodyWidth)
		if helpBodyLines > maximumHelpBodyLines {
			continue
		}
		minimumRightHeight := rightColumnOuterHeight(gameBodyLines, minimumLogBodyLines, helpBodyLines)
		if minimumRightHeight > mainHeight {
			continue
		}

		logBodyLines := fixedMoveLogBodyLines(mainHeight, gameBodyLines, helpBodyLines)
		rightColumnHeight := rightColumnOuterHeight(gameBodyLines, logBodyLines, helpBodyLines)
		if rightColumnHeight > mainHeight {
			continue
		}

		rowWidth := leftWidth + layoutGapWidth + rightWidth
		if rowWidth > usableWidth {
			continue
		}

		return layoutSpec{
			ViewportWidth:   width,
			ViewportHeight:  height,
			UsableWidth:     usableWidth,
			RowWidth:        rowWidth,
			MainHeight:      mainHeight,
			LeftWidth:       leftWidth,
			RightWidth:      rightWidth,
			RightBodyWidth:  rightBodyWidth,
			BoardBodyWidth:  boardMetrics.Width,
			BoardBodyLines:  boardMetrics.Height,
			BoardOuterWidth: boardOuterWidth,
			BoardOuterLines: boardOuterLines,
			BoardCellWidth:  boardMetrics.CellWidth,
			BoardRowHeight:  boardMetrics.RowHeight,
			LogBodyLines:    logBodyLines,
			GameBodyLines:   gameBodyLines,
			HelpBodyLines:   helpBodyLines,
			InputBodyLines:  inputBodyLines,
		}, true
	}

	return layoutSpec{}, false
}

func viewportForScale(scale int) (int, int) {
	return viewportForContentScale(scale, defaultLayoutContent())
}

func viewportForContentScale(scale int, content layoutContent) (int, int) {
	boardMetrics := rendertext.MetricsForScale(scale)
	rightWidth := preferredRightBodyWidth + panelChromeWidth
	gameBodyLines, helpBodyLines := fixedInfoPanelBodyLines(content, preferredRightBodyWidth)
	rightColumnHeight := rightColumnOuterHeight(gameBodyLines, minimumLogBodyLines, helpBodyLines)
	baseWidth := boardMetrics.Width + boardPanelChromeWidth + layoutGapWidth + rightWidth + (appPaddingX * 2)
	baseHeight := headerLines + maxInt(boardMetrics.Height+boardPanelChromeHeight, rightColumnHeight) + (minimumInputBodyLine + panelChromeHeight)
	fallbackWidth, fallbackHeight := baseWidth, baseHeight
	fallbackFound := false

	for currentHeight := baseHeight; currentHeight <= baseHeight+80; currentHeight++ {
		for currentWidth := baseWidth; currentWidth <= baseWidth+160; currentWidth++ {
			layout, ok := computeLayoutForContent(currentWidth, currentHeight, content)
			if ok && layout.BoardBodyWidth >= boardMetrics.Width && layout.BoardBodyLines >= boardMetrics.Height {
				return tightenViewportForScale(currentWidth, currentHeight, content, boardMetrics)
			}
			if ok && (layout.BoardRowHeight > boardMetrics.RowHeight || layout.BoardCellWidth > boardMetrics.CellWidth) && !fallbackFound {
				fallbackWidth, fallbackHeight = currentWidth, currentHeight
				fallbackFound = true
			}
		}
	}

	if fallbackFound {
		return tightenViewportForScale(fallbackWidth, fallbackHeight, content, boardMetrics)
	}
	return baseWidth, baseHeight
}

func tightenViewportForScale(width, height int, content layoutContent, minimumBoard rendertext.BoardMetrics) (int, int) {
	for width > 1 {
		layout, ok := computeLayoutForContent(width-1, height, content)
		if !ok || layout.BoardBodyWidth < minimumBoard.Width || layout.BoardBodyLines < minimumBoard.Height {
			break
		}
		width--
	}
	for height > 1 {
		layout, ok := computeLayoutForContent(width, height-1, content)
		if !ok || layout.BoardBodyWidth < minimumBoard.Width || layout.BoardBodyLines < minimumBoard.Height {
			break
		}
		height--
	}
	return width, height
}

func defaultLayoutContent() layoutContent {
	return layoutContent{
		GameLines: []string{
			alignedDualRow("Turn", "white", "Status", "in play", 6),
			alignedDualRow("Cursor", "e2", "Select", "-", 6),
			alignedDualRow("Last", "-", "Castle", "W:KQ B:KQ", 6),
			alignedRow("Render", "engine", 6),
		},
		LogLines: []string{
			"01. e2e4  e7e5",
			"02. g1f3  b8c6",
		},
		HelpLines: []string{
			alignedDualRow("Tab", "focus", ":", "prompt", 5),
			alignedDualRow("Enter", "select", "Esc", "back", 5),
			alignedDualRow("u", "undo", "?", "help", 5),
			alignedDualRow("Move", "keys", "Log", "PgUp/PgDn", 5),
			alignedDualRow("Type", "prompt", "Debug", "render", 5),
		},
		InputLines: []string{
			alignedDualRow("Mode", "move/command", "Focus", "board", 7),
			"Enter a move like e2e4. Type help for commands.",
			"Hint: Enter move (e2e4 / e7e8q) or command (help / undo / clear / quit).",
			"move> ",
		},
	}
}

func minimumLayoutContent() layoutContent {
	return layoutContent{
		GameLines: []string{
			alignedRow("Turn", "white", 6),
			alignedRow("Status", "in play", 6),
			alignedRow("Cursor", "e2", 6),
			alignedRow("Select", "-", 6),
			alignedRow("Last", "-", 6),
			alignedRow("Castle", "W:KQ B:KQ", 6),
		},
		LogLines: []string{
			"01. e2e4  e7e5",
			"02. g1f3  b8c6",
		},
		HelpLines: []string{
			alignedDualRow("Tab", "focus", ":", "prompt", 5),
			alignedDualRow("Enter", "select", "Esc", "back", 5),
			alignedDualRow("u", "undo", "?", "help", 5),
		},
		InputLines: []string{
			alignedDualRow("Mode", "move/command", "Focus", "board", 7),
			"Enter a move like e2e4. Type help for commands.",
			"Hint: Enter move (e2e4 / e7e8q) or command (help / undo / clear / quit).",
			"move> ",
		},
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func reducedBoardWidth(width, minimum int) int {
	reduced := (width * boardWidthPercent) / 100
	if reduced < minimum {
		return minimum
	}
	return reduced
}
