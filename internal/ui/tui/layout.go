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
	minimumInputBodyLine    = 3
	minimumRightBodyWidth   = 22
	preferredRightBodyWidth = 26
	minimumGameBodyLines    = 4
	minimumHelpBodyLines    = 3
	maximumHelpBodyLines    = 5
	minimumLogBodyLines     = 8
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
	return viewportForScale(rendertext.BaseBoardScale)
}

func computeLayout(width, height int) (layoutSpec, bool) {
	return computeLayoutForContent(width, height, defaultLayoutContent())
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

	rightBodyWidth := targetRightBodyWidth(usableWidth, minimumBoardOuterWidth)
	if rightBodyWidth == 0 {
		return layoutSpec{}, false
	}
	rightWidth := rightBodyWidth + panelChromeWidth

	availableBoardOuterWidth := usableWidth - rightWidth - layoutGapWidth
	if availableBoardOuterWidth < minimumBoardOuterWidth {
		return layoutSpec{}, false
	}
	availableBoardWidth := reducedBoardWidth(maxInt(availableBoardOuterWidth-boardPanelChromeWidth, minimumBoardWidth), minimumBoardWidth)

	inputBodyLines := minimumInputBodyLine
	mainHeight := height - headerLines - (inputBodyLines + panelChromeHeight)
	if mainHeight <= boardPanelChromeHeight {
		return layoutSpec{}, false
	}

	gameBodyLines, helpBodyLines := fixedInfoPanelBodyLines(content, rightBodyWidth)
	minimumRightHeight := rightColumnOuterHeight(gameBodyLines, minimumLogBodyLines, helpBodyLines)
	if minimumRightHeight > mainHeight {
		return layoutSpec{}, false
	}

	boardMetrics := rendertext.FitMetrics(availableBoardWidth, mainHeight-boardPanelChromeHeight)
	if boardMetrics.Width == 0 || boardMetrics.Height == 0 {
		return layoutSpec{}, false
	}

	boardOuterWidth := boardMetrics.Width + boardPanelChromeWidth
	boardOuterLines := boardMetrics.Height + boardPanelChromeHeight
	if boardOuterLines > mainHeight {
		return layoutSpec{}, false
	}

	logBodyLines := fixedMoveLogBodyLines(mainHeight, gameBodyLines, helpBodyLines)
	rightColumnHeight := rightColumnOuterHeight(gameBodyLines, logBodyLines, helpBodyLines)
	if rightColumnHeight > mainHeight {
		return layoutSpec{}, false
	}

	leftWidth := usableWidth - layoutGapWidth - rightWidth
	rowWidth := leftWidth + layoutGapWidth + rightWidth
	if leftWidth < boardOuterWidth || rowWidth > usableWidth {
		return layoutSpec{}, false
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

func targetRightBodyWidth(usableWidth, minimumBoardOuterWidth int) int {
	maxRightBodyWidth := usableWidth - minimumBoardOuterWidth - layoutGapWidth - panelChromeWidth
	if maxRightBodyWidth < minimumRightBodyWidth {
		return 0
	}
	rightBodyWidth := minInt(maxRightBodyWidth, preferredRightBodyWidth)
	if usableWidth >= largeViewportWidthThreshold {
		targetLeftWidth := (usableWidth * targetBoardColumnPercent) / 100
		if targetLeftWidth < minimumBoardOuterWidth {
			targetLeftWidth = minimumBoardOuterWidth
		}
		targetRightBodyWidth := usableWidth - targetLeftWidth - layoutGapWidth - panelChromeWidth
		if targetRightBodyWidth > rightBodyWidth {
			rightBodyWidth = minInt(targetRightBodyWidth, maxRightBodyWidth)
		}
	}
	return rightBodyWidth
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

func viewportForScale(scale int) (int, int) {
	boardMetrics := rendertext.MetricsForScale(scale)
	content := defaultLayoutContent()
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
			alignedDualRow("Move", "arrows", "Log", "PgUp/PgDn", 5),
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
