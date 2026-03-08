package cli

import "unicode"

type Editor struct {
	buffer []rune
	cursor int
}

func (e *Editor) String() string {
	return string(e.buffer)
}

func (e *Editor) Cursor() int {
	return e.cursor
}

func (e *Editor) Insert(text string) bool {
	runes := []rune(text)
	if len(runes) == 0 {
		return false
	}

	head := append([]rune{}, e.buffer[:e.cursor]...)
	head = append(head, runes...)
	e.buffer = append(head, e.buffer[e.cursor:]...)
	e.cursor += len(runes)
	return true
}

func (e *Editor) Backspace() bool {
	if e.cursor == 0 {
		return false
	}

	e.replaceRange(e.cursor-1, e.cursor)
	return true
}

func (e *Editor) Delete() bool {
	if e.cursor >= len(e.buffer) {
		return false
	}

	e.replaceRange(e.cursor, e.cursor+1)
	return true
}

func (e *Editor) MoveLeft() bool {
	if e.cursor == 0 {
		return false
	}

	e.cursor--
	return true
}

func (e *Editor) MoveRight() bool {
	if e.cursor >= len(e.buffer) {
		return false
	}

	e.cursor++
	return true
}

func (e *Editor) MoveHome() bool {
	if e.cursor == 0 {
		return false
	}

	e.cursor = 0
	return true
}

func (e *Editor) MoveEnd() bool {
	if e.cursor == len(e.buffer) {
		return false
	}

	e.cursor = len(e.buffer)
	return true
}

func (e *Editor) Clear() bool {
	if len(e.buffer) == 0 && e.cursor == 0 {
		return false
	}

	e.buffer = nil
	e.cursor = 0
	return true
}

func (e *Editor) DeleteWordBackward() bool {
	if e.cursor == 0 {
		return false
	}

	start := e.cursor
	for start > 0 && unicode.IsSpace(e.buffer[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(e.buffer[start-1]) {
		start--
	}

	e.replaceRange(start, e.cursor)
	return true
}

func (e *Editor) replaceRange(start, end int) {
	head := append([]rune{}, e.buffer[:start]...)
	e.buffer = append(head, e.buffer[end:]...)
	e.cursor = start
}
