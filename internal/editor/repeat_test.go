package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDotRepeat_InsertI(t *testing.T) {
	ed := newTestEditor(t, "")

	// Type 'i' to enter insert mode, type "hello", press Esc
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "hello" {
		t.Fatalf("after insert, expected 'hello', got %q", got)
	}

	// Press '.' to repeat
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))

	if got := ed.buffer.String(); got != "hellohello" {
		t.Fatalf("after dot-repeat, expected 'hellohello', got %q", got)
	}
}

func TestDotRepeat_InsertA(t *testing.T) {
	ed := newTestEditor(t, "test")

	// Move to start, press 'a', type "x", Esc
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "txest" {
		t.Fatalf("after 'a', expected 'txest', got %q", got)
	}

	// Move to position 2, press '.'
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))

	if got := ed.buffer.String(); got != "txexst" {
		t.Fatalf("after dot-repeat at pos 2, expected 'txexst', got %q", got)
	}
}

func TestDotRepeat_InsertO(t *testing.T) {
	ed := newTestEditor(t, "first\nlast")

	// Press 'o', type "middle", Esc
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "first\nmiddle\nlast" {
		t.Fatalf("after 'o', expected 'first\\nmiddle\\nlast', got %q", got)
	}

	// Press '.' to add another line
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))

	if got := ed.buffer.String(); got != "first\nmiddle\nmiddle\nlast" {
		t.Fatalf("after dot-repeat, expected 'first\\nmiddle\\nmiddle\\nlast', got %q", got)
	}
}

func TestDotRepeat_InsertWithNewline(t *testing.T) {
	ed := newTestEditor(t, "")

	// Press 'i', type "first", Enter, "second", Esc
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "first\nsecond" {
		t.Fatalf("after insert with newline, expected 'first\\nsecond', got %q", got)
	}

	// Press '.' to repeat
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))

	if got := ed.buffer.String(); got != "first\nsecondfirst\nsecond" {
		t.Fatalf("after dot-repeat, expected 'first\\nsecondfirst\\nsecond', got %q", got)
	}
}

func TestDotRepeat_InsertWithBackspace(t *testing.T) {
	ed := newTestEditor(t, "")

	// Press 'i', type "hello", backspace twice (removes "lo"), type "p", Esc
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "help" {
		t.Fatalf("after insert with backspace, expected 'help', got %q", got)
	}

	// Press '.' to repeat
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))

	// Should replay the exact same sequence: type hello, backspace twice, type p
	if got := ed.buffer.String(); got != "helphelp" {
		t.Fatalf("after dot-repeat, expected 'helphelp', got %q", got)
	}
}

func TestDotRepeat_MultipleInserts(t *testing.T) {
	ed := newTestEditor(t, "")

	// First insert: type "a"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	// Second insert: type "b" (should override the repeat)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "ab" {
		t.Fatalf("after two inserts, expected 'ab', got %q", got)
	}

	// Press '.' - should repeat last insert ("b")
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))

	if got := ed.buffer.String(); got != "abb" {
		t.Fatalf("after dot-repeat, expected 'abb', got %q", got)
	}
}
