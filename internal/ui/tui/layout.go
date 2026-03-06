package tui

import rendertext "github.com/divijg19/Swapchess/internal/render/text"

const (
	appPaddingX       = 1
	layoutGapWidth    = 1
	gameBodyLines     = 7
	helpBodyLines     = 6
	inputBodyLines    = 4
	minLogBodyLines   = 5
	minRightBodyWidth = 32
)

type layoutSpec struct {
	UsableWidth    int
	LeftWidth      int
	RightWidth     int
	BoardBodyLines int
	LogBodyLines   int
}

func minimumViewport() (int, int) {
	boardWidth, boardLines := rendertext.Dimensions()
	leftOuterWidth := boardWidth + 4
	rightOuterWidth := minRightBodyWidth + 4
	width := leftOuterWidth + layoutGapWidth + rightOuterWidth + (appPaddingX * 2)

	boardOuterHeight := boardLines + 3
	rightOuterHeight := (gameBodyLines + 3) + (helpBodyLines + 3) + (minLogBodyLines + 3)
	mainHeight := maxInt(boardOuterHeight, rightOuterHeight)
	height := 1 + mainHeight + (inputBodyLines + 3)
	return width, height
}

func computeLayout(width, height int) (layoutSpec, bool) {
	minWidth, minHeight := minimumViewport()
	if width < minWidth || height < minHeight {
		return layoutSpec{}, false
	}

	boardWidth, boardLines := rendertext.Dimensions()
	usableWidth := width - (appPaddingX * 2)
	leftWidth := boardWidth + 4
	rightWidth := usableWidth - leftWidth - layoutGapWidth
	if rightWidth < minRightBodyWidth+4 {
		return layoutSpec{}, false
	}

	mainHeight := height - 1 - (inputBodyLines + 3)
	boardOuterHeight := boardLines + 3
	if mainHeight < boardOuterHeight {
		return layoutSpec{}, false
	}

	logBodyLines := mainHeight - (gameBodyLines + 3) - (helpBodyLines + 3) - 3
	if logBodyLines < minLogBodyLines {
		return layoutSpec{}, false
	}

	return layoutSpec{
		UsableWidth:    usableWidth,
		LeftWidth:      leftWidth,
		RightWidth:     rightWidth,
		BoardBodyLines: boardLines,
		LogBodyLines:   logBodyLines,
	}, true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
