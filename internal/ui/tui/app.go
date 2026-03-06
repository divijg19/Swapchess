package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Swapchess/internal/app"
	"github.com/divijg19/Swapchess/internal/pieces"
	rendertext "github.com/divijg19/Swapchess/internal/render/text"
	"github.com/divijg19/Swapchess/view"
)

type focusZone string

const (
	focusBoard  focusZone = "board"
	focusPrompt focusZone = "prompt"
)

var (
	appStyle       = lipgloss.NewStyle().Padding(0, appPaddingX)
	basePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)
	focusPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("220")).
			Padding(0, 1)
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("24")).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
)

type model struct {
	session      *app.Session
	input        textinput.Model
	pieceCatalog *pieces.Catalog
	focus        focusZone
	helpExpanded bool
}

func Run(debugRenderer string) error {
	program := tea.NewProgram(initialModel(debugRenderer), tea.WithAltScreen())
	return program.Start()
}

func initialModel(debugRenderer string) model {
	session := app.NewSession(debugRenderer)
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = session.PromptPlaceholder()
	input.CharLimit = 64
	input.Width = 48
	input.Blur()

	return model{
		session:      session,
		input:        input,
		pieceCatalog: pieces.NewCatalog(filepath.Join("assets", "pieces")),
		focus:        focusBoard,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.session.Resize(msg.Width, msg.Height)
		m.syncInput()
		return m, nil
	case tea.KeyMsg:
		if strings.EqualFold(msg.String(), "ctrl+c") {
			return m, tea.Quit
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			m.toggleFocus()
			m.syncInput()
			return m, nil
		case tea.KeyEsc:
			if m.session.InputMode != app.InputModeCommand {
				m.session.CancelTransient()
				m.focus = focusBoard
				m.input.SetValue("")
				m.syncInput()
				return m, nil
			}
			if m.focus == focusPrompt {
				m.focus = focusBoard
				m.syncInput()
				return m, nil
			}
			return m, nil
		case tea.KeyEnter:
			if m.focus == focusBoard {
				result := m.session.ActivateCursor()
				if result.Quit {
					return m, tea.Quit
				}
				if result.InputMode == app.InputModePromotion {
					m.focus = focusPrompt
				}
				m.syncInput()
				return m, nil
			}

			result := m.session.Submit(m.input.Value())
			m.input.SetValue("")
			if result.Quit {
				return m, tea.Quit
			}
			if result.InputMode == app.InputModePromotion {
				m.focus = focusPrompt
			}
			m.syncInput()
			return m, nil
		}

		switch msg.String() {
		case "?":
			m.helpExpanded = !m.helpExpanded
			return m, nil
		case ":":
			m.focus = focusPrompt
			m.syncInput()
			return m, nil
		}

		if m.focus == focusBoard {
			return m.handleBoardKey(msg)
		}
	}

	if m.focus == focusPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.session.Preview(m.input.Value())
		m.syncInput()
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.session.Width == 0 || m.session.Height == 0 {
		return "Initializing TUI..."
	}

	fullLayout, ok := computeLayout(m.session.Width, m.session.Height)
	if !ok {
		return m.padToViewport(appStyle.Render(m.renderConstraintView("viewport below full-layout threshold")))
	}

	rendered := appStyle.Render(m.renderFull(fullLayout))
	if lipgloss.Width(rendered) > m.session.Width || lipgloss.Height(rendered) > m.session.Height {
		return m.padToViewport(appStyle.Render(m.renderConstraintView("layout overflow detected after resize")))
	}
	return m.padToViewport(rendered)
}

func (m model) handleBoardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.session.MoveCursor(-1, 0)
	case "right", "l":
		m.session.MoveCursor(1, 0)
	case "up", "k":
		m.session.MoveCursor(0, 1)
	case "down", "j":
		m.session.MoveCursor(0, -1)
	case "u":
		if m.session.InputMode == app.InputModeCommand {
			m.session.Submit("undo")
		}
	}
	return m, nil
}

