package cli

import "errors"

var ErrTerminalRequired = errors.New("CLI mode requires a real terminal")

type terminalOpener func() (Terminal, error)

func Run(debugRenderer string) error {
	return run(debugRenderer, openTerminal)
}

func run(debugRenderer string, open terminalOpener) error {
	terminal, err := open()
	if err != nil {
		return err
	}

	return newController(terminal, debugRenderer).Run()
}
