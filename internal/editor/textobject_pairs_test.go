package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTextObject_InnerQuotes(t *testing.T) {
	ed := newTestEditor(t, `hello "world" test`)

	// Position cursor inside quotes
	ed.cx = 8

	// di" - delete inner quotes
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone))

	if got := ed.buffer.String(); got != `hello "" test` {
		t.Fatalf("expected 'hello \"\" test', got %q", got)
	}
}

func TestTextObject_AroundQuotes(t *testing.T) {
	ed := newTestEditor(t, `foo "bar" baz`)

	// Position cursor inside quotes
	ed.cx = 6

	// da" - delete around quotes (including quotes)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone))

	if got := ed.buffer.String(); got != "foo  baz" {
		t.Fatalf("expected 'foo  baz', got %q", got)
	}
}

func TestTextObject_InnerParens(t *testing.T) {
	ed := newTestEditor(t, "func(arg1, arg2)")

	// Position cursor inside parens
	ed.cx = 6

	// di( - delete inner parentheses
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '(', tcell.ModNone))

	if got := ed.buffer.String(); got != "func()" {
		t.Fatalf("expected 'func()', got %q", got)
	}
}

func TestTextObject_AroundParens(t *testing.T) {
	ed := newTestEditor(t, "call(x)")

	// Position cursor inside parens
	ed.cx = 6

	// da) - delete around parentheses (same as da()
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, ')', tcell.ModNone))

	if got := ed.buffer.String(); got != "call" {
		t.Fatalf("expected 'call', got %q", got)
	}
}

func TestTextObject_InnerBraces(t *testing.T) {
	ed := newTestEditor(t, "if true { code here }")

	// Position cursor inside braces
	ed.cx = 12

	// di{ - delete inner braces
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '{', tcell.ModNone))

	if got := ed.buffer.String(); got != "if true {}" {
		t.Fatalf("expected 'if true {}', got %q", got)
	}
}

func TestTextObject_AroundBraces(t *testing.T) {
	ed := newTestEditor(t, "x{y}")

	// Position cursor inside braces
	ed.cx = 2

	// da} - delete around braces
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone))

	if got := ed.buffer.String(); got != "x" {
		t.Fatalf("expected 'x', got %q", got)
	}
}

func TestTextObject_InnerBrackets(t *testing.T) {
	ed := newTestEditor(t, "arr[index]")

	// Position cursor inside brackets
	ed.cx = 5

	// di[ - delete inner brackets
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModNone))

	if got := ed.buffer.String(); got != "arr[]" {
		t.Fatalf("expected 'arr[]', got %q", got)
	}
}

func TestTextObject_AroundBrackets(t *testing.T) {
	ed := newTestEditor(t, "list[0]")

	// Position cursor inside brackets
	ed.cx = 5

	// da] - delete around brackets
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, ']', tcell.ModNone))

	if got := ed.buffer.String(); got != "list" {
		t.Fatalf("expected 'list', got %q", got)
	}
}

func TestTextObject_NestedParens(t *testing.T) {
	ed := newTestEditor(t, "outer(inner(x))")

	// Position cursor in inner parens
	ed.cx = 13

	// di( - should delete content of inner parens
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '(', tcell.ModNone))

	if got := ed.buffer.String(); got != "outer(inner())" {
		t.Fatalf("expected 'outer(inner())', got %q", got)
	}
}

func TestTextObject_YankInnerQuotes(t *testing.T) {
	ed := newTestEditor(t, `text "copy this" more`)

	// Position cursor inside quotes
	ed.cx = 8

	// yi" - yank inner quotes
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone))

	if ed.regs.unnamed.text != "copy this" {
		t.Fatalf("expected 'copy this' in register, got %q", ed.regs.unnamed.text)
	}

	// Buffer should be unchanged
	if got := ed.buffer.String(); got != `text "copy this" more` {
		t.Fatalf("buffer should be unchanged, got %q", got)
	}
}

func TestTextObject_ChangeInnerParens(t *testing.T) {
	ed := newTestEditor(t, "func(old)")

	// Position cursor inside parens
	ed.cx = 6

	// ci( - change inner parens
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '(', tcell.ModNone))

	if ed.mode != ModeInsert {
		t.Fatalf("expected ModeInsert, got %v", ed.mode)
	}

	if got := ed.buffer.String(); got != "func()" {
		t.Fatalf("expected 'func()', got %q", got)
	}
}
