package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
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
	appStyle        = lipgloss.NewStyle().Padding(0, appPaddingX)
	boardPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
	boardFocusPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("220"))
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
	session       *app.Session
	input         textinput.Model
	pieceCatalog  *pieces.Catalog
	focus         focusZone
	helpExpanded  bool
	moveLogScroll int
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
	input.Cursor.SetMode(cursor.CursorStatic)

	return model{
		session:      session,
		input:        input,
		pieceCatalog: pieces.NewCatalog(filepath.Join("assets", "pieces")),
		focus:        focusBoard,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.session.Resize(msg.Width, msg.Height)
		m.normalizeMoveLogScroll()
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
			m.normalizeMoveLogScroll()
			m.syncInput()
			return m, nil
		case tea.KeyEsc:
			if m.session.InputMode != app.InputModeCommand {
				m.session.CancelTransient()
				m.focus = focusBoard
				m.input.SetValue("")
				m.normalizeMoveLogScroll()
				m.syncInput()
				return m, nil
			}
			if m.focus == focusPrompt {
				m.focus = focusBoard
				m.normalizeMoveLogScroll()
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
				m.moveLogScroll = 0
				m.normalizeMoveLogScroll()
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
			m.moveLogScroll = 0
			m.normalizeMoveLogScroll()
			m.syncInput()
			return m, nil
		}

		switch msg.String() {
		case "?":
			m.helpExpanded = !m.helpExpanded
			m.normalizeMoveLogScroll()
			return m, nil
		case ":":
			m.focus = focusPrompt
			m.normalizeMoveLogScroll()
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
		m.normalizeMoveLogScroll()
		m.syncInput()
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.session.Width == 0 || m.session.Height == 0 {
		return "Initializing TUI..."
	}

	fullLayout, ok := m.layout()
	if !ok {
		return m.padToViewport(appStyle.Render(m.renderConstraintView("viewport below compact-layout threshold")))
	}

	rendered := appStyle.Render(m.renderFull(fullLayout))
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
			m.moveLogScroll = 0
		}
	case "pgup":
		m.scrollMoveLog(1)
	case "pgdown":
		m.scrollMoveLog(-1)
	case "home":
		m.scrollMoveLogToTop()
	case "end":
		m.moveLogScroll = 0
	}
	m.normalizeMoveLogScroll()
	return m, nil
}

func (m model) renderFull(layout layoutSpec) string {
	header := headerStyle.Width(layout.UsableWidth).Render("SwapChess TUI")

	boardPanel := m.renderBoardPanel(layout)
	gamePanel := m.renderRightWrappedPanel("Game State", m.gameLines(layout.RightBodyWidth), layout.RightWidth, layout.GameBodyLines, false)
	logPanel := m.renderRightPreparedPanel("Move Log", m.moveLogLines(layout.RightBodyWidth, layout.LogBodyLines), layout.RightWidth, layout.LogBodyLines, false)
	helpPanel := m.renderRightWrappedPanel("Help", m.helpLines(layout.RightBodyWidth), layout.RightWidth, layout.HelpBodyLines, m.helpExpanded)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, gamePanel, logPanel, helpPanel)
	mainRow := lipgloss.NewStyle().Width(layout.UsableWidth).Height(layout.MainHeight).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, boardPanel, strings.Repeat(" ", layoutGapWidth), rightColumn),
	)
	inputPanel := m.renderInputPanel(layout)

	return lipgloss.JoinVertical(lipgloss.Left, header, mainRow, inputPanel)
}

func (m model) renderConstraintView(reason string) string {
	usableWidth := maxInt(m.session.Width-(appPaddingX*2), 30)
	header := headerStyle.Width(usableWidth).Render("SwapChess TUI")

	infoLines := []string{
		"Full layout paused until the viewport is large enough.",
		alignedRow("Viewport", fmt.Sprintf("%dx%d", m.session.Width, m.session.Height), 10),
	}
	minWidth, minHeight := minimumViewport()
	infoLines = append(infoLines,
		alignedRow("Required", fmt.Sprintf("%dx%d", minWidth, minHeight), 10),
		alignedRow("Reason", reason, 10),
		"Resize the terminal; the board stays left-anchored and drives the full HUD layout.",
	)

	contentWidth := maxInt(usableWidth-panelChromeWidth, 1)
	infoBodyLines := maxInt(len(wrapPlainLines(infoLines, contentWidth)), 4)
	infoPanel := m.renderWrappedPanel("Viewport Constraints", infoLines, usableWidth, infoBodyLines, true)
	inputPanel := m.renderInputPanelForWidth(usableWidth)

	return lipgloss.JoinVertical(lipgloss.Left, header, infoPanel, inputPanel)
}

