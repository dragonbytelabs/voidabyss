package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTextObject_InnerWord_Delete(t *testing.T) {
	ed := newTestEditor(t, "hello world test")

	// Position cursor on 'w' in "world"
	ed.cx = 6

	// diw - delete inner word
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	if got := ed.buffer.String(); got != "hello  test" {
		t.Fatalf("expected 'hello  test', got %q", got)
	}
}

func TestTextObject_AroundWord_Delete(t *testing.T) {
	ed := newTestEditor(t, "foo bar baz")

	// Position cursor on 'b' in "bar"
	ed.cx = 4

	// daw - delete a word (includes trailing space)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	if got := ed.buffer.String(); got != "foo baz" {
		t.Fatalf("expected 'foo baz', got %q", got)
	}
}

func TestTextObject_InnerWord_Yank(t *testing.T) {
	ed := newTestEditor(t, "quick brown fox")

	// Position cursor in "brown"
	ed.cx = 7

	// yiw - yank inner word
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	if ed.regs.unnamed.text != "brown" {
		t.Fatalf("expected register to have 'brown', got %q", ed.regs.unnamed.text)
	}

	// Buffer unchanged
	if got := ed.buffer.String(); got != "quick brown fox" {
		t.Fatalf("buffer should be unchanged, got %q", got)
	}
}

func TestTextObject_AroundWord_Change(t *testing.T) {
	ed := newTestEditor(t, "one two three")

	// Position on "two"
	ed.cx = 5

	// caw - change a word
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	if ed.mode != ModeInsert {
		t.Fatalf("expected ModeInsert, got %v", ed.mode)
	}

	if got := ed.buffer.String(); got != "one three" {
		t.Fatalf("expected 'one three', got %q", got)
	}

	// Type replacement
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if got := ed.buffer.String(); got != "one fourthree" {
		t.Fatalf("expected 'one fourthree', got %q", got)
	}
}

func TestTextObject_BigWord_Delete(t *testing.T) {
	ed := newTestEditor(t, "foo-bar baz_qux test")

	// Position on "foo-bar" (WORD includes punctuation)
	ed.cx = 2

	// diW - delete inner WORD
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'W', tcell.ModNone))

	if got := ed.buffer.String(); got != " baz_qux test" {
		t.Fatalf("expected ' baz_qux test', got %q", got)
	}
}

func TestTextObject_BigWord_Around(t *testing.T) {
	ed := newTestEditor(t, "alpha beta-gamma delta")

	// Position on "beta-gamma"
	ed.cx = 8

	// daW - delete a WORD (with trailing space)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'W', tcell.ModNone))

	if got := ed.buffer.String(); got != "alpha delta" {
		t.Fatalf("expected 'alpha delta', got %q", got)
	}
}

func TestTextObject_CursorOnWhitespace(t *testing.T) {
	ed := newTestEditor(t, "aaa   bbb")

	// Position cursor on whitespace between words
	ed.cx = 4

	// diw should find nearest word (searches right, then left)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	// Should delete "bbb"
	if got := ed.buffer.String(); got != "aaa   " {
		t.Fatalf("expected 'aaa   ', got %q", got)
	}
}

func TestTextObject_AtEndOfWord(t *testing.T) {
	ed := newTestEditor(t, "word1 word2")

	// Position on last char of "word1"
	ed.cx = 4

	// ciw - change inner word
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	if got := ed.buffer.String(); got != " word2" {
		t.Fatalf("expected ' word2', got %q", got)
	}
}

func TestTextObject_Count(t *testing.T) {
	ed := newTestEditor(t, "one two three four")

	// Position on "two"
	ed.cx = 4

	// 2daw - delete 2 words
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	// Should delete "two " then "three "
	if got := ed.buffer.String(); got != "one four" {
		t.Fatalf("expected 'one four', got %q", got)
	}
}

func TestTextObject_SingleChar(t *testing.T) {
	ed := newTestEditor(t, "a b c")

	// Position on "b"
	ed.cx = 2

	// diw on single char
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	if got := ed.buffer.String(); got != "a  c" {
		t.Fatalf("expected 'a  c', got %q", got)
	}
}
