package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Swapchess/internal/app"
	"github.com/divijg19/Swapchess/internal/pieces"
	rendertext "github.com/divijg19/Swapchess/internal/render/text"
)

type model struct {
	session      *app.Session
	input        textinput.Model
	pieceCatalog *pieces.Catalog
}

func Run(debugRenderer string) error {
	program := tea.NewProgram(initialModel(debugRenderer))
	return program.Start()
}

func initialModel(debugRenderer string) model {
	session := app.NewSession(debugRenderer)
	input := textinput.New()
	input.Prompt = ""
	input.Focus()
	input.CharLimit = 64
	input.Placeholder = session.PromptPlaceholder()
	input.Width = 48

	return model{
		session:      session,
		input:        input,
		pieceCatalog: pieces.NewCatalog(filepath.Join("assets", "pieces")),
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
		case tea.KeyEsc:
			m.session.CancelTransient()
			m.input.SetValue("")
			m.syncInput()
			return m, nil
		case tea.KeyEnter:
			result := m.session.Submit(m.input.Value())
			m.input.SetValue("")
			m.syncInput()
			if result.Quit {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.session.Preview(m.input.Value())
	m.syncInput()
	return m, cmd
}

func (m model) View() string {
	if m.session.Width == 0 {
		m.syncInput()
	}

	frame := m.renderFull()
	if m.session.Height > 0 {
		lines := strings.Split(frame, "\n")
		if len(lines) > m.session.Height {
			frame = m.renderCompact(fmt.Sprintf("Need at least %d rows for board mode.", len(lines)))
		}
	}
	if m.session.Width > 0 && renderWidth(frame) > m.session.Width {
		frame = m.renderCompact(fmt.Sprintf("Need at least %d columns for board mode.", renderWidth(frame)))
	}
	if (m.session.Width > 0 && m.session.Width < rendertext.BoardWidth) || (m.session.Height > 0 && m.session.Height < 8) {
		frame = m.renderCompact("Viewport too small for full board mode.")
	}
	return frame
}

func (m model) renderFull() string {
	var out strings.Builder
	out.WriteString("SwapChess CLI\n")
	out.WriteString(m.renderBoard())
	out.WriteString(fmt.Sprintf("Turn: %s | Status: %s | Last: %s | Mode: %s\n", m.session.View.Turn.String(), rendertext.StatusLabel(m.session.View.Status), m.session.LastMoveNotation(), m.session.ModeLabel()))
	out.WriteString(clipLine(m.session.Message+" | Hint: "+m.session.Hint, maxInt(m.session.Width, 80)))
	out.WriteString("\n")
	out.WriteString(m.promptLine())
	return out.String()
}

func (m model) renderCompact(reason string) string {
	lines := []string{
		"SwapChess CLI",
		reason,
		fmt.Sprintf("Turn: %s | Status: %s | Mode: %s", m.session.View.Turn.String(), rendertext.StatusLabel(m.session.View.Status), m.session.ModeLabel()),
		m.session.Message,
		"Hint: " + m.session.Hint,
		m.promptLine(),
	}
	return strings.Join(lines, "\n")
}

func (m model) promptLine() string {
	return m.session.PromptLabel() + "> " + m.input.View()
}

func (m model) renderBoard() string {
	switch m.session.Renderer {
	case app.RendererEngine:
		return rendertext.RenderEngineBoard(m.session.Game, m.pieceCatalog, rendertext.BoardOptions{})
	default:
		return rendertext.RenderBoard(m.session.View, m.pieceCatalog, rendertext.BoardOptions{})
	}
}

func (m *model) syncInput() {
	labelWidth := len(m.session.PromptLabel()) + 2
	width := m.session.Width - labelWidth
	if width < 12 {
		width = 12
	}
	m.input.Width = width
	m.input.Placeholder = m.session.PromptPlaceholder()
	if m.session.InputMode == app.InputModePromotion {
		m.input.CharLimit = 8
	} else {
		m.input.CharLimit = 64
	}
}

func clipLine(line string, width int) string {
	if width <= 0 || len(line) <= width {
		return line
	}
	if width == 1 {
		return line[:1]
	}
	return line[:width-1] + "..."
}

func renderWidth(frame string) int {
	maxWidth := 0
	for _, line := range strings.Split(frame, "\n") {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