func (m model) renderBoardPanel(layout layoutSpec) string {
	options := rendertext.BoardOptions{
		Cursor:    &m.session.Cursor,
		Selected:  m.session.SelectedSquare(),
		CellWidth: layout.BoardCellWidth,
		RowHeight: layout.BoardRowHeight,
		Decorator: func(content string, file, rank int, square view.ViewSquare, selected, cursor bool) string {
			return rendertext.StyleBoardCell(content, file, rank, square, selected, cursor, m.focus == focusBoard)
		},
	}

	board := rendertext.RenderBoard(m.session.View, m.pieceCatalog, options)
	if m.session.Renderer == app.RendererEngine {
		board = rendertext.RenderEngineBoard(m.session.Game, m.pieceCatalog, options)
	}
	board = strings.TrimRight(board, "\n")
	boardLines := strings.Split(board, "\n")
	fillLines := maxInt(layout.MainHeight-layout.BoardOuterLines, 0)
	topFillLines := fillLines / 2
	bottomFillLines := fillLines - topFillLines
	for i := 0; i < topFillLines; i++ {
		boardLines = append([]string{""}, boardLines...)
	}
	for i := 0; i < bottomFillLines; i++ {
		boardLines = append(boardLines, "")
	}
	innerWidth := maxInt(layout.LeftWidth-boardPanelChromeWidth, 1)
	board = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(strings.Join(boardLines, "\n"))

	title := titleStyle.Render("Board")
	if m.focus == focusBoard {
		title = titleStyle.Foreground(lipgloss.Color("220")).Render("Board")
	}
	title = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(title)
	style := boardPanelStyle
	if m.focus == focusBoard {
		style = boardFocusPanelStyle
	}
	return style.Width(maxInt(layout.LeftWidth-boardPanelChromeWidth, 1)).Render(title + "\n" + board)
}

func (m model) gameFields() []infoField {
	fields := []infoField{
		{Label: "Turn", Value: m.session.View.Turn.String()},
		{Label: "Status", Value: rendertext.StatusLabel(m.session.View.Status)},
		{Label: "Cursor", Value: app.PositionString(m.session.Cursor)},
		{Label: "Select", Value: m.selectedLabel()},
		{Label: "Last", Value: m.session.LastMoveNotation()},
		{Label: "Castle", Value: castlingLabel(m.session.View.CastlingRights)},
	}
	if m.session.DebugRendererEnabled {
		fields = append(fields, infoField{Label: "Render", Value: string(m.session.Renderer)})
	}
	return fields
}

func (m model) gameLines(bodyWidth int) []string {
	return formatInfoLines(m.gameFields(), 6, bodyWidth)
}

func (m model) helpFields() []infoField {
	fields := []infoField{
		{Label: "Tab", Value: "focus"},
		{Label: ":", Value: "prompt"},
		{Label: "Enter", Value: "select"},
		{Label: "Esc", Value: "back"},
	}
	if m.helpExpanded {
		fields = append(fields,
			infoField{Label: "Log", Value: "PgUp/PgDn"},
			infoField{Label: "Move", Value: "keys"},
			infoField{Label: "u", Value: "undo"},
			infoField{Label: "Type", Value: "prompt"},
		)
		if m.session.DebugRendererEnabled {
			fields = append(fields, infoField{Label: "Debug", Value: "render"})
		} else {
			fields = append(fields, infoField{Label: "Alt", Value: "h j k l also move"})
		}
		return fields
	}
	fields = append(fields,
		infoField{Label: "u", Value: "undo"},
		infoField{Label: "?", Value: "help"},
	)
	return fields
}

func (m model) helpLines(bodyWidth int) []string {
	return formatInfoLines(m.helpFields(), 5, bodyWidth)
}

func (m model) moveLogLines(bodyWidth, limitLines int) []string {
	wrapped := m.wrappedMoveLogLines(bodyWidth)
	if len(wrapped) == 0 {
		return []string{"No moves yet."}
	}
	if limitLines <= 0 {
		return []string{wrapped[len(wrapped)-1]}
	}
	maxOffset := maxInt(len(wrapped)-limitLines, 0)
	offset := m.moveLogScroll
	if offset > maxOffset {
		offset = maxOffset
	}
	end := len(wrapped) - offset
	start := maxInt(end-limitLines, 0)
	return wrapped[start:end]
}

func (m model) moveLogEntries() []string {
	if len(m.session.MoveLog) == 0 {
		return nil
	}

	fullMoves := make([]string, 0, (len(m.session.MoveLog)+1)/2)
	for i := 0; i < len(m.session.MoveLog); i += 2 {
		fullMoveNumber := (i / 2) + 1
		white := formatMoveRecord(m.session.MoveLog[i])
		line := fmt.Sprintf("%02d. %s", fullMoveNumber, white)
		if i+1 < len(m.session.MoveLog) {
			line += "  " + formatMoveRecord(m.session.MoveLog[i+1])
		}
		fullMoves = append(fullMoves, line)
	}

	return fullMoves
}

