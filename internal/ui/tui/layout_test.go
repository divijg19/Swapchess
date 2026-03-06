package tui

import "testing"

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
	layout, ok := computeLayout(minWidth, minHeight)
	if !ok {
		t.Fatalf("expected layout at minimum viewport")
	}
	if layout.UsableWidth <= 0 || layout.LeftWidth <= 0 || layout.RightWidth <= 0 {
		t.Fatalf("expected positive layout dimensions, got %+v", layout)
	}
	if layout.BoardBodyLines <= 0 || layout.LogBodyLines < minLogBodyLines {
		t.Fatalf("expected positive body line counts, got %+v", layout)
	}
}
