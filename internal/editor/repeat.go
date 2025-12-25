package editor

func (e *Editor) repeatLast() {
	switch e.last.kind {
	case RepeatOpMotion:
		// text object repeat
		if e.last.textObjPrefix != 0 && (e.last.textObjUnit == 'w' || e.last.textObjUnit == 'W') {
			// restore explicit register (if any)
			if e.last.reg != 0 {
				e.regOverrideSet = true
				e.regOverride = e.last.reg
			}
			e.applyOperatorTextObject(e.last.op, e.last.textObjPrefix, e.last.textObjUnit, e.last.count)
			return
		}

		if e.last.reg != 0 {
			e.regOverrideSet = true
			e.regOverride = e.last.reg
		}
		e.applyOperatorMotion(e.last.op, e.last.motion, e.last.count)
	case RepeatPasteAfter:
		e.pasteAfter()
	case RepeatPasteBefore:
		e.pasteBefore(e.last.count)
	case RepeatDeleteChar:
		for i := 0; i < max(1, e.last.count); i++ {
			e.deleteAtCursor()
		}
	}
}