func (m model) wrappedMoveLogLines(bodyWidth int) []string {
	entries := m.moveLogEntries()
	if len(entries) == 0 {
		return nil
	}
	return wrapPlainLines(entries, bodyWidth)
}

func (m model) renderInputPanel(layout layoutSpec) string {
	bodyWidth := maxInt(layout.UsableWidth-panelChromeWidth, 1)
	highlight := m.focus == focusPrompt || m.session.InputMode == app.InputModePromotion
	return m.renderPreparedPanel("Command Line", wrapMixedLines(m.inputPanelLines(), bodyWidth), layout.UsableWidth, layout.InputBodyLines, highlight)
}

func (m model) renderInputPanelForWidth(width int) string {
	bodyWidth := maxInt(width-panelChromeWidth, 1)
	bodyLines := measureMixedLines(m.inputMeasureLines(), bodyWidth)
	highlight := m.focus == focusPrompt || m.session.InputMode == app.InputModePromotion
	return m.renderPreparedPanel("Command Line", wrapMixedLines(m.inputPanelLines(), bodyWidth), width, bodyLines, highlight)
}

func (m model) inputPanelLines() []string {
	prompt := m.session.PromptLabel() + "> " + m.input.View()
	return []string{
		alignedDualRow("Mode", m.session.ModeLabel(), "Focus", string(m.focus), 7),
		m.session.Message,
		"Hint: " + m.session.Hint,
		prompt,
	}
}

func (m model) inputMeasureLines() []string {
	promptText := m.input.Value()
	if promptText == "" {
		promptText = m.session.PromptPlaceholder()
	}
	return []string{
		alignedDualRow("Mode", m.session.ModeLabel(), "Focus", string(m.focus), 7),
		m.session.Message,
		"Hint: " + m.session.Hint,
		m.session.PromptLabel() + "> " + promptText,
	}
}

func (m model) layoutContent() layoutContent {
	return layoutContent{
		GameLines:  m.gameLines(preferredRightBodyWidth),
		LogLines:   []string{"01. e2e4  e7e5", "02. g1f3  b8c6", "03. f1b5  a7a6", "04. b5a4  g8f6"},
		HelpLines:  m.helpLines(preferredRightBodyWidth + 14),
		InputLines: m.inputMeasureLines(),
	}
}

func (m model) layout() (layoutSpec, bool) {
	return computeLayoutForContent(m.session.Width, m.session.Height, m.layoutContent())
}

func (m model) renderWrappedPanel(title string, lines []string, width, bodyLines int, focused bool) string {
	contentWidth := maxInt(width-panelChromeWidth, 1)
	return m.renderPreparedPanel(title, wrapPlainLines(lines, contentWidth), width, bodyLines, focused)
}

func (m model) renderPreparedPanel(title string, lines []string, width, bodyLines int, focused bool) string {
	return m.renderPreparedPanelAligned(title, lines, width, bodyLines, focused, lipgloss.Left)
}

func (m model) renderRightWrappedPanel(title string, lines []string, width, bodyLines int, focused bool) string {
	contentWidth := maxInt(width-panelChromeWidth, 1)
	return m.renderRightPreparedPanel(title, wrapPlainLines(lines, contentWidth), width, bodyLines, focused)
}

func (m model) renderRightPreparedPanel(title string, lines []string, width, bodyLines int, focused bool) string {
	return m.renderPreparedPanelAligned(title, lines, width, bodyLines, focused, lipgloss.Left)
}

func (m model) renderPreparedPanelAligned(title string, lines []string, width, bodyLines int, focused bool, alignment lipgloss.Position) string {
	contentWidth := maxInt(width-panelChromeWidth, 1)
	prepared := fitLines(lines, bodyLines)
	prepared = alignLines(prepared, contentWidth, alignment)
	style := basePanelStyle
	if focused {
		style = focusPanelStyle
	}
	head := titleStyle.Width(contentWidth).Align(alignment).Render(title)
	body := strings.Join(prepared, "\n")
	return style.Width(maxInt(width-panelBorderWidth, 1)).Render(head + "\n" + body)
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

	promptPrefix := m.session.PromptLabel() + "> "
	bodyWidth := 48
	if layout, ok := m.layout(); ok {
		bodyWidth = layout.UsableWidth - panelChromeWidth
	} else if m.session.Width == 0 {
		bodyWidth = 48
	} else {
		bodyWidth = m.session.Width - (appPaddingX * 2) - panelChromeWidth
	}
	width := bodyWidth - lipgloss.Width(promptPrefix)
	if width < 12 {
		width = 12
	}
	m.input.Width = width
	for m.input.Width > 12 && lipgloss.Width(promptPrefix+m.input.View()) > bodyWidth {
		m.input.Width--
	}
}

