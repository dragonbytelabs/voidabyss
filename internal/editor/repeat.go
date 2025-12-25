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
	case RepeatInsert:
		// Replay insert mode: execute the insert command, type the text, exit
		e.repeatInsertMode()
	}
}

func (e *Editor) repeatInsertMode() {
	if e.last.insertCmd == 0 {
		return
	}

	// Execute the insert command to position cursor correctly
	switch e.last.insertCmd {
	case 'i':
		// Insert at current position - no movement needed
	case 'a':
		e.moveRight(1)
	case 'A':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx
	case 'o':
		e.openBelow()
	case 'O':
		e.openAbove()
	}

	// Start undo group for the repeated insert
	e.buffer.BeginUndoGroup()

	// Type the captured text
	for _, r := range e.last.insertText {
		if r == '\n' {
			e.newline()
		} else {
			e.insertRune(r)
		}
	}

	// End undo group
	e.buffer.EndUndoGroup()
}
