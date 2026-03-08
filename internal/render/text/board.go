package text

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/pieces"
	"github.com/divijg19/Swapchess/view"
)

const (
	XLargeBoardWidth  = 66
	XLargeBoardLines  = 35
	LargeBoardWidth   = 50
	LargeBoardLines   = 27
	BoardWidth        = 37
	BoardLines        = 19
	CompactBoardWidth = 19
	CompactBoardLines = 10
)

const (
	BaseBoardScale = 1
	MinBoardScale  = 1
)

type BoardSize int

const (
	BoardSizeFull BoardSize = iota
	BoardSizeLarge
	BoardSizeXLarge
	BoardSizeCompact
)

type CellDecorator func(content string, file, rank int, square view.ViewSquare, selected, cursor bool) string

type BoardOptions struct {
	Cursor    *engine.Position
	Selected  *engine.Position
	Decorator CellDecorator
	Size      BoardSize
	Scale     int
	CellWidth int
	RowHeight int
}

type BoardMetrics struct {
	Scale       int
	CellWidth   int
	RowHeight   int
	Width       int
	Height      int
	OuterWidth  int
	OuterHeight int
}

var (
	whitePieceColor = lipgloss.AdaptiveColor{Light: "250", Dark: "255"}
	blackPieceColor = lipgloss.AdaptiveColor{Light: "238", Dark: "245"}
	lightSquareTone = lipgloss.AdaptiveColor{Light: "248", Dark: "245"}
	darkSquareTone  = lipgloss.AdaptiveColor{Light: "241", Dark: "239"}
	coordinateTone  = lipgloss.AdaptiveColor{Light: "243", Dark: "242"}
	borderTone      = lipgloss.AdaptiveColor{Light: "240", Dark: "238"}
)

var (
	coordinateStyle = lipgloss.NewStyle().Foreground(coordinateTone)
	borderStyle     = lipgloss.NewStyle().Foreground(borderTone)
)

func Dimensions() (int, int) {
	return BoardWidth, BoardLines
}

func MetricsForScale(scale int) BoardMetrics {
	if scale < MinBoardScale {
		scale = BaseBoardScale
	}
	return MetricsForCellGeometry(boardCellWidth(scale), boardRowHeight(scale))
}

func MetricsForCellGeometry(cellWidth, rowHeight int) BoardMetrics {
	if cellWidth < 1 {
		cellWidth = 1
	}
	if rowHeight < 1 {
		rowHeight = 1
	}
	width := 8*cellWidth + 26
	height := 8*rowHeight + 11
	return BoardMetrics{
		Scale:       rowHeight,
		CellWidth:   cellWidth,
		RowHeight:   rowHeight,
		Width:       width,
		Height:      height,
		OuterWidth:  width,
		OuterHeight: height,
	}
}

func FitMetrics(maxWidth, maxHeight int) BoardMetrics {
	maxCellWidth := (maxWidth - 26) / 8
	maxRowHeight := (maxHeight - 11) / 8
	if maxCellWidth < 1 || maxRowHeight < 1 {
		return BoardMetrics{}
	}

	type candidate struct {
		metrics       BoardMetrics
		area          int
		slack         int
		aspectPenalty float64
	}

	const (
		targetAspectRatio = 1.85
		maxAspectRatio    = 2.25
		minAreaRetention  = 0.80
	)

	candidates := make([]candidate, 0, maxCellWidth*maxRowHeight)
	maxArea := -1

	for rowHeight := 1; rowHeight <= maxRowHeight; rowHeight++ {
		preferredCellWidth := boardCellWidth(rowHeight)
		minimumCellWidth := preferredCellWidth - 1
		if minimumCellWidth < 1 {
			minimumCellWidth = 1
		}
		maximumCellWidth := preferredCellWidth + 1
		for cellWidth := 1; cellWidth <= maxCellWidth; cellWidth++ {
			if cellWidth < minimumCellWidth || cellWidth > maximumCellWidth {
				continue
			}
			metrics := MetricsForCellGeometry(cellWidth, rowHeight)
			if metrics.Width > maxWidth || metrics.Height > maxHeight {
				continue
			}

			area := metrics.Width * metrics.Height
			slack := (maxWidth - metrics.Width) + (maxHeight - metrics.Height)
			aspectRatio := float64(metrics.Width) / float64(metrics.Height)
			if aspectRatio > maxAspectRatio {
				continue
			}
			aspectPenalty := absFloat(aspectRatio - targetAspectRatio)
			candidates = append(candidates, candidate{metrics: metrics, area: area, slack: slack, aspectPenalty: aspectPenalty})
			if area > maxArea {
				maxArea = area
			}
		}
	}

	if len(candidates) == 0 {
		return BoardMetrics{}
	}

	best := candidates[0]
	minArea := int(float64(maxArea) * minAreaRetention)
	if minArea < 1 {
		minArea = 1
	}

	for _, current := range candidates {
		if current.area < minArea {
			continue
		}
		if best.area < minArea ||
			current.area > best.area ||
			(current.area == best.area && current.aspectPenalty < best.aspectPenalty) ||
			(current.aspectPenalty == best.aspectPenalty && current.area == best.area && current.slack < best.slack) ||
			(current.aspectPenalty == best.aspectPenalty && current.area == best.area && current.slack == best.slack && current.metrics.RowHeight > best.metrics.RowHeight) ||
			(current.aspectPenalty == best.aspectPenalty && current.area == best.area && current.slack == best.slack && current.metrics.RowHeight == best.metrics.RowHeight && current.metrics.CellWidth > best.metrics.CellWidth) {
			best = current
		}
	}

	return best.metrics
}

