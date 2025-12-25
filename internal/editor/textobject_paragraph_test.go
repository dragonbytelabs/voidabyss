package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTextObject_InnerParagraph(t *testing.T) {
	ed := newTestEditor(t, "first para\nsecond line\n\nthird para\n\nfifth")

	// Position cursor in first paragraph
	ed.cy = 1
	ed.cx = 0

	// dip - delete inner paragraph
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	got := ed.buffer.String()
	// Should delete the paragraph but leave blank line
	if got != "\nthird para\n\nfifth" && got != "third para\n\nfifth" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestTextObject_AroundParagraph(t *testing.T) {
	ed := newTestEditor(t, "para one\n\npara two\nline two\n\npara three")

	// Position cursor in middle paragraph
	ed.cy = 2
	ed.cx = 0

	// dap - delete around paragraph (includes trailing blank lines)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	got := ed.buffer.String()
	if got != "para one\n\npara three" {
		t.Fatalf("expected 'para one\\n\\npara three', got %q", got)
	}
}

func TestTextObject_YankParagraph(t *testing.T) {
	ed := newTestEditor(t, "first\nsecond\n\nthird")

	// Position cursor in first paragraph
	ed.cy = 0

	// yip - yank inner paragraph
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	yanked := ed.regs.unnamed.text
	if yanked != "first\nsecond\n" && yanked != "first\nsecond" {
		t.Fatalf("unexpected yank result: %q", yanked)
	}

	// Buffer unchanged
	if got := ed.buffer.String(); got != "first\nsecond\n\nthird" {
		t.Fatalf("buffer should be unchanged, got %q", got)
	}
}

func TestTextObject_SingleLineParagraph(t *testing.T) {
	ed := newTestEditor(t, "line1\n\nline2\n\nline3")

	// Position on line2 (single-line paragraph)
	ed.cy = 2

	// dip - delete inner paragraph
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	got := ed.buffer.String()
	if got != "line1\n\n\nline3" && got != "line1\n\nline3" {
		t.Fatalf("unexpected result: %q", got)
	}
}
