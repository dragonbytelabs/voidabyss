package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestFindCharForward(t *testing.T) {
	ed := newTestEditor(t, "hello world")

	// Start at 'h'
	ed.cx = 0

	// Find 'o' forward
	pos := ed.findCharForward('o', false, 1)
	if pos != 4 {
		t.Fatalf("expected position 4, got %d", pos)
	}

	// Find second 'o'
	pos = ed.findCharForward('o', false, 2)
	if pos != 7 {
		t.Fatalf("expected position 7, got %d", pos)
	}

	// Till before 'o'
	pos = ed.findCharForward('o', true, 1)
	if pos != 3 {
		t.Fatalf("expected position 3 (till before o), got %d", pos)
	}

	// Not found
	pos = ed.findCharForward('z', false, 1)
	if pos != -1 {
		t.Fatalf("expected -1 for not found, got %d", pos)
	}
}

func TestFindCharBackward(t *testing.T) {
	ed := newTestEditor(t, "hello world")

	// Start at end
	ed.cx = 10

	// Find 'o' backward
	pos := ed.findCharBackward('o', false, 1)
	if pos != 7 {
		t.Fatalf("expected position 7, got %d", pos)
	}

	// Find second 'o' backward
	pos = ed.findCharBackward('o', false, 2)
	if pos != 4 {
		t.Fatalf("expected position 4, got %d", pos)
	}

	// Till before 'o'
	pos = ed.findCharBackward('o', true, 1)
	if pos != 8 {
		t.Fatalf("expected position 8 (till after o), got %d", pos)
	}
}

func TestCharFindMotions(t *testing.T) {
	ed := newTestEditor(t, "the quick brown fox")

	// Start at beginning
	ed.cx = 0

	// f motion - find 'q'
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	if ed.cx != 4 {
		t.Fatalf("f motion: expected position 4, got %d", ed.cx)
	}

	// ; - repeat find
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, ';', tcell.ModNone))

	// Should stay at 4 since there's no second 'q'
	if ed.cx != 4 {
		t.Fatalf("; motion: expected position 4, got %d", ed.cx)
	}

	// F motion - find 'e' backward
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'F', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))

	if ed.cx != 2 {
		t.Fatalf("F motion: expected position 2, got %d", ed.cx)
	}
}

func TestMatchingBracket(t *testing.T) {
	ed := newTestEditor(t, "func main() { return (1 + 2) }")

	// Position on opening ( at position 9
	ed.cx = 9

	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '%', tcell.ModNone))

	// Should jump to matching ) at position 10
	if ed.cx != 10 {
		t.Fatalf("expected position 10, got %d", ed.cx)
	}

	// Position on opening { at position 12
	ed.cx = 12

	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '%', tcell.ModNone))

	// Should jump to closing } - let's find it
	line := ed.getLine(0)
	expectedPos := len(line) - 1

	if ed.cx != expectedPos {
		t.Fatalf("expected position %d (closing brace), got %d. Line: %q", expectedPos, ed.cx, line)
	}

	// Jump back
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '%', tcell.ModNone))

	if ed.cx != 12 {
		t.Fatalf("expected position 12, got %d", ed.cx)
	}
}

func TestGGMotion(t *testing.T) {
	ed := newTestEditor(t, "line 1\nline 2\nline 3\nline 4\nline 5")

	// Start at line 3
	ed.cy = 2
	ed.cx = 0

	// gg - go to first line
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))

	if ed.cy != 0 {
		t.Fatalf("gg: expected line 0, got %d", ed.cy)
	}

	// 3G - go to line 3
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone))

	if ed.cy != 2 {
		t.Fatalf("3G: expected line 2, got %d", ed.cy)
	}

	// G - go to last line
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone))

	if ed.cy != 4 {
		t.Fatalf("G: expected line 4, got %d", ed.cy)
	}
}

func TestLineStartMotions(t *testing.T) {
	ed := newTestEditor(t, "  hello world")

	// Start in middle
	ed.cx = 8

	// 0 - go to column 0
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone))

	if ed.cx != 0 {
		t.Fatalf("0: expected position 0, got %d", ed.cx)
	}

	// ^ - go to first non-blank
	ed.cx = 8
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone))

	if ed.cx != 2 {
		t.Fatalf("^: expected position 2, got %d", ed.cx)
	}
}
