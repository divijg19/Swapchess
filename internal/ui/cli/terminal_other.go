//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos)

package cli

func openTerminal() (Terminal, error) {
	return nil, ErrTerminalRequired
}
