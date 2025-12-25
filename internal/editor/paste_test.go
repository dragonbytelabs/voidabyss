package editor

import "testing"

func TestPasteCharwiseAfter(t *testing.T) {
	e := newTestEditor(t, "abc")
	e.cy, e.cx = 0, 1 // on 'b'

	// set unnamed register to "X"
	e.regs.unnamed = Register{kind: RegCharwise, text: "X"}

	e.pasteAfter()
	if got := e.buffer.String(); got != "abXc" {
		t.Fatalf("expected %q got %q", "abXc", got)
	}
}

func TestPasteLinewiseBelow(t *testing.T) {
	e := newTestEditor(t, "one\ntwo")
	e.cy, e.cx = 0, 0

	// Linewise content should include the newline
	e.regs.unnamed = Register{kind: RegLinewise, text: "HELLO\n"}

	e.pasteAfter()
	if got := e.buffer.String(); got != "one\nHELLO\ntwo" {
		t.Fatalf("unexpected result: %q", got)
	}
}
