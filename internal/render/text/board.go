package text

import (
	"fmt"

	"github.com/divijg19/Swapchess/engine"
	"github.com/divijg19/Swapchess/internal/pieces"
	"github.com/divijg19/Swapchess/view"
)

const (
	BoardWidth = 37
	BoardLines = 19
)

type CellDecorator func(content string, file, rank int, square view.ViewSquare, selected, cursor bool) string

type BoardOptions struct {
	Cursor    *engine.Position
	Selected  *engine.Position
	Decorator CellDecorator
}

func Dimensions() (int, int) {
	return BoardWidth, BoardLines
}

func EmptySquareGlyph(file, rank int) string {
	if (file+rank)%2 == 0 {
		return "."
	}
	return "o"
}

func RenderBoardGrid(cells [8][8]string) string {
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
	return RenderBoardGrid(cells)
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
	return RenderBoardGrid(cells)
}

func StatusLabel(status view.GameStatus) string {
	return status.String()
}

func positionEquals(pos *engine.Position, file, rank int) bool {
	return pos != nil && pos.File == file && pos.Rank == rank
}
