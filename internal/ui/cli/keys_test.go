package cli

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseControlByteMappings(t *testing.T) {
	cases := []struct {
		name string
		b    byte
		kind KeyKind
	}{
		{name: "ctrl-c", b: 3, kind: KeyQuit},
		{name: "ctrl-j", b: 10, kind: KeySubmit},
		{name: "ctrl-m", b: 13, kind: KeySubmit},
		{name: "ctrl-h", b: 8, kind: KeyBackspace},
		{name: "del", b: 127, kind: KeyBackspace},
		{name: "ctrl-a", b: 1, kind: KeyHome},
		{name: "ctrl-e", b: 5, kind: KeyEnd},
		{name: "ctrl-u", b: 21, kind: KeyClearLine},
		{name: "ctrl-w", b: 23, kind: KeyDeleteWord},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event, ok := parseControlByte(tc.b)
			if !ok {
				t.Fatalf("expected control byte %d to map", tc.b)
			}
			if event.Kind != tc.kind {
				t.Fatalf("expected kind %v, got %v", tc.kind, event.Kind)
			}
		})
	}
}

func TestParseEscapeSequenceMappings(t *testing.T) {
	cases := []struct {
		seq  string
		kind KeyKind
	}{
		{seq: "\x1b", kind: KeyCancel},
		{seq: "\x1bOM", kind: KeySubmit},
		{seq: "\x1b[D", kind: KeyLeft},
		{seq: "\x1b[C", kind: KeyRight},
		{seq: "\x1b[H", kind: KeyHome},
		{seq: "\x1b[F", kind: KeyEnd},
		{seq: "\x1b[3~", kind: KeyDelete},
		{seq: "\x1b[127u", kind: KeyBackspace},
		{seq: "\x1b[13u", kind: KeySubmit},
		{seq: "\x1b[99;5u", kind: KeyQuit},
	}

	for _, tc := range cases {
		event, ok := parseEscapeSequence([]byte(tc.seq))
		if !ok {
			t.Fatalf("expected escape sequence %q to map", tc.seq)
		}
		if event.Kind != tc.kind {
			t.Fatalf("expected kind %v for %q, got %v", tc.kind, tc.seq, event.Kind)
		}
	}
}

func TestParseEscapeSequenceCSIuCtrlVariants(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		kind KeyKind
	}{
		{name: "ctrl-h lower", seq: "\x1b[104;5u", kind: KeyBackspace},
		{name: "ctrl-h upper", seq: "\x1b[72;5u", kind: KeyBackspace},
		{name: "ctrl-j lower", seq: "\x1b[106;5u", kind: KeySubmit},
		{name: "ctrl-m upper", seq: "\x1b[77;5u", kind: KeySubmit},
		{name: "ctrl-c upper", seq: "\x1b[67;5u", kind: KeyQuit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event, ok := parseEscapeSequence([]byte(tc.seq))
			if !ok {
				t.Fatalf("expected escape sequence %q to map", tc.seq)
			}
			if event.Kind != tc.kind {
				t.Fatalf("expected kind %v for %q, got %v", tc.kind, tc.seq, event.Kind)
			}
		})
	}
}

func TestParseEscapeSequenceRejectsUnknownCSIu(t *testing.T) {
	if _, ok := parseEscapeSequence([]byte("\x1b[120;5u")); ok {
		t.Fatalf("expected unsupported CSI-u sequence to remain unmapped")
	}
}

func TestParsePrintableChunkSupportsBufferedPaste(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("e2e4\n"))
	first, err := reader.ReadByte()
	if err != nil {
		t.Fatalf("expected to read first byte: %v", err)
	}

	text, err := parsePrintableChunk(reader, first)
	if err != nil {
		t.Fatalf("parsePrintableChunk returned error: %v", err)
	}
	if text != "e2e4" {
		t.Fatalf("expected pasted text e2e4, got %q", text)
	}

	next, err := reader.ReadByte()
	if err != nil {
		t.Fatalf("expected control byte to remain buffered: %v", err)
	}
	if next != '\n' {
		t.Fatalf("expected newline to remain unread, got %q", next)
	}
}
