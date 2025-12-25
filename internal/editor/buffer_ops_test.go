package editor

import "testing"

func TestInsertBackspaceNewline(t *testing.T) {
	e := newTestEditor(t, "")
	e.mode = ModeInsert

	// Insert 'a'
	e.insertRune('a')
	if got := e.buffer.String(); got != "a" {
		t.Fatalf("expected %q got %q", "a", got)
	}
	if e.cx != 1 || e.cy != 0 {
		t.Fatalf("cursor expected (0,1) got (%d,%d)", e.cy, e.cx)
	}

	// Newline
	e.newline()
	if got := e.buffer.String(); got != "a\n" {
		t.Fatalf("expected %q got %q", "a\n", got)
	}
	if e.cy != 1 || e.cx != 0 {
		t.Fatalf("cursor expected (1,0) got (%d,%d)", e.cy, e.cx)
	}

	// Insert 'b' on new line
	e.insertRune('b')
	if got := e.buffer.String(); got != "a\nb" {
		t.Fatalf("expected %q got %q", "a\nb", got)
	}

	// Backspace deletes 'b'
	e.backspace()
	if got := e.buffer.String(); got != "a\n" {
		t.Fatalf("expected %q got %q", "a\n", got)
	}
}
