package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestWriteYankSetsUnnamedAndZero(t *testing.T) {
	e := newTestEditor(t, "hello")
	e.cy, e.cx = 0, 0

	e.writeYank(Register{kind: RegCharwise, text: "X"})
	if e.regs.numbered[0].text != "X" {
		t.Fatalf("register 0 expected 'X' got %q", e.regs.numbered[0].text)
	}
	if e.regs.unnamed.text != "X" {
		t.Fatalf("unnamed expected 'X' got %q", e.regs.unnamed.text)
	}
}

func TestWriteDeleteShiftsNumberedRegisters(t *testing.T) {
	e := newTestEditor(t, "hello")

	e.writeDelete(Register{kind: RegCharwise, text: "A"})
	if e.regs.numbered[1].text != "A" {
		t.Fatalf("reg1 expected A got %q", e.regs.numbered[1].text)
	}

	e.writeDelete(Register{kind: RegCharwise, text: "B"})
	if e.regs.numbered[1].text != "B" {
		t.Fatalf("reg1 expected B got %q", e.regs.numbered[1].text)
	}
	if e.regs.numbered[2].text != "A" {
		t.Fatalf("reg2 expected A got %q", e.regs.numbered[2].text)
	}
}

func TestNamedRegisterAppendUppercase(t *testing.T) {
	e := newTestEditor(t, "hello")

	// Write to "a"
	e.regOverrideSet = true
	e.regOverride = 'a'
	e.writeYank(Register{kind: RegCharwise, text: "one"})

	// Append to "A" (should append to 'a')
	e.regOverrideSet = true
	e.regOverride = 'A'
	e.writeYank(Register{kind: RegCharwise, text: "two"})

	r, ok := e.getRegister('a')
	if !ok {
		t.Fatalf("expected register a present")
	}
	if r.text != "onetwo" {
		t.Fatalf("expected appended 'onetwo' got %q", r.text)
	}
}

// Test the classic Vim workflow: yy p dd P
func TestRegisterWorkflow_YY_P_DD_P(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\nline3\n")

	// Start on line2 (cy=1)
	e.cy = 1
	e.cx = 2

	// yy - yank current line
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))

	// Check register 0 (yank register) has "line2\n"
	if reg, ok := e.getRegister('0'); !ok || reg.text != "line2\n" {
		t.Errorf("yy failed: register 0 should have 'line2\\n', got %q", reg.text)
	}

	// p - paste after (should create line below)
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	// Buffer should now be: line1\nline2\nline2\nline3\n
	expected := "line1\nline2\nline2\nline3\n"
	if got := e.buffer.String(); got != expected {
		t.Errorf("After yy p, buffer mismatch.\nExpected: %q\nGot: %q", expected, got)
	}

	// Cursor should be on the pasted line (line 2, cy=2)
	if e.cy != 2 {
		t.Errorf("After p, cursor should be on line 2, got cy=%d", e.cy)
	}

	// dd - delete current line (the pasted line2)
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	// Buffer should be back to: line1\nline2\nline3\n
	expected = "line1\nline2\nline3\n"
	if got := e.buffer.String(); got != expected {
		t.Errorf("After dd, buffer mismatch.\nExpected: %q\nGot: %q", expected, got)
	}

	// Register 1 should have the deleted line
	if reg, ok := e.getRegister('1'); !ok || reg.text != "line2\n" {
		t.Errorf("dd failed: register 1 should have deleted line, got %q", reg.text)
	}

	// P - paste before current line
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'P', tcell.ModNone))

	// Buffer should be: line1\nline2\nline2\nline3\n again
	expected = "line1\nline2\nline2\nline3\n"
	if got := e.buffer.String(); got != expected {
		t.Errorf("After P, buffer mismatch.\nExpected: %q\nGot: %q", expected, got)
	}
}

// Test small delete register for x
func TestSmallDeleteRegister_X(t *testing.T) {
	e := newTestEditor(t, "hello world\n")

	e.cy = 0
	e.cx = 0

	// x - delete single character
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))

	// Small delete register "-" should have 'h'
	if reg, ok := e.getRegister('-'); !ok || reg.text != "h" {
		t.Errorf("x should use small delete register, got %q", reg.text)
	}

	// Unnamed register should also have it
	if reg, ok := e.getRegister('"'); !ok || reg.text != "h" {
		t.Errorf("unnamed register should mirror small delete, got %q", reg.text)
	}

	// Buffer should be "ello world\n"
	if got := e.buffer.String(); got != "ello world\n" {
		t.Errorf("After x, expected 'ello world\\n', got %q", got)
	}
}

// Test that small deletes don't pollute numbered registers
func TestSmallDeleteDoesNotPollute(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\nline3\n")

	// First, do a big delete (dd)
	e.cy = 0
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	// Register 1 should have "line1\n"
	reg1Before, _ := e.getRegister('1')

	// Now do a small delete (x)
	e.cx = 0
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))

	// Register 1 should still have "line1\n" (small delete doesn't shift numbered)
	reg1After, _ := e.getRegister('1')
	if reg1After.text != reg1Before.text {
		t.Errorf("Small delete polluted numbered registers: before=%q, after=%q", reg1Before.text, reg1After.text)
	}

	// Small delete register should have the 'l'
	if reg, ok := e.getRegister('-'); !ok || reg.text != "l" {
		t.Errorf("Small delete register should have 'l', got %q", reg.text)
	}
}

// Test linewise vs charwise paste cursor placement
func TestPasteCursorPlacement_Linewise(t *testing.T) {
	e := newTestEditor(t, "  line1\n  line2\n")

	// Yank line with indentation
	e.cy = 0
	e.cx = 4
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))

	// Move to next line
	e.cy = 1
	e.cx = 0

	// Paste after
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	// Cursor should be at first non-blank (column 2 where 'l' starts)
	if e.cx != 2 {
		t.Errorf("After linewise p, cursor should be at first non-blank (cx=2), got cx=%d", e.cx)
	}
}

// Test charwise paste cursor placement
func TestPasteCursorPlacement_Charwise(t *testing.T) {
	e := newTestEditor(t, "hello world\n")

	// Yank word "hello" (charwise)
	e.cy = 0
	e.cx = 0
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))

	// Move to after space
	e.cx = 6

	// Paste after
	e.handleNormal(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	// Cursor should be on last char of pasted text
	// "yw" yanks "hello " (6 chars including space)
	// Pasted after 'w' at position 6, so: "hello w" + "hello " + "orld\n"
	// Last char of pasted text is the space at position 12
	if e.cx != 12 {
		t.Errorf("After charwise p, cursor should be at end of pasted text (cx=12), got cx=%d", e.cx)
	}
}