func DimensionsForScale(scale int) (int, int) {
	metrics := MetricsForScale(scale)
	return metrics.Width, metrics.Height
}

func DimensionsFor(size BoardSize) (int, int) {
	if size == BoardSizeXLarge {
		return XLargeBoardWidth, XLargeBoardLines
	}
	if size == BoardSizeLarge {
		return LargeBoardWidth, LargeBoardLines
	}
	if size == BoardSizeCompact {
		return CompactBoardWidth, CompactBoardLines
	}
	return BoardWidth, BoardLines
}

func boardCellWidth(scale int) int {
	if scale < MinBoardScale {
		scale = BaseBoardScale
	}
	return (2 * scale) - 1
}

func boardRowHeight(scale int) int {
	if scale < MinBoardScale {
		return BaseBoardScale
	}
	return scale
}

func EmptySquareGlyph(file, rank int) string {
	if (file+rank)%2 == 0 {
		return "·"
	}
	return "•"
}

func StyleBoardCell(content string, file, rank int, square view.ViewSquare, selected, cursor, boardFocused bool) string {
	style := lipgloss.NewStyle().Width(1).Align(lipgloss.Center)

	if square.Occupied {
		if square.Color == engine.White {
			style = style.Bold(true).Foreground(whitePieceColor)
		} else {
			style = style.Foreground(blackPieceColor)
		}
	} else {
		if (file+rank)%2 == 0 {
			style = style.Foreground(lightSquareTone)
		} else {
			style = style.Foreground(darkSquareTone)
		}
	}

	switch {
	case selected && cursor:
		style = style.Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("166"))
	case selected:
		style = style.Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24"))
	case cursor && boardFocused:
		style = style.Bold(true).Foreground(lipgloss.Color("16")).Background(lipgloss.Color("220"))
	case cursor:
		style = style.Underline(true)
	}

	return style.Render(content)
}

func RenderBoardGrid(cells [8][8]string) string {
	return renderBoardGridFull(cells)
}

func renderBoardGridFull(cells [8][8]string) string {
	out := "    a   b   c   d   e   f   g   h\n"
	out += "  +---+---+---+---+---+---+---+---+\n"
	for rank := 7; rank >= 0; rank-- {
		out += fmt.Sprintf("%d |", rank+1)
		for file := 0; file < 8; file++ {
			out += fmt.Sprintf(" %s |", cells[rank][file])
		}
		out += fmt.Sprintf(" %d\n", rank+1)
		out += "  +---+---+---+---+---+---+---+---+\n"
	}
	out += "    a   b   c   d   e   f   g   h\n"
	return out
}

func renderBoardGridCompact(cells [8][8]string) string {
	out := "  a b c d e f g h\n"
	for rank := 7; rank >= 0; rank-- {
		out += fmt.Sprintf("%d ", rank+1)
		for file := 0; file < 8; file++ {
			out += cells[rank][file]
			if file < 7 {
				out += " "
			}
		}
		out += fmt.Sprintf(" %d\n", rank+1)
	}
	out += "  a b c d e f g h\n"
	return out
}

func renderBoardGridLarge(cells [8][8]string) string {
	return renderBoardGridScaled(cells, 3, 2)
}

func renderBoardGridXLarge(cells [8][8]string) string {
	return renderBoardGridScaled(cells, 5, 3)
}

func renderBoardGridScaled(cells [8][8]string, cellWidth, rowHeight int) string {
	var out strings.Builder
	out.WriteString(renderCoordinateHeader(cellWidth))
	out.WriteString(renderBorderLine(cellWidth, "┌", "┬", "┐"))

	for rank := 7; rank >= 0; rank-- {
		for subRow := 0; subRow < rowHeight; subRow++ {
			showContent := subRow == rowHeight-1
			if showContent {
				out.WriteString(renderRankPrefix(rank + 1))
			} else {
				out.WriteString("  ")
			}

			for file := 0; file < 8; file++ {
				cellContent := ""
				if showContent {
					cellContent = cells[rank][file]
				}
				out.WriteString(" ")
				out.WriteString(centerText(cellContent, cellWidth))
				out.WriteString(" ")
				out.WriteString(borderStyle.Render("│"))
			}

			out.WriteString("\n")
		}
		if rank > 0 {
			out.WriteString(renderBorderLine(cellWidth, "├", "┼", "┤"))
		} else {
			out.WriteString(renderBorderLine(cellWidth, "└", "┴", "┘"))
		}
	}
	out.WriteString(renderCoordinateHeader(cellWidth))
	return out.String()
}

