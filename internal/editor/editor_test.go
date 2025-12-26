package editor

import (
	"testing"

	"github.com/dragonbytelabs/voidabyss/core/buffer"
	"github.com/gdamore/tcell/v2"
)

func newTestEditor(t *testing.T, txt string) *Editor {
	t.Helper()

	s := tcell.NewSimulationScreen("UTF-8")
	if s == nil {
		t.Fatalf("simulation screen is nil")
	}
	if err := s.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	s.SetSize(80, 24)

	e := &Editor{
		s:           s,
		buffer:      buffer.NewFromString(txt),
		mode:        ModeNormal,
		indentWidth: 4,
	}
	e.regs.named = make(map[rune]Register)
	e.marks = make(map[rune]Mark)
	e.macros = make(map[rune]Macro)
	e.jumpList = make([]JumpListEntry, 0, 100)
	e.jumpListIndex = -1
	return e
}

func TestLineStartsAndGetLine(t *testing.T) {
	e := newTestEditor(t, "one\ntwo\nthree")
	starts := e.lineStarts()

	if len(starts) != 3 {
		t.Fatalf("expected 3 line starts got %d (%v)", len(starts), starts)
	}

	if got := e.getLine(0); got != "one" {
		t.Fatalf("line0 expected 'one' got %q", got)
	}
	if got := e.getLine(1); got != "two" {
		t.Fatalf("line1 expected 'two' got %q", got)
	}
	if got := e.getLine(2); got != "three" {
		t.Fatalf("line2 expected 'three' got %q", got)
	}
}

func TestPosFromCursorAndSetCursorFromPos(t *testing.T) {
	e := newTestEditor(t, "ab\ncde\nf")
	// line 1 "cde", col 2 -> rune offset: "ab\n" (3) + 2 = 5
	e.cy = 1
	e.cx = 2
	pos := e.posFromCursor()
	if pos != 5 {
		t.Fatalf("expected pos 5 got %d", pos)
	}

	e.setCursorFromPos(5)
	if e.cy != 1 || e.cx != 2 {
		t.Fatalf("expected (1,2) got (%d,%d)", e.cy, e.cx)
	}

	// clamp past end
	e.setCursorFromPos(999)
	if e.posFromCursor() != e.buffer.Len() {
		t.Fatalf("expected clamped pos == len (%d) got %d", e.buffer.Len(), e.posFromCursor())
	}
}
