package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMarks_SetAndJumpLine(t *testing.T) {
	e := newTestEditor(t, "line one\nline two\nline three\nline four")

	// Move to line 2 (index 1)
	e.cy = 1
	e.cx = 5

	// Set mark 'a' with ma
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	// Move somewhere else
	e.cy = 3
	e.cx = 2

	// Jump to mark 'a with 'a (line jump)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	// Should be on line 1, column 0 (line jump goes to start of line)
	if e.cy != 1 {
		t.Errorf("expected line 1, got %d", e.cy)
	}
	if e.cx != 0 {
		t.Errorf("expected column 0, got %d", e.cx)
	}
}

func TestMarks_SetAndJumpExact(t *testing.T) {
	e := newTestEditor(t, "line one\nline two\nline three\nline four")

	// Move to line 2, column 5
	e.cy = 1
	e.cx = 5

	// Set mark 'b' with mb
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	// Move somewhere else
	e.cy = 3
	e.cx = 2

	// Jump to mark 'b with `b (exact jump)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	// Should be at exact position (1, 5)
	if e.cy != 1 {
		t.Errorf("expected line 1, got %d", e.cy)
	}
	if e.cx != 5 {
		t.Errorf("expected column 5, got %d", e.cx)
	}
}

func TestMarks_MultipleMarks(t *testing.T) {
	e := newTestEditor(t, "aaa\nbbb\nccc\nddd\neee")

	// Set mark 'a' at line 0
	e.cy = 0
	e.cx = 1
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	// Set mark 'b' at line 2
	e.cy = 2
	e.cx = 2
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	// Set mark 'c' at line 4
	e.cy = 4
	e.cx = 0
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))

	// Jump to mark 'a
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	if e.cy != 0 || e.cx != 1 {
		t.Errorf("jump to 'a: expected (0,1), got (%d,%d)", e.cy, e.cx)
	}

	// Jump to mark 'b
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	if e.cy != 2 || e.cx != 2 {
		t.Errorf("jump to 'b: expected (2,2), got (%d,%d)", e.cy, e.cx)
	}

	// Jump to mark 'c
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	if e.cy != 4 || e.cx != 0 {
		t.Errorf("jump to 'c: expected (4,0), got (%d,%d)", e.cy, e.cx)
	}
}

func TestMarks_InvalidMark(t *testing.T) {
	e := newTestEditor(t, "line one\nline two")

	// Try to jump to non-existent mark
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))

	if e.statusMsg != "mark not set" {
		t.Errorf("expected 'mark not set' message, got %q", e.statusMsg)
	}
}

func TestJumpList_BackAndForward(t *testing.T) {
	e := newTestEditor(t, "line 0\nline 1\nline 2\nline 3\nline 4")

	// Start at line 0
	e.cy = 0
	e.cx = 0

	// Set mark at line 2
	e.cy = 2
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	// Set mark at line 4
	e.cy = 4
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	// Go back to line 0
	e.cy = 0

	// Jump to mark 'a (line 2) - adds line 0 to jump list
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	if e.cy != 2 {
		t.Fatalf("expected to be at line 2 after mark jump, got %d", e.cy)
	}

	// Jump to mark 'b (line 4) - adds line 2 to jump list
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	if e.cy != 4 {
		t.Fatalf("expected to be at line 4 after mark jump, got %d", e.cy)
	}

	// Now at line 4, jump list should have [line 0, line 2]
	// Ctrl-O should go back to line 2
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, tcell.ModNone))
	if e.cy != 2 {
		t.Errorf("after first Ctrl-O: expected line 2, got %d", e.cy)
	}

	// Ctrl-O again should go back to line 0
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, tcell.ModNone))
	if e.cy != 0 {
		t.Errorf("after second Ctrl-O: expected line 0, got %d", e.cy)
	}
}

func TestJumpList_Search(t *testing.T) {
	e := newTestEditor(t, "hello world\nfoo bar\nhello again")

	// Start at line 0
	e.cy = 0
	e.cx = 0

	// Search for "hello"
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// Should be at first "hello" (line 0)
	if e.cy != 0 {
		t.Fatalf("first search: expected line 0, got %d", e.cy)
	}

	// Press 'n' to find next (should jump to line 2 and add to jump list)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	if e.cy != 2 {
		t.Fatalf("after 'n': expected line 2, got %d", e.cy)
	}

	// Ctrl-O should go back to line 0
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, tcell.ModNone))
	if e.cy != 0 {
		t.Errorf("after Ctrl-O: expected line 0, got %d", e.cy)
	}
}

func TestJumpList_MarkJump(t *testing.T) {
	e := newTestEditor(t, "aaa\nbbb\nccc\nddd")

	// Start at line 0
	e.cy = 0
	e.cx = 0

	// Set mark at line 3
	e.cy = 3
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	// Move back to line 0
	e.cy = 0

	// Jump to mark (should add to jump list)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	if e.cy != 3 {
		t.Fatalf("after mark jump: expected line 3, got %d", e.cy)
	}

	// Ctrl-O should go back to line 0
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, tcell.ModNone))
	if e.cy != 0 {
		t.Errorf("after Ctrl-O: expected line 0, got %d", e.cy)
	}
}

func TestJumpList_DoubleTick(t *testing.T) {
	e := newTestEditor(t, "line 0\nline 1\nline 2\nline 3")

	// Start at line 0
	e.cy = 0

	// Set mark at line 0
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	// Jump to line 3
	e.cy = 3
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone))

	// '' should jump back to previous position (line 0)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone))

	if e.cy != 3 {
		t.Errorf("after '': expected to jump back to line 3, got %d", e.cy)
	}
}