func renderBoardGridForScale(cells [8][8]string, scale int) string {
	return renderBoardGridScaled(cells, boardCellWidth(scale), boardRowHeight(scale))
}

func renderCoordinateHeader(cellWidth int) string {
	var out strings.Builder
	out.WriteString("   ")
	for file := 0; file < 8; file++ {
		out.WriteString(" ")
		out.WriteString(coordinateStyle.Render(centerText(string(rune('a'+file)), cellWidth)))
		out.WriteString(" ")
		if file < 7 {
			out.WriteString(" ")
		}
	}
	out.WriteString("\n")
	return out.String()
}

func renderBorderLine(cellWidth int, left, join, right string) string {
	segment := strings.Repeat("─", cellWidth+2)
	var out strings.Builder
	out.WriteString(" ")
	out.WriteString(borderStyle.Render(left))
	for file := 0; file < 8; file++ {
		if file > 0 {
			out.WriteString(borderStyle.Render(join))
		}
		out.WriteString(borderStyle.Render(segment))
	}
	out.WriteString(borderStyle.Render(right))
	out.WriteString("\n")
	return out.String()
}

func renderRankPrefix(rank int) string {
	return coordinateStyle.Render(fmt.Sprintf("%d ", rank))
}

func centerText(content string, width int) string {
	contentWidth := lipgloss.Width(content)
	if contentWidth >= width {
		return content
	}
	left := (width - contentWidth) / 2
	right := width - contentWidth - left
	return strings.Repeat(" ", left) + content + strings.Repeat(" ", right)
}

func renderBoardGridBySize(cells [8][8]string, size BoardSize) string {
	if size == BoardSizeXLarge {
		return renderBoardGridXLarge(cells)
	}
	if size == BoardSizeLarge {
		return renderBoardGridLarge(cells)
	}
	if size == BoardSizeCompact {
		return renderBoardGridCompact(cells)
	}
	return renderBoardGridFull(cells)
}

func renderBoardGrid(cells [8][8]string, opts BoardOptions) string {
	if opts.CellWidth > 0 || opts.RowHeight > 0 {
		cellWidth := opts.CellWidth
		if cellWidth <= 0 {
			cellWidth = 1
		}
		rowHeight := opts.RowHeight
		if rowHeight <= 0 {
			rowHeight = 1
		}
		return renderBoardGridScaled(cells, cellWidth, rowHeight)
	}
	if opts.Scale > 0 {
		return renderBoardGridForScale(cells, opts.Scale)
	}
	return renderBoardGridBySize(cells, opts.Size)
}

func RenderBoard(v view.ViewState, catalog *pieces.Catalog, opts BoardOptions) string {
	var cells [8][8]string
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			square := v.Board[rank][file]
			content := EmptySquareGlyph(file, rank)
			if square.Occupied {
				content = catalog.GlyphFor(square.Kind, square.Color)
			}
			if opts.Decorator != nil {
				content = opts.Decorator(content, file, rank, square, positionEquals(opts.Selected, file, rank), positionEquals(opts.Cursor, file, rank))
			}
			cells[rank][file] = content
		}
	}
	return renderBoardGrid(cells, opts)
}

func RenderEngineBoard(state *engine.GameState, catalog *pieces.Catalog, opts BoardOptions) string {
	var cells [8][8]string
	for file := 0; file < 8; file++ {
		for rank := 0; rank < 8; rank++ {
			piece := state.Board.Squares[file][rank]
			square := view.ViewSquare{}
			content := EmptySquareGlyph(file, rank)
			if piece != nil {
				square = view.ViewSquare{
					Occupied: true,
					Kind:     piece.Kind,
					Color:    piece.Color,
				}
				content = catalog.Glyph(piece)
			}
			if opts.Decorator != nil {
				content = opts.Decorator(content, file, rank, square, positionEquals(opts.Selected, file, rank), positionEquals(opts.Cursor, file, rank))
			}
			cells[rank][file] = content
		}
	}
	return renderBoardGrid(cells, opts)
}

func StatusLabel(status view.GameStatus) string {
	return status.String()
}

func positionEquals(pos *engine.Position, file, rank int) bool {
	return pos != nil && pos.File == file && pos.Rank == rank
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
