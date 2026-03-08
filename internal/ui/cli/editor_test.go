package cli

import "testing"

func TestEditorInsertAndCursorMoves(t *testing.T) {
	var editor Editor

	editor.Insert("e24")
	if !editor.MoveLeft() {
		t.Fatalf("expected move left to succeed")
	}
	editor.Insert("e")

	if editor.String() != "e2e4" {
		t.Fatalf("expected inserted move to be e2e4, got %q", editor.String())
	}
	if editor.Cursor() != 3 {
		t.Fatalf("expected cursor at 3, got %d", editor.Cursor())
	}

	if !editor.MoveHome() || editor.Cursor() != 0 {
		t.Fatalf("expected home to move cursor to start, got %d", editor.Cursor())
	}
	if !editor.MoveEnd() || editor.Cursor() != len([]rune(editor.String())) {
		t.Fatalf("expected end to move cursor to end, got %d", editor.Cursor())
	}
}

func TestEditorBackspaceDeleteAndUnicodeSafety(t *testing.T) {
	var editor Editor
	editor.Insert("a♟b")

	if !editor.Backspace() {
		t.Fatalf("expected backspace to delete last rune")
	}
	if editor.String() != "a♟" {
		t.Fatalf("expected unicode rune to remain intact, got %q", editor.String())
	}

	if !editor.MoveLeft() {
		t.Fatalf("expected move left to succeed")
	}
	if !editor.Delete() {
		t.Fatalf("expected delete to remove rune at cursor")
	}
	if editor.String() != "a" {
		t.Fatalf("expected delete to remove unicode rune, got %q", editor.String())
	}
}

func TestEditorClearLineAndDeleteWord(t *testing.T) {
	var editor Editor
	editor.Insert("move e2e4 now")

	if !editor.DeleteWordBackward() {
		t.Fatalf("expected delete word to succeed")
	}
	if editor.String() != "move e2e4 " {
		t.Fatalf("expected final word to be removed, got %q", editor.String())
	}

	if !editor.Clear() {
		t.Fatalf("expected clear to empty buffer")
	}
	if editor.String() != "" || editor.Cursor() != 0 {
		t.Fatalf("expected editor to reset, got %q at %d", editor.String(), editor.Cursor())
	}
}
