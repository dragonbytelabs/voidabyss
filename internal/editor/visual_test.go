package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestVisualChar_Delete(t *testing.T) {
	ed := newTestEditor(t, "hello world")

	// Move to 'w' at position 6
	ed.cx = 6

	// Enter visual mode, select "world"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))

	// Delete selection
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	if got := ed.buffer.String(); got != "hello " {
		t.Fatalf("expected 'hello ', got %q", got)
	}

	if ed.mode != ModeNormal {
		t.Fatalf("expected ModeNormal after delete, got %v", ed.mode)
	}
}

func TestVisualChar_Yank(t *testing.T) {
	ed := newTestEditor(t, "test data")

	// Select "data" starting from position 5
	ed.cx = 5
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))

	// Yank
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))

	// Buffer should be unchanged
	if got := ed.buffer.String(); got != "test data" {
		t.Fatalf("expected 'test data', got %q", got)
	}

	// Register should have "data"
	if ed.regs.unnamed.text != "data" {
		t.Fatalf("expected register to have 'data', got %q", ed.regs.unnamed.text)
	}

	if ed.mode != ModeNormal {
		t.Fatalf("expected ModeNormal after yank, got %v", ed.mode)
	}
}

func TestVisualChar_Change(t *testing.T) {
	ed := newTestEditor(t, "foo bar")

	// Select "bar" starting from position 4
	ed.cx = 4
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))

	// Change (should delete and enter insert mode)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))

	if got := ed.buffer.String(); got != "foo " {
		t.Fatalf("after change, expected 'foo ', got %q", got)
	}

	if ed.mode != ModeInsert {
		t.Fatalf("expected ModeInsert after change, got %v", ed.mode)
	}

	// Type replacement text
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "foo baz" {
		t.Fatalf("after typing, expected 'foo baz', got %q", got)
	}
}

func TestVisualChar_BackwardSelection(t *testing.T) {
	ed := newTestEditor(t, "abcdef")

	// Start at position 4 (e), select backward
	ed.cx = 4
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))

	// Delete (should delete "cde")
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	if got := ed.buffer.String(); got != "abf" {
		t.Fatalf("expected 'abf', got %q", got)
	}
}

func TestVisualChar_ToggleToLinewise(t *testing.T) {
	ed := newTestEditor(t, "line1\nline2\nline3")

	// Start visual charwise
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))

	if ed.visualKind != VisualChar {
		t.Fatalf("expected VisualChar, got %v", ed.visualKind)
	}

	// Press V to toggle to linewise
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone))

	if ed.visualKind != VisualLine {
		t.Fatalf("expected VisualLine after V, got %v", ed.visualKind)
	}
}

func TestVisualLine_Delete(t *testing.T) {
	ed := newTestEditor(t, "first\nsecond\nthird")

	// Start on second line
	ed.cy = 1
	ed.cx = 0

	// Enter visual line mode
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone))

	// Delete line
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	if got := ed.buffer.String(); got != "first\nthird" {
		t.Fatalf("expected 'first\\nthird', got %q", got)
	}
}

func TestVisualLine_MultipleLines(t *testing.T) {
	ed := newTestEditor(t, "a\nb\nc\nd")

	// Start on line 1 (b)
	ed.cy = 1

	// Visual line and select down
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))

	// Delete (should remove "b\nc\n")
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	if got := ed.buffer.String(); got != "a\nd" {
		t.Fatalf("expected 'a\\nd', got %q", got)
	}
}
