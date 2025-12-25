package editor

import "testing"

func TestDeleteLines_dd(t *testing.T) {
	e := newTestEditor(t, "one\ntwo\nthree\n")
	e.cy, e.cx = 1, 0 // on "two"

	e.applyOperatorMotion('d', 'd', 1) // dd

	if got := e.buffer.String(); got != "one\nthree\n" {
		t.Fatalf("expected %q got %q", "one\nthree\n", got)
	}
	// reg1 should contain deleted linewise text (including newline since deleteLines slices line start to next start)
	if e.regs.numbered[1].text == "" {
		t.Fatalf("expected reg1 to be filled after dd")
	}
	if e.regs.numbered[1].kind != RegLinewise {
		t.Fatalf("expected reg1 kind linewise got %v", e.regs.numbered[1].kind)
	}
}

func TestYankLines_yy(t *testing.T) {
	e := newTestEditor(t, "one\ntwo\nthree\n")
	e.cy, e.cx = 0, 0 // on "one"

	e.applyOperatorMotion('y', 'y', 1) // yy

	// Linewise yank includes the newline
	if e.regs.numbered[0].text != "one\n" {
		t.Fatalf("expected reg0 yanked 'one\\n' got %q", e.regs.numbered[0].text)
	}
	if e.regs.numbered[0].kind != RegLinewise {
		t.Fatalf("expected linewise yank")
	}
}

func TestDeleteWord_dw(t *testing.T) {
	e := newTestEditor(t, "hello world")
	e.cy, e.cx = 0, 0 // at 'h'

	e.applyOperatorMotion('d', 'w', 1) // dw

	// Vim behavior: deletes "hello " (word + trailing space)
	if got := e.buffer.String(); got != "world" {
		t.Fatalf("expected %q got %q", "world", got)
	}

	// Register should contain exactly what was deleted
	if e.regs.numbered[1].text != "hello " {
		t.Fatalf("expected deleted text %q in reg1 got %q", "hello ", e.regs.numbered[1].text)
	}
}

func TestDeleteToBOL_d0(t *testing.T) {
	e := newTestEditor(t, "abcdef")
	e.cy, e.cx = 0, 3 // at 'd'

	e.applyOperatorMotion('d', '0', 1) // d0

	if got := e.buffer.String(); got != "def" {
		t.Fatalf("expected %q got %q", "def", got)
	}
}

func TestDeleteToEOL_dDollar(t *testing.T) {
	e := newTestEditor(t, "abcdef")
	e.cy, e.cx = 0, 2 // at 'c'

	e.applyOperatorMotion('d', '$', 1) // d$

	if got := e.buffer.String(); got != "ab" {
		t.Fatalf("expected %q got %q", "ab", got)
	}
}
