package cli

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type KeyKind int

const (
	KeyText KeyKind = iota
	KeySubmit
	KeyQuit
	KeyCancel
	KeyBackspace
	KeyDelete
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyClearLine
	KeyDeleteWord
	KeyResize
)

type KeyEvent struct {
	Kind   KeyKind
	Text   string
	Width  int
	Height int
}

func parseControlByte(b byte) (KeyEvent, bool) {
	switch b {
	case 3:
		return KeyEvent{Kind: KeyQuit}, true
	case 8, 127:
		return KeyEvent{Kind: KeyBackspace}, true
	case 10, 13:
		return KeyEvent{Kind: KeySubmit}, true
	case 1:
		return KeyEvent{Kind: KeyHome}, true
	case 5:
		return KeyEvent{Kind: KeyEnd}, true
	case 21:
		return KeyEvent{Kind: KeyClearLine}, true
	case 23:
		return KeyEvent{Kind: KeyDeleteWord}, true
	default:
		return KeyEvent{}, false
	}
}

func parseEscapeSequence(seq []byte) (KeyEvent, bool) {
	switch string(seq) {
	case "\x1b":
		return KeyEvent{Kind: KeyCancel}, true
	case "\x1bOM":
		return KeyEvent{Kind: KeySubmit}, true
	case "\x1b[D":
		return KeyEvent{Kind: KeyLeft}, true
	case "\x1b[C":
		return KeyEvent{Kind: KeyRight}, true
	case "\x1b[H", "\x1bOH", "\x1b[1~", "\x1b[7~":
		return KeyEvent{Kind: KeyHome}, true
	case "\x1b[F", "\x1bOF", "\x1b[4~", "\x1b[8~":
		return KeyEvent{Kind: KeyEnd}, true
	case "\x1b[3~":
		return KeyEvent{Kind: KeyDelete}, true
	default:
		return parseCSIuSequence(string(seq))
	}
}

func parseCSIuSequence(seq string) (KeyEvent, bool) {
	if !strings.HasPrefix(seq, "\x1b[") || !strings.HasSuffix(seq, "u") {
		return KeyEvent{}, false
	}

	body := strings.TrimSuffix(strings.TrimPrefix(seq, "\x1b["), "u")
	parts := strings.Split(body, ";")
	if len(parts) == 0 || len(parts) > 2 {
		return KeyEvent{}, false
	}

	codePoint, err := strconv.Atoi(parts[0])
	if err != nil {
		return KeyEvent{}, false
	}

	modifier := 1
	if len(parts) == 2 {
		modifier, err = strconv.Atoi(parts[1])
		if err != nil {
			return KeyEvent{}, false
		}
	}

	switch codePoint {
	case 8, 127:
		return KeyEvent{Kind: KeyBackspace}, true
	case 10, 13:
		return KeyEvent{Kind: KeySubmit}, true
	case 27:
		return KeyEvent{Kind: KeyCancel}, true
	}

	if hasCtrlModifier(modifier) {
		switch codePoint {
		case 3, int('c'), int('C'):
			return KeyEvent{Kind: KeyQuit}, true
		case int('h'), int('H'):
			return KeyEvent{Kind: KeyBackspace}, true
		case int('m'), int('M'), int('j'), int('J'):
			return KeyEvent{Kind: KeySubmit}, true
		}
	}

	return KeyEvent{}, false
}

func hasCtrlModifier(modifier int) bool {
	if modifier < 1 {
		return false
	}
	return (modifier-1)&4 != 0
}

func parsePrintableChunk(reader *bufio.Reader, first byte) (string, error) {
	data := []byte{first}
	for reader.Buffered() > 0 {
		next, err := reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if len(next) == 0 || next[0] == 0x1b {
			break
		}
		if _, ok := parseControlByte(next[0]); ok {
			break
		}

		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		data = append(data, b)
	}

	if utf8.Valid(data) {
		return string(data), nil
	}

	r, _ := utf8.DecodeRune(data)
	if r == utf8.RuneError {
		return string(first), nil
	}
	return string(r), nil
}