func (m model) renderFull(layout layoutSpec) string {
	header := headerStyle.Width(layout.UsableWidth).Render("SwapChess TUI")

	boardPanel := m.renderBoardPanel(layout)
	gamePanel := m.panel("Game State", strings.Join(m.gameLines(), "\n"), layout.RightWidth, gameBodyLines, false)
	logPanel := m.panel("Move Log", strings.Join(m.recentMoveLines(layout.LogBodyLines), "\n"), layout.RightWidth, layout.LogBodyLines, false)
	helpPanel := m.panel("Help", strings.Join(m.helpLines(), "\n"), layout.RightWidth, helpBodyLines, m.helpExpanded)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, gamePanel, logPanel, helpPanel)

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, boardPanel, strings.Repeat(" ", layoutGapWidth), rightColumn)
	inputPanel := m.renderInputPanel(layout.UsableWidth)

	return lipgloss.JoinVertical(lipgloss.Left, header, mainRow, inputPanel)
}

func (m model) renderConstraintView(reason string) string {
	usableWidth := maxInt(m.session.Width-(appPaddingX*2), 30)
	header := headerStyle.Width(usableWidth).Render("SwapChess TUI")

	infoLines := []string{
		"Full panel layout is temporarily disabled because the viewport is below the safe threshold.",
		fmt.Sprintf("Current viewport: %dx%d", m.session.Width, m.session.Height),
	}
	minWidth, minHeight := minimumViewport()
	infoLines = append(infoLines,
		fmt.Sprintf("Required minimum: %dx%d", minWidth, minHeight),
		"Reason: "+reason,
		"Resize the terminal to restore the full synced layout.",
		"Use --cli for a compact terminal mode.",
	)

	infoBodyLines := maxInt(m.session.Height-1-(inputBodyLines+3)-3, 4)
	infoPanel := m.panel("Viewport Constraints", strings.Join(infoLines, "\n"), usableWidth, infoBodyLines, true)
	inputPanel := m.renderInputPanel(usableWidth)

	return lipgloss.JoinVertical(lipgloss.Left, header, infoPanel, inputPanel)
}

func (m model) renderBoardPanel(layout layoutSpec) string {
	options := rendertext.BoardOptions{
		Cursor:   &m.session.Cursor,
		Selected: m.session.SelectedSquare(),
		Decorator: func(content string, file, rank int, square view.ViewSquare, selected, cursor bool) string {
			style := lipgloss.NewStyle().Width(1).Align(lipgloss.Center)
			switch {
			case selected && cursor:
				style = style.Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("160"))
			case selected:
				style = style.Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("94"))
			case cursor && m.focus == focusBoard:
				style = style.Bold(true).Foreground(lipgloss.Color("16")).Background(lipgloss.Color("220"))
			case cursor:
				style = style.Bold(true).Foreground(lipgloss.Color("220"))
			}
			return style.Render(content)
		},
	}

	board := rendertext.RenderBoard(m.session.View, m.pieceCatalog, options)
	if m.session.Renderer == app.RendererEngine {
		board = rendertext.RenderEngineBoard(m.session.Game, m.pieceCatalog, options)
	}

	return m.panel("Board", board, layout.LeftWidth, layout.BoardBodyLines, m.focus == focusBoard)
}

func (m model) renderInputPanel(width int) string {
	prompt := m.session.PromptLabel() + "> " + m.input.View()
	body := strings.Join([]string{
		"Input mode: " + m.session.ModeLabel() + " | Focus: " + string(m.focus),
		m.session.Message,
		"Hint: " + m.session.Hint,
		prompt,
	}, "\n")

	highlight := m.focus == focusPrompt || m.session.InputMode == app.InputModePromotion
	return m.panel("Command Line", body, width, inputBodyLines, highlight)
}

func (m model) gameLines() []string {
	lines := []string{
		fmt.Sprintf("Turn: %s", m.session.View.Turn.String()),
		fmt.Sprintf("Status: %s", rendertext.StatusLabel(m.session.View.Status)),
		fmt.Sprintf("Cursor: %s", app.PositionString(m.session.Cursor)),
		fmt.Sprintf("Selected: %s", m.selectedLabel()),
		fmt.Sprintf("Last move: %s", m.session.LastMoveNotation()),
		fmt.Sprintf("Castling: %s", castlingLabel(m.session.View.CastlingRights)),
		fmt.Sprintf("Swap suppressed: %t", m.session.View.SuppressNextSwap),
	}
	if m.session.DebugRendererEnabled {
		lines = append(lines, "Renderer: "+string(m.session.Renderer))
	}
	return lines
}

