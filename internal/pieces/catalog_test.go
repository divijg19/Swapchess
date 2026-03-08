package pieces

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/divijg19/Swapchess/engine"
)

func TestCatalogUsesDefaultGlyphsAndNilHandling(t *testing.T) {
	catalog := NewCatalog("")

	if got := catalog.Glyph(nil); got != "" {
		t.Fatalf("expected empty glyph for nil piece, got %q", got)
	}
	if got := catalog.GlyphFor(engine.King, engine.White); got != "♚" {
		t.Fatalf("expected default white king glyph, got %q", got)
	}
	if got := catalog.GlyphFor(engine.Pawn, engine.White); got != "♙" {
		t.Fatalf("expected default white pawn glyph, got %q", got)
	}
	if got := catalog.GlyphFor(engine.Queen, engine.Black); got != "♛" {
		t.Fatalf("expected default black queen glyph, got %q", got)
	}
}

func TestCatalogOverridesGlyphsFromAssets(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "white_king.txt"), []byte("WK\n"), 0o644); err != nil {
		t.Fatalf("write white king glyph: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "black_queen.txt"), []byte("BQ\n"), 0o644); err != nil {
		t.Fatalf("write black queen glyph: %v", err)
	}

	catalog := NewCatalog(dir)

	if got := catalog.GlyphFor(engine.King, engine.White); got != "WK" {
		t.Fatalf("expected overridden white king glyph, got %q", got)
	}
	if got := catalog.GlyphFor(engine.Queen, engine.Black); got != "BQ" {
		t.Fatalf("expected overridden black queen glyph, got %q", got)
	}
}

func TestCatalogIgnoresBlankOverrides(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "white_queen.txt"), []byte("   \n"), 0o644); err != nil {
		t.Fatalf("write blank override: %v", err)
	}

	catalog := NewCatalog(dir)

	if got := catalog.GlyphFor(engine.Queen, engine.White); got != "♛" {
		t.Fatalf("expected default glyph when override is blank, got %q", got)
	}
}

func TestCatalogFallsBackToASCIIWhenGlyphMissing(t *testing.T) {
	catalog := &Catalog{glyphs: map[string]string{}}

	if got := catalog.GlyphFor(engine.Rook, engine.White); got != "R" {
		t.Fatalf("expected white rook ASCII fallback, got %q", got)
	}
	if got := catalog.GlyphFor(engine.Knight, engine.Black); got != "n" {
		t.Fatalf("expected black knight ASCII fallback, got %q", got)
	}
	if got := catalog.GlyphFor(engine.PieceKind(99), engine.White); got != "?" {
		t.Fatalf("expected unknown piece fallback, got %q", got)
	}
}

func TestBundledPieceAssetsUseChessGlyphs(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime caller unavailable")
	}

	dir := filepath.Join(filepath.Dir(file), "..", "..", "assets", "pieces")
	catalog := NewCatalog(dir)

	if got := catalog.GlyphFor(engine.Pawn, engine.White); got != "♙" {
		t.Fatalf("expected bundled white pawn glyph, got %q", got)
	}
	if got := catalog.GlyphFor(engine.King, engine.White); got != "♚" {
		t.Fatalf("expected bundled white king glyph, got %q", got)
	}
	if got := catalog.GlyphFor(engine.King, engine.Black); got != "♚" {
		t.Fatalf("expected bundled black king glyph, got %q", got)
	}
}
