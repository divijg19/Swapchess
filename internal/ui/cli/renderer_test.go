package cli

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Swapchess/internal/app"
	"github.com/divijg19/Swapchess/internal/pieces"
)

func TestRendererMinimumSizeIsStableAcrossContent(t *testing.T) {
	renderer := newRenderer(pieces.NewCatalog(""))
	width, height := fullMinimumSize()

	short := app.NewSession("")
	shortFrame := renderer.Render(short, Editor{}, width, height)
	if shortFrame.Compact {
		t.Fatalf("expected full layout at minimum size")
	}

	long := app.NewSession("engine")
	long.Message = strings.Repeat("very long message ", 20)
	long.Hint = strings.Repeat("very long hint ", 20)
	var editor Editor
	editor.Insert(strings.Repeat("command ", 10))
	longFrame := renderer.Render(long, editor, width, height)
	if longFrame.Compact {
		t.Fatalf("expected full layout to remain stable with long content")
	}

	if len(shortFrame.Lines) != height || len(longFrame.Lines) != height {
		t.Fatalf("expected full layout height %d, got %d and %d", height, len(shortFrame.Lines), len(longFrame.Lines))
	}
}

func TestRendererCompactFallbackAtBoundaries(t *testing.T) {
	renderer := newRenderer(pieces.NewCatalog(""))
	width, height := fullMinimumSize()
	session := app.NewSession("")

	if frame := renderer.Render(session, Editor{}, width, height); frame.Compact {
		t.Fatalf("expected full layout at exact minimum size")
	}
	if frame := renderer.Render(session, Editor{}, width-1, height); !frame.Compact {
		t.Fatalf("expected compact layout when width is below minimum")
	}
	if frame := renderer.Render(session, Editor{}, width, height-1); !frame.Compact {
		t.Fatalf("expected compact layout when height is below minimum")
	}
}

func TestRendererClipsMessageAndHintWithinViewport(t *testing.T) {
	renderer := newRenderer(pieces.NewCatalog(""))
	width, height := fullMinimumSize()
	session := app.NewSession("")
	session.Message = strings.Repeat("message ", 40)
	session.Hint = strings.Repeat("hint ", 40)

	frame := renderer.Render(session, Editor{}, width, height)
	if len(frame.Lines) != height {
		t.Fatalf("expected full layout height %d, got %d", height, len(frame.Lines))
	}

	for _, line := range frame.Lines {
		if lipgloss.Width(line) > width {
			t.Fatalf("expected line width to fit viewport, got %d > %d for %q", lipgloss.Width(line), width, line)
		}
	}
}

func TestRendererCursorColumnIgnoresPlaceholderLength(t *testing.T) {
	renderer := newRenderer(pieces.NewCatalog(""))
	width, height := compactMinimumSize()

	plain := app.NewSession("")
	debug := app.NewSession("engine")

	plainFrame := renderer.Render(plain, Editor{}, width, height)
	debugFrame := renderer.Render(debug, Editor{}, width, height)

	want := lipgloss.Width(plain.PromptLabel() + "> ")
	if plainFrame.CursorColumn != want {
		t.Fatalf("expected plain cursor column %d, got %d", want, plainFrame.CursorColumn)
	}
	if debugFrame.CursorColumn != want {
		t.Fatalf("expected debug cursor column %d, got %d", want, debugFrame.CursorColumn)
	}
}