func (m model) recentMoveLines(limit int) []string {
	if len(m.session.MoveLog) == 0 {
		return []string{"No moves yet."}
	}
	start := len(m.session.MoveLog) - limit
	if start < 0 {
		start = 0
	}

	lines := make([]string, 0, len(m.session.MoveLog[start:]))
	for _, record := range m.session.MoveLog[start:] {
		line := fmt.Sprintf("%02d. %-5s %s", record.Index, record.Player.String(), record.Notation)
		if record.SwapEvent != nil {
			line += fmt.Sprintf(" [%s<->%s]", app.PositionString(record.SwapEvent.A), app.PositionString(record.SwapEvent.B))
		}
		lines = append(lines, line)
	}
	return lines
}

func (m model) helpLines() []string {
	lines := []string{
		"Tab: switch focus",
		": : focus prompt",
		"Enter: select or submit",
		"Esc: cancel/back",
		"u: undo from board",
		"Ctrl+C: quit",
	}
	if m.helpExpanded {
		lines = append(lines, "Arrows or h/j/k/l: move board cursor")
		lines = append(lines, "Type moves directly in the prompt")
		if m.session.DebugRendererEnabled {
			lines = append(lines, "Debug: renderer view|engine|toggle")
		}
	}
	return lines
}

func (m model) panel(title, body string, width, bodyLines int, focused bool) string {
	head := titleStyle.Render(title)
	renderWidth := maxInt(width-2, 1)
	body = clipMultiline(body, maxInt(renderWidth-2, 8))
	body = trimAndPadLines(body, bodyLines)
	style := basePanelStyle
	if focused {
		style = focusPanelStyle
	}
	return style.Width(renderWidth).Render(head + "\n" + body)
}

func (m *model) syncInput() {
	m.session.Preview(m.input.Value())
	m.input.Placeholder = m.session.PromptPlaceholder()
	if m.focus == focusPrompt {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
	if m.session.InputMode == app.InputModePromotion {
		m.input.CharLimit = 8
	} else {
		m.input.CharLimit = 64
	}

	width := m.session.Width - 12
	if width < 12 {
		width = 12
	}
	m.input.Width = width
}

func (m *model) toggleFocus() {
	if m.focus == focusBoard {
		m.focus = focusPrompt
		return
	}
	m.focus = focusBoard
}

func (m model) padToViewport(rendered string) string {
	if lipgloss.Height(rendered) < m.session.Height {
		rendered += strings.Repeat("\n", m.session.Height-lipgloss.Height(rendered))
	}
	return rendered
}

func (m model) selectedLabel() string {
	if selected := m.session.SelectedSquare(); selected != nil {
		return app.PositionString(*selected)
	}
	return "-"
}

func castlingLabel(rights view.CastlingRights) string {
	var white, black string
	if rights.WhiteKingSide {
		white += "K"
	}
	if rights.WhiteQueenSide {
		white += "Q"
	}
	if rights.BlackKingSide {
		black += "K"
	}
	if rights.BlackQueenSide {
		black += "Q"
	}
	if white == "" {
		white = "-"
	}
	if black == "" {
		black = "-"
	}
	return fmt.Sprintf("W:%s B:%s", white, black)
}

func clipMultiline(value string, width int) string {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = clipLine(line, width)
	}
	return strings.Join(lines, "\n")
}

func trimAndPadLines(value string, lines int) string {
	if lines <= 0 {
		return ""
	}
	parts := strings.Split(value, "\n")
	if len(parts) > lines {
		parts = parts[:lines]
	}
	for len(parts) < lines {
		parts = append(parts, "")
	}
	return strings.Join(parts, "\n")
}

func clipLine(value string, width int) string {
	if width <= 0 || lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "..."
}
