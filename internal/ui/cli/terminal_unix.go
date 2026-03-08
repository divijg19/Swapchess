//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

const (
	pollIntervalMS = 100
	escapeWaitMS   = 16
)

type realTerminal struct {
	input  *os.File
	output *os.File
	owned  *os.File
	state  *term.State

	events chan KeyEvent
	errs   chan error
	done   chan struct{}

	signals chan os.Signal

	renderMu      sync.Mutex
	closeOnce     sync.Once
	renderedLines int
}

func openTerminal() (Terminal, error) {
	input := os.Stdin
	output := os.Stdout
	var owned *os.File

	if !term.IsTerminal(input.Fd()) || !term.IsTerminal(output.Fd()) {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to open controlling terminal", ErrTerminalRequired)
		}
		input = tty
		output = tty
		owned = tty
	}

	if !term.IsTerminal(input.Fd()) || !term.IsTerminal(output.Fd()) {
		if owned != nil {
			owned.Close()
		}
		return nil, ErrTerminalRequired
	}

	state, err := term.MakeRaw(input.Fd())
	if err != nil {
		if owned != nil {
			owned.Close()
		}
		return nil, fmt.Errorf("%w: %v", ErrTerminalRequired, err)
	}

	terminal := &realTerminal{
		input:   input,
		output:  output,
		owned:   owned,
		state:   state,
		events:  make(chan KeyEvent, 32),
		errs:    make(chan error, 1),
		done:    make(chan struct{}),
		signals: make(chan os.Signal, 8),
	}

	signal.Notify(terminal.signals, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
	go terminal.readLoop()
	go terminal.signalLoop()

	return terminal, nil
}

func (t *realTerminal) Size() (int, int, error) {
	width, height, err := term.GetSize(t.output.Fd())
	if err == nil {
		return width, height, nil
	}
	return term.GetSize(t.input.Fd())
}

func (t *realTerminal) NextEvent() (KeyEvent, error) {
	select {
	case event := <-t.events:
		return event, nil
	case err := <-t.errs:
		return KeyEvent{}, err
	case <-t.done:
		return KeyEvent{}, io.EOF
	}
}

func (t *realTerminal) Render(frame Frame) error {
	lines := frame.Lines
	if len(lines) == 0 {
		lines = []string{""}
	}

	t.renderMu.Lock()
	defer t.renderMu.Unlock()

	var out strings.Builder
	if t.renderedLines > 0 {
		out.WriteString("\r")
		if t.renderedLines > 1 {
			fmt.Fprintf(&out, "\x1b[%dA", t.renderedLines-1)
		}
		out.WriteString("\r\x1b[J")
	}

	out.WriteString(strings.Join(lines, "\r\n"))
	out.WriteString("\r")
	if frame.CursorColumn > 0 {
		fmt.Fprintf(&out, "\x1b[%dC", frame.CursorColumn)
	}

	if _, err := io.WriteString(t.output, out.String()); err != nil {
		return err
	}

	t.renderedLines = len(lines)
	return nil
}

func (t *realTerminal) Close() error {
	var closeErr error

	t.closeOnce.Do(func() {
		close(t.done)
		signal.Stop(t.signals)

		t.renderMu.Lock()
		defer t.renderMu.Unlock()

		if t.state != nil {
			closeErr = term.Restore(t.input.Fd(), t.state)
		}
		if t.renderedLines > 0 {
			if _, err := io.WriteString(t.output, "\r\n"); closeErr == nil {
				closeErr = err
			}
		}
		if t.owned != nil {
			if err := t.owned.Close(); closeErr == nil {
				closeErr = err
			}
		}
	})

	return closeErr
}

func (t *realTerminal) readLoop() {
	reader := bufio.NewReader(t.input)
	fd := int(t.input.Fd())

	for {
		select {
		case <-t.done:
			return
		default:
		}

		if reader.Buffered() == 0 {
			ready, err := pollReadable(fd, pollIntervalMS)
			if err != nil {
				t.sendErr(err)
				return
			}
			if !ready {
				continue
			}
		}

		b, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, os.ErrClosed) {
				return
			}
			t.sendErr(err)
			return
		}

		if b == 0x1b {
			seq, err := readEscapeSequence(reader, fd)
			if err != nil {
				t.sendErr(err)
				return
			}
			if event, ok := parseEscapeSequence(seq); ok {
				t.sendEvent(event)
			}
			continue
		}

		if event, ok := parseControlByte(b); ok {
			t.sendEvent(event)
			continue
		}

		text, err := parsePrintableChunk(reader, b)
		if err != nil {
			t.sendErr(err)
			return
		}
		t.sendEvent(KeyEvent{Kind: KeyText, Text: text})
	}
}

func (t *realTerminal) signalLoop() {
	for {
		select {
		case <-t.done:
			return
		case sig := <-t.signals:
			switch sig {
			case syscall.SIGWINCH:
				width, height, err := t.Size()
				if err != nil {
					t.sendErr(err)
					return
				}
				t.sendEvent(KeyEvent{Kind: KeyResize, Width: width, Height: height})
			case syscall.SIGINT, syscall.SIGTERM:
				t.sendEvent(KeyEvent{Kind: KeyQuit})
			}
		}
	}
}

func (t *realTerminal) sendEvent(event KeyEvent) {
	select {
	case <-t.done:
	case t.events <- event:
	}
}

func (t *realTerminal) sendErr(err error) {
	select {
	case <-t.done:
	case t.errs <- err:
	default:
	}
}

func readEscapeSequence(reader *bufio.Reader, fd int) ([]byte, error) {
	seq := []byte{0x1b}
	for len(seq) < 8 {
		if reader.Buffered() == 0 {
			ready, err := pollReadable(fd, escapeWaitMS)
			if err != nil {
				return nil, err
			}
			if !ready {
				break
			}
		}

		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		seq = append(seq, b)
	}
	return seq, nil
}

func pollReadable(fd int, timeoutMS int) (bool, error) {
	pollFds := []unix.PollFd{{
		Fd:     int32(fd),
		Events: unix.POLLIN,
	}}

	n, err := unix.Poll(pollFds, timeoutMS)
	if err != nil {
		if errors.Is(err, syscall.EINTR) {
			return false, nil
		}
		return false, err
	}
	return n > 0 && pollFds[0].Revents&unix.POLLIN != 0, nil
}
