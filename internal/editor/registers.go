package editor

import "fmt"

func isRegisterName(r rune) bool {
	if r == '"' || r == '-' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	return false
}

func (e *Editor) setRegister(name rune, r Register) {
	if name == 0 {
		name = '"'
	}

	// Append behavior for A-Z
	if name >= 'A' && name <= 'Z' {
		lower := name + ('a' - 'A')
		prev, _ := e.getRegister(lower)

		kind := r.kind
		if prev.text != "" && prev.kind == r.kind {
			kind = prev.kind
		}
		e.regs.named[lower] = Register{kind: kind, text: prev.text + r.text}
		e.regs.unnamed = e.regs.named[lower]
		return
	}

	switch {
	case name == '"':
		e.regs.unnamed = r
	case name == '-':
		e.regs.small = r
	case name >= '0' && name <= '9':
		e.regs.numbered[name-'0'] = r
	case name >= 'a' && name <= 'z':
		e.regs.named[name] = r
	}
}

func (e *Editor) getRegister(name rune) (Register, bool) {
	if name == 0 {
		name = '"'
	}
	if name >= 'A' && name <= 'Z' {
		name = name + ('a' - 'A')
	}

	switch {
	case name == '"':
		return e.regs.unnamed, e.regs.unnamed.text != ""
	case name == '-':
		return e.regs.small, e.regs.small.text != ""
	case name >= '0' && name <= '9':
		r := e.regs.numbered[name-'0']
		return r, r.text != ""
	case name >= 'a' && name <= 'z':
		r, ok := e.regs.named[name]
		return r, ok && r.text != ""
	default:
		return Register{}, false
	}
}

func (e *Editor) consumeRegister() rune {
	if e.regOverrideSet {
		r := e.regOverride
		e.regOverrideSet = false
		e.regOverride = 0
		return r
	}
	return '"'
}

func (e *Editor) ensureRegsInit() {
	if e.regs.named == nil {
		e.regs.named = make(map[rune]Register)
	}
}

func (e *Editor) writeDelete(r Register) {
	e.ensureRegsInit()

	target := e.consumeRegister()
	e.setRegister(target, r)

	// shift 9..2
	for i := 9; i >= 2; i-- {
		e.regs.numbered[i] = e.regs.numbered[i-1]
	}
	e.regs.numbered[1] = r

	// unnamed mirrors last delete too
	if stored, ok := e.getRegister(target); ok && stored.text != "" {
		e.regs.unnamed = stored
	} else {
		e.regs.unnamed = r
	}
}

func (e *Editor) writeYank(r Register) {
	e.ensureRegsInit()

	target := e.consumeRegister()
	e.setRegister(target, r)

	// yank register 0 gets yanks
	e.regs.numbered[0] = r

	if stored, ok := e.getRegister(target); ok && stored.text != "" {
		e.regs.unnamed = stored
	} else {
		e.regs.unnamed = r
	}
}

func (e *Editor) readPaste() (Register, bool) {
	name := e.consumeRegister()
	return e.getRegister(name)
}

// popup register formatting
func (e *Editor) formatRegisters() []string {
	lines := make([]string, 0, 40)

	lines = append(lines, fmt.Sprintf("\"  %s", e.formatRegValue(e.regs.unnamed)))
	if e.regs.small.text != "" {
		lines = append(lines, fmt.Sprintf("-  %s", e.formatRegValue(e.regs.small)))
	}

	for i := 0; i <= 9; i++ {
		name := rune('0' + i)
		r := e.regs.numbered[i]
		if r.text == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%c  %s", name, e.formatRegValue(r)))
	}

	for ch := 'a'; ch <= 'z'; ch++ {
		r, ok := e.regs.named[ch]
		if !ok || r.text == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%c  %s", ch, e.formatRegValue(r)))
	}

	if len(lines) == 0 {
		return []string{"(no registers)"}
	}
	return lines
}

func (e *Editor) formatRegValue(r Register) string {
	if r.text == "" {
		return "(empty)"
	}
	kind := "char"
	switch r.kind {
	case RegLinewise:
		kind = "line"
	case RegBlockwise:
		kind = "block"
	}
	return fmt.Sprintf("[%s] %s", kind, e.previewText(r.text, 60))
}
