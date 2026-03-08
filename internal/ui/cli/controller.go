package cli

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/divijg19/Swapchess/internal/app"
	"github.com/divijg19/Swapchess/internal/pieces"
)

type controller struct {
	terminal Terminal
	session  *app.Session
	editor   Editor
	renderer renderer
}

func newController(terminal Terminal, debugRenderer string) *controller {
	return &controller{
		terminal: terminal,
		session:  app.NewSession(debugRenderer),
		renderer: newRenderer(pieces.NewCatalog(filepath.Join("assets", "pieces"))),
	}
}

func (c *controller) Run() error {
	defer c.terminal.Close()

	width, height, err := c.terminal.Size()
	if err != nil {
		return err
	}
	c.session.Resize(width, height)

	if err := c.render(); err != nil {
		return err
	}

	for {
		event, err := c.terminal.NextEvent()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		shouldRender, quit := c.handleEvent(event)
		if shouldRender {
			if err := c.render(); err != nil {
				return err
			}
		}
		if quit {
			return nil
		}
	}
}

func (c *controller) render() error {
	frame := c.renderer.Render(c.session, c.editor, c.session.Width, c.session.Height)
	return c.terminal.Render(frame)
}

func (c *controller) handleEvent(event KeyEvent) (bool, bool) {
	switch event.Kind {
	case KeyResize:
		c.session.Resize(event.Width, event.Height)
		return true, false
	case KeyQuit:
		return false, true
	case KeySubmit:
		result := c.session.Submit(c.editor.String())
		if result.ClearInput {
			c.editor.Clear()
		}
		return !result.Quit, result.Quit
	case KeyCancel:
		if c.session.InputMode == app.InputModeCommand {
			return false, false
		}
		result := c.session.CancelTransient()
		if result.ClearInput {
			c.editor.Clear()
		}
		return true, false
	case KeyBackspace:
		if c.editor.Backspace() {
			c.session.Preview(c.editor.String())
			return true, false
		}
		return false, false
	case KeyDelete:
		if c.editor.Delete() {
			c.session.Preview(c.editor.String())
			return true, false
		}
		return false, false
	case KeyLeft:
		return c.editor.MoveLeft(), false
	case KeyRight:
		return c.editor.MoveRight(), false
	case KeyHome:
		return c.editor.MoveHome(), false
	case KeyEnd:
		return c.editor.MoveEnd(), false
	case KeyClearLine:
		if c.editor.Clear() {
			c.session.Preview(c.editor.String())
			return true, false
		}
		return false, false
	case KeyDeleteWord:
		if c.editor.DeleteWordBackward() {
			c.session.Preview(c.editor.String())
			return true, false
		}
		return false, false
	case KeyText:
		if c.editor.Insert(event.Text) {
			c.session.Preview(c.editor.String())
			return true, false
		}
		return false, false
	default:
		return false, false
	}
}