func (m *model) toggleFocus() {
	if m.focus == focusBoard {
		m.focus = focusPrompt
		return
	}
	m.focus = focusBoard
}

func (m *model) normalizeMoveLogScroll() {
	layout, ok := m.layout()
	if !ok {
		m.moveLogScroll = 0
		return
	}
	maxOffset := maxInt(len(m.wrappedMoveLogLines(layout.RightBodyWidth))-layout.LogBodyLines, 0)
	if m.moveLogScroll < 0 {
		m.moveLogScroll = 0
	}
	if m.moveLogScroll > maxOffset {
		m.moveLogScroll = maxOffset
	}
}

func (m *model) scrollMoveLog(direction int) {
	layout, ok := m.layout()
	if !ok {
		return
	}
	page := maxInt(layout.LogBodyLines-1, 1)
	m.moveLogScroll += direction * page
}

func (m *model) scrollMoveLogToTop() {
	layout, ok := m.layout()
	if !ok {
		return
	}
	m.moveLogScroll = maxInt(len(m.wrappedMoveLogLines(layout.RightBodyWidth))-layout.LogBodyLines, 0)
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

func alignedRow(label, value string, labelWidth int) string {
	return fmt.Sprintf("%-*s %s", labelWidth, label, value)
}

func alignedDualRow(leftLabel, leftValue, rightLabel, rightValue string, labelWidth int) string {
	left := alignedRow(leftLabel, leftValue, labelWidth)
	right := alignedRow(rightLabel, rightValue, labelWidth)
	return left + "   " + right
}

func formatMoveRecord(record app.MoveRecord) string {
	move := record.Notation
	if record.SwapEvent == nil {
		return move
	}
	return fmt.Sprintf("%s [%s<->%s]", move, app.PositionString(record.SwapEvent.A), app.PositionString(record.SwapEvent.B))
}

func wrapPlainLines(lines []string, width int) []string {
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapPlainLine(line, width)...)
	}
	return wrapped
}

func wrapMixedLines(lines []string, width int) []string {
	wrapped := make([]string, 0, len(lines))
	for idx, line := range lines {
		if idx == len(lines)-1 {
			wrapped = append(wrapped, line)
			continue
		}
		wrapped = append(wrapped, wrapPlainLine(line, width)...)
	}
	return wrapped
}

func measureMixedLines(lines []string, width int) int {
	return len(wrapMixedLines(lines, width))
}

func wrapPlainLine(line string, width int) []string {
	if width <= 0 || lipgloss.Width(line) <= width {
		return []string{line}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, 2)
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		if lipgloss.Width(candidate) <= width {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
}

func padLines(lines []string, minLines int) []string {
	if len(lines) == 0 {
		lines = []string{""}
	}
	for len(lines) < minLines {
		lines = append(lines, "")
	}
	return lines
}

func fitLines(lines []string, bodyLines int) []string {
	if bodyLines <= 0 {
		return []string{""}
	}
	lines = padLines(lines, bodyLines)
	if len(lines) > bodyLines {
		return lines[:bodyLines]
	}
	return lines
}

func alignLines(lines []string, width int, alignment lipgloss.Position) []string {
	if width <= 0 {
		return lines
	}
	if alignment == lipgloss.Left {
		return lines
	}
	aligned := make([]string, 0, len(lines))
	style := lipgloss.NewStyle().Width(width).Align(alignment)
	for _, line := range lines {
		aligned = append(aligned, style.Render(line))
	}
	return aligned
}

type infoField struct {
	Label string
	Value string
}

func formatInfoLines(fields []infoField, labelWidth, bodyWidth int) []string {
	if bodyWidth <= 0 {
		return formatInfoMeasureLines(fields, labelWidth)
	}

	lines := make([]string, 0, len(fields))
	for i := 0; i < len(fields); {
		left := alignedRow(fields[i].Label, fields[i].Value, labelWidth)
		if i+1 < len(fields) {
			right := alignedRow(fields[i+1].Label, fields[i+1].Value, labelWidth)
			dual := left + "   " + right
			if lipgloss.Width(dual) <= bodyWidth {
				lines = append(lines, dual)
				i += 2
				continue
			}
		}
		lines = append(lines, left)
		i++
	}

	return lines
}

func formatInfoMeasureLines(fields []infoField, labelWidth int) []string {
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		lines = append(lines, alignedRow(field.Label, field.Value, labelWidth))
	}
	return lines
}
