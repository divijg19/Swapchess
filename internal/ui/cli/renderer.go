package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Swapchess/internal/app"
	"github.com/divijg19/Swapchess/internal/pieces"
	rendertext "github.com/divijg19/Swapchess/internal/render/text"
	"github.com/divijg19/Swapchess/view"
)

const (
	cliTitle           = "SwapChess CLI"
	fullStaticLines    = 5
	compactStaticLines = 5
	minPromptColumns   = 8
)

type Frame struct {
	Lines        []string
	CursorColumn int
	Compact      bool
}

type renderer struct {
	pieceCatalog *pieces.Catalog
}

func newRenderer(pieceCatalog *pieces.Catalog) renderer {
	return renderer{pieceCatalog: pieceCatalog}
}

func compactMinimumSize() (int, int) {
	minWidth := maxInt(len("move> ")+minPromptColumns, len(cliTitle))
	return minWidth, compactStaticLines
}

func fullMinimumSize() (int, int) {
	boardWidth, boardLines := rendertext.Dimensions()
	compactWidth, _ := compactMinimumSize()
	return maxInt(boardWidth, compactWidth), boardLines + fullStaticLines
}

func (r renderer) Render(session *app.Session, editor Editor, width, height int) Frame {
	if width < 1 {
		width = 1
	}

	fullMinWidth, fullMinHeight := fullMinimumSize()
	if width >= fullMinWidth && height >= fullMinHeight {
		return fitFrameHeight(r.renderFull(session, editor, width), height)
	}
	return fitFrameHeight(r.renderCompact(session, editor, width), height)
}

func (r renderer) renderFull(session *app.Session, editor Editor, width int) Frame {
	lines := []string{
		clipPlain(cliTitle, width),
	}
	lines = append(lines, r.boardLines(session)...)
	lines = append(lines,
		clipPlain(fullStatusLine(session), width),
		clipPlain("Message: "+session.Message, width),
		clipPlain("Hint: "+session.Hint, width),
	)
	prompt, cursor := promptLine(session, editor, width)
	lines = append(lines, prompt)

	return Frame{
		Lines:        lines,
		CursorColumn: clamp(cursor, 0, maxInt(width-1, 0)),
	}
}

func (r renderer) renderCompact(session *app.Session, editor Editor, width int) Frame {
	fullMinWidth, fullMinHeight := fullMinimumSize()
	summary := fmt.Sprintf("Turn: %s | Status: %s | Resize to: %dx%d", session.View.Turn.String(), rendertext.StatusLabel(session.View.Status), fullMinWidth, fullMinHeight)
	prompt, cursor := promptLine(session, editor, width)
	lines := []string{
		clipPlain(cliTitle, width),
		clipPlain(summary, width),
		clipPlain("Message: "+session.Message, width),
		clipPlain("Hint: "+session.Hint, width),
		prompt,
	}

	return Frame{
		Lines:        lines,
		CursorColumn: clamp(cursor, 0, maxInt(width-1, 0)),
		Compact:      true,
	}
}

func (r renderer) boardLines(session *app.Session) []string {
	options := rendertext.BoardOptions{
		Decorator: func(content string, file, rank int, square view.ViewSquare, selected, cursor bool) string {
			return rendertext.StyleBoardCell(content, file, rank, square, false, false, false)
		},
	}

	var board string
	switch session.Renderer {
	case app.RendererEngine:
		board = rendertext.RenderEngineBoard(session.Game, r.pieceCatalog, options)
	default:
		board = rendertext.RenderBoard(session.View, r.pieceCatalog, options)
	}

	return strings.Split(strings.TrimRight(board, "\n"), "\n")
}

func fullStatusLine(session *app.Session) string {
	return fmt.Sprintf("Turn: %s | Status: %s | Last: %s | Mode: %s", session.View.Turn.String(), rendertext.StatusLabel(session.View.Status), session.LastMoveNotation(), session.ModeLabel())
}

func promptLine(session *app.Session, editor Editor, width int) (string, int) {
	prefix := session.PromptLabel() + "> "
	prefixWidth := lipgloss.Width(prefix)
	if width <= prefixWidth {
		return clipPlain(prefix, width), clamp(prefixWidth, 0, maxInt(width-1, 0))
	}

	available := width - prefixWidth
	if editor.String() == "" {
		return prefix + clipPlain(session.PromptPlaceholder(), available), clamp(prefixWidth, 0, maxInt(width-1, 0))
	}

	visible, cursorOffset := visibleInput(editor, available)
	return prefix + visible, prefixWidth + cursorOffset
}

func visibleInput(editor Editor, width int) (string, int) {
	if width <= 0 || len(editor.buffer) == 0 {
		return "", 0
	}

	start := 0
	for start < editor.cursor && runeWidth(editor.buffer[start:editor.cursor]) > width {
		start++
	}

	end := editor.cursor
	for end < len(editor.buffer) && runeWidth(editor.buffer[start:end+1]) <= width {
		end++
	}

	return string(editor.buffer[start:end]), runeWidth(editor.buffer[start:editor.cursor])
}

func fitFrameHeight(frame Frame, height int) Frame {
	if len(frame.Lines) == 0 {
		frame.Lines = []string{""}
	}
	if height <= 0 || len(frame.Lines) <= height {
		return frame
	}

	frame.Lines = frame.Lines[len(frame.Lines)-height:]
	return frame
}

func clipPlain(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= width {
		return text
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	var out strings.Builder
	for _, r := range text {
		next := out.String() + string(r)
		if lipgloss.Width(next)+3 > width {
			break
		}
		out.WriteRune(r)
	}
	return out.String() + "..."
}

func runeWidth(runes []rune) int {
	return lipgloss.Width(string(runes))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
