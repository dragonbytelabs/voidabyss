package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMotion_GG_ToFirstLine(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\nline3\nline4\n")

	// Start at line 2
	e.cy = 2
	e.cx = 3

	// Simulate pressing 'g' twice for standalone gg
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))

	// Should jump to first line, first column
	if e.cy != 0 || e.cx != 0 {
		t.Errorf("gg motion failed: expected cy=0, cx=0, got cy=%d, cx=%d", e.cy, e.cx)
	}
}

func TestMotion_Caret_FirstNonBlank(t *testing.T) {
	e := newTestEditor(t, "line1\n  \tindented line\nline3\n")

	// Move to line with leading whitespace
	e.cy = 1
	e.cx = 0

	// Simulate pressing '^'
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone))

	// Should move to 'i' in "indented"
	if e.cx != 3 {
		t.Errorf("^ motion failed: expected cx=3, got cx=%d", e.cx)
	}
}

func TestMotion_Paragraph_Forward(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\n\nline4\nline5\n")

	// Start at first line
	e.cy = 0
	e.cx = 3

	// Simulate pressing '}'
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone))

	// Should jump to line after blank (line4)
	if e.cy != 3 {
		t.Errorf("} motion failed: expected cy=3, got cy=%d", e.cy)
	}
	if e.cx != 0 {
		t.Errorf("} motion should reset cx to 0, got cx=%d", e.cx)
	}
}

func TestMotion_Paragraph_Backward(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\n\nline4\nline5\n")

	// Start at line4
	e.cy = 3
	e.cx = 3

	// Simulate pressing '{'
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, '{', tcell.ModNone))

	// Should jump back to line1 (beginning of previous paragraph)
	if e.cy != 0 {
		t.Errorf("{ motion failed: expected cy=0, got cy=%d", e.cy)
	}
	if e.cx != 0 {
		t.Errorf("{ motion should reset cx to 0, got cx=%d", e.cx)
	}
}

func TestMotion_WordForward(t *testing.T) {
	e := newTestEditor(t, "hello world test\n")

	// Start at beginning
	e.cy = 0
	e.cx = 0

	// Simulate pressing 'w'
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	// Should move to 'w' in "world"
	if e.cx != 6 {
		t.Errorf("w motion failed: expected cx=6, got cx=%d", e.cx)
	}
}

func TestMotion_WordBackward(t *testing.T) {
	e := newTestEditor(t, "hello world test\n")

	// Start at 'w' in world
	e.cy = 0
	e.cx = 6

	// Simulate pressing 'b'
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	// Should move to 'h' in "hello"
	if e.cx != 0 {
		t.Errorf("b motion failed: expected cx=0, got cx=%d", e.cx)
	}
}

func TestMotion_WordEnd(t *testing.T) {
	e := newTestEditor(t, "hello world test\n")

	// Start at beginning
	e.cy = 0
	e.cx = 0

	// Simulate pressing 'e'
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))

	// Should move to 'o' (last char of "hello")
	if e.cx != 4 {
		t.Errorf("e motion failed: expected cx=4, got cx=%d", e.cx)
	}
}
