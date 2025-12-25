package editor

import "testing"

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
