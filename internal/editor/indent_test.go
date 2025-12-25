package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestIndent_SingleLine(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Indent current line with >>
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))

	expected := "    hello world"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after >>: expected %q, got %q", expected, got)
	}
}

func TestIndent_MultipleLines(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\nline3")

	// Indent 2 lines with 2>>
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))

	expected := "    line1\n    line2\nline3"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after 2>>: expected %q, got %q", expected, got)
	}
}

func TestUnindent_SingleLine(t *testing.T) {
	e := newTestEditor(t, "    hello world")

	// Unindent current line with <<
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))

	expected := "hello world"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after <<: expected %q, got %q", expected, got)
	}
}

func TestUnindent_PartialIndent(t *testing.T) {
	e := newTestEditor(t, "  hello world")

	// Unindent line with only 2 spaces
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))

	expected := "hello world"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after <<: expected %q, got %q", expected, got)
	}
}

func TestIndent_Visual(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\nline3")

	// Enter visual line mode on line 0
	e.cy = 0
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone))

	// Move down to select 2 lines
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))

	// Indent with >
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))

	expected := "    line1\n    line2\nline3"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after visual >: expected %q, got %q", expected, got)
	}
}

func TestUnindent_Visual(t *testing.T) {
	e := newTestEditor(t, "    line1\n    line2\n    line3")

	// Enter visual line mode on line 0
	e.cy = 0
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone))

	// Move down to select 2 lines
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))

	// Unindent with <
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))

	expected := "line1\nline2\n    line3"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after visual <: expected %q, got %q", expected, got)
	}
}

func TestAutoIndent_SingleLine(t *testing.T) {
	e := newTestEditor(t, "  hello world")

	// Auto-indent with ==
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '=', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '=', tcell.ModNone))

	// Should normalize to 0 or 4 spaces (2 spaces rounds to 0)
	expected := "hello world"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after ==: expected %q, got %q", expected, got)
	}
}

func TestAutoIndent_MultipleLines(t *testing.T) {
	e := newTestEditor(t, "  line1\n      line2\n    line3")

	// Auto-indent 3 lines with 3==
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '=', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '=', tcell.ModNone))

	// 2 spaces -> 0, 6 spaces -> 8, 4 spaces -> 4
	expected := "line1\n        line2\n    line3"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after 3==: expected %q, got %q", expected, got)
	}
}

func TestIndent_PreservesCursorColumn(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Move cursor to middle of line
	e.cx = 5

	// Indent
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))

	expected := "    hello world"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after >>: expected %q, got %q", expected, got)
	}
}

func TestIndent_EmptyLine(t *testing.T) {
	e := newTestEditor(t, "line1\n\nline3")

	// Move to empty line
	e.cy = 1

	// Indent empty line
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone))

	expected := "line1\n    \nline3"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after >> on empty line: expected %q, got %q", expected, got)
	}
}

func TestUnindent_NoIndent(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Unindent line with no indentation
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone))

	// Should remain unchanged
	expected := "hello world"
	got := e.buffer.String()
	if got != expected {
		t.Errorf("after << on non-indented line: expected %q, got %q", expected, got)
	}
}
