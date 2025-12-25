package editor

import "fmt"

func (e *Editor) applyOperatorMotion(op rune, motion rune, count int) {
	// record repeat (operator/motion)
	e.last = RepeatAction{
		kind:   RepeatOpMotion,
		op:     op,
		motion: motion,
		count:  count,
		reg:    e.regOverrideIfAny(), // capture explicit reg for repeat
	}

	// linewise dd/cc/yy and gg
	if motion == op {
		switch op {
		case 'd':
			e.deleteLines(count)
			return
		case 'c':
			e.deleteLines(count)
			e.mode = ModeInsert
			return
		case 'y':
			e.yankLines(count)
			return
		case 'g': // gg
			e.cy, e.cx, e.wantX = 0, 0, 0
			return
		}
	}

	switch motion {
	case '0':
		if op == 'd' || op == 'c' {
			e.deleteToBOL()
			if op == 'c' {
				e.mode = ModeInsert
			}
			return
		}
		if op == 'y' {
			e.yankToBOL()
			return
		}
	case '$':
		if op == 'd' || op == 'c' {
			e.deleteToEOL()
			if op == 'c' {
				e.mode = ModeInsert
			}
			return
		}
		if op == 'y' {
			e.yankToEOL()
			return
		}
	case 'w', 'W':
		if op == 'd' || op == 'c' {
			e.deleteByWordMotion(count, motion == 'W')
			if op == 'c' {
				e.mode = ModeInsert
			}
			return
		}
		if op == 'y' {
			e.yankByWordMotion(count, motion == 'W')
			return
		}
	case 'b', 'B':
		if op == 'd' || op == 'c' {
			e.deleteByBackWordMotion(count, motion == 'B')
			if op == 'c' {
				e.mode = ModeInsert
			}
			return
		}
		if op == 'y' {
			e.yankByBackWordMotion(count, motion == 'B')
			return
		}
	case 'e', 'E':
		if op == 'd' || op == 'c' {
			e.deleteByEndWordMotion(count, motion == 'E')
			if op == 'c' {
				e.mode = ModeInsert
			}
			return
		}
		if op == 'y' {
			e.yankByEndWordMotion(count, motion == 'E')
			return
		}
	}

	e.statusMsg = fmt.Sprintf("Unsupported motion %q for %q (for now)", motion, op)
}

func (e *Editor) applyOperatorTextObject(op rune, prefix rune, unit rune, count int) {
	// record repeat for textobj
	e.last = RepeatAction{
		kind:          RepeatOpMotion,
		op:            op,
		count:         count,
		reg:           e.regOverrideIfAny(),
		textObjPrefix: prefix,
		textObjUnit:   unit,
	}

	// apply count by repeating range expansion (simple: apply op count times from cursor)
	for i := 0; i < max(1, count); i++ {
		start, end, _, ok := e.textObjectRange(prefix, unit)
		if !ok || end <= start {
			e.statusMsg = "nothing"
			return
		}
		switch op {
		case 'd':
			deleted, _ := e.buffer.Slice(start, end)
			e.writeDelete(Register{kind: RegCharwise, text: deleted})
			_ = e.buffer.Delete(start, end)
			e.setCursorFromPos(start)
			e.wantX = e.cx
			e.dirty = true
			e.statusMsg = "deleted"
		case 'y':
			yanked, _ := e.buffer.Slice(start, end)
			e.writeYank(Register{kind: RegCharwise, text: yanked})
			e.statusMsg = "yanked"
		case 'c':
			deleted, _ := e.buffer.Slice(start, end)
			e.writeDelete(Register{kind: RegCharwise, text: deleted})
			_ = e.buffer.Delete(start, end)
			e.setCursorFromPos(start)
			e.wantX = e.cx
			e.dirty = true
			e.statusMsg = "deleted"
			e.mode = ModeInsert
		}
	}
}

func (e *Editor) regOverrideIfAny() rune {
	if e.regOverrideSet {
		return e.regOverride
	}
	return 0
}
