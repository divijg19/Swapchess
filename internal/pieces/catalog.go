package pieces

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/divijg19/Swapchess/engine"
)

type Catalog struct {
	glyphs map[string]string
}

// NewCatalog loads an assets-first piece catalog from dir, with built-in Unicode defaults.
func NewCatalog(dir string) *Catalog {
	glyphs := defaultGlyphs()
	colors := []struct{ code, name string }{{"w", "white"}, {"b", "black"}}
	kinds := []string{"pawn", "knight", "bishop", "rook", "queen", "king"}

	for _, c := range colors {
		for _, k := range kinds {
			fname := filepath.Join(dir, fmt.Sprintf("%s_%s.txt", c.name, k))
			data, err := os.ReadFile(fname)
			if err != nil {
				continue
			}
			value := strings.TrimSpace(string(data))
			if value == "" {
				continue
			}
			glyphs[pieceKey(c.code, k)] = value
		}
	}

	return &Catalog{glyphs: glyphs}
}

func (c *Catalog) Glyph(piece *engine.Piece) string {
	if piece == nil {
		return ""
	}
	return c.GlyphFor(piece.Kind, piece.Color)
}

func (c *Catalog) GlyphFor(kind engine.PieceKind, color engine.Color) string {
	key := pieceKey(colorCode(color), kindName(kind))
	if glyph, ok := c.glyphs[key]; ok {
		return glyph
	}
	return asciiFallback(kind, color)
}

func defaultGlyphs() map[string]string {
	return map[string]string{
		"w_pawn":   "♙",
		"w_knight": "♘",
		"w_bishop": "♗",
		"w_rook":   "♖",
		"w_queen":  "♕",
		"w_king":   "♔",
		"b_pawn":   "♟",
		"b_knight": "♞",
		"b_bishop": "♝",
		"b_rook":   "♜",
		"b_queen":  "♛",
		"b_king":   "♚",
	}
}

func pieceKey(color, kind string) string {
	return color + "_" + kind
}

func colorCode(color engine.Color) string {
	if color == engine.White {
		return "w"
	}
	return "b"
}

func kindName(kind engine.PieceKind) string {
	switch kind {
	case engine.Pawn:
		return "pawn"
	case engine.Knight:
		return "knight"
	case engine.Bishop:
		return "bishop"
	case engine.Rook:
		return "rook"
	case engine.Queen:
		return "queen"
	case engine.King:
		return "king"
	default:
		return "unknown"
	}
}

func asciiFallback(kind engine.PieceKind, color engine.Color) string {
	ch := '?'
	switch kind {
	case engine.Pawn:
		ch = 'p'
	case engine.Knight:
		ch = 'n'
	case engine.Bishop:
		ch = 'b'
	case engine.Rook:
		ch = 'r'
	case engine.Queen:
		ch = 'q'
	case engine.King:
		ch = 'k'
	}
	if color == engine.White {
		return strings.ToUpper(string(ch))
	}
	return string(ch)
}
