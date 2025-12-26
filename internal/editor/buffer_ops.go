package editor

func (e *Editor) insertRune(r rune) {
	pos := e.posFromCursor()
	_ = e.buffer.Insert(pos, string(r))
	e.setCursorFromPos(pos + 1)
	e.wantX = e.cx
	e.dirty = true
	e.reparseBuffer()
	e.FireTextChangedI()
}

func (e *Editor) backspace() {
	pos := e.posFromCursor()
	if pos == 0 {
		return
	}
	_ = e.buffer.Delete(pos-1, pos)
	e.setCursorFromPos(pos - 1)
	e.wantX = e.cx
	e.dirty = true
	e.reparseBuffer()
	e.FireTextChangedI()
}

func (e *Editor) newline() {
	// End current undo group and start a new one
	// This makes each line a separate undo, like Vim
	if e.mode == ModeInsert {
		e.buffer.EndUndoGroup()
		e.buffer.BeginUndoGroup()
	}

	pos := e.posFromCursor()
	_ = e.buffer.Insert(pos, "\n")
	e.setCursorFromPos(pos + 1)
	e.wantX = 0
	e.dirty = true
	e.reparseBuffer()
	if e.mode == ModeInsert {
		e.FireTextChangedI()
	} else {
		e.FireTextChanged()
	}
}

func (e *Editor) deleteAtCursor() {
	pos := e.posFromCursor()
	if pos >= e.buffer.Len() {
		return
	}
	deleted, _ := e.buffer.Slice(pos, pos+1)
	// Single char deletes go to small delete register "-" unless explicitly overridden
	if !e.regOverrideSet {
		e.regs.small = Register{kind: RegCharwise, text: deleted}
		// Unnamed register gets it too (Vim behavior)
		e.regs.unnamed = e.regs.small
	} else {
		// User specified a register explicitly
		e.writeDelete(Register{kind: RegCharwise, text: deleted})
	}
	_ = e.buffer.Delete(pos, pos+1)
	e.setCursorFromPos(pos)
	e.dirty = true
	e.reparseBuffer()
	e.FireTextChanged()
	e.statusMsg = "deleted"
}

func (e *Editor) undo() {
	if e.buffer == nil {
		return
	}
	if ok := e.buffer.Undo(); !ok {
		e.statusMsg = "Already at oldest change"
		return
	}
	pos := min(e.posFromCursor(), e.buffer.Len())
	e.setCursorFromPos(pos)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "undo"
	e.clearPending()
	e.reparseBuffer()
	e.FireTextChanged()
}

func (e *Editor) redo() {
	if e.buffer == nil {
		return
	}
	if ok := e.buffer.Redo(); !ok {
		e.statusMsg = "Already at newest change"
		return
	}
	pos := min(e.posFromCursor(), e.buffer.Len())
	e.setCursorFromPos(pos)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "redo"
	e.clearPending()
	e.reparseBuffer()
	e.FireTextChanged()
}

/* linewise */
func (e *Editor) deleteLines(n int) {
	if n <= 0 {
		return
	}
	starts := e.lineStarts()
	if len(starts) == 0 {
		return
	}

	startLine := e.cy
	endLine := min(e.lineCount(), e.cy+n)

	startPos := starts[startLine]
	endPos := e.buffer.Len()
	if endLine < e.lineCount() {
		endPos = starts[endLine]
	}

	deleted, _ := e.buffer.Slice(startPos, endPos)
	e.writeDelete(Register{kind: RegLinewise, text: deleted})
	_ = e.buffer.Delete(startPos, endPos)

	e.setCursorFromPos(startPos)
	e.wantX = 0
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankLines(n int) {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return
	}

	startLine := e.cy
	endLine := min(e.lineCount(), e.cy+n)

	startPos := starts[startLine]
	endPos := e.buffer.Len()
	if endLine < e.lineCount() {
		endPos = starts[endLine]
	}

	s, _ := e.buffer.Slice(startPos, endPos)
	e.writeYank(Register{kind: RegLinewise, text: s})
	e.statusMsg = "yanked"
}

/* bol/eol */
func (e *Editor) deleteToBOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	if pos > lineStart {
		deleted, _ := e.buffer.Slice(lineStart, pos)
		e.writeDelete(Register{kind: RegCharwise, text: deleted})
		_ = e.buffer.Delete(lineStart, pos)
		e.setCursorFromPos(lineStart)
		e.wantX = 0
		e.dirty = true
		e.statusMsg = "deleted"
	}
}

func (e *Editor) deleteToEOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	eol := lineStart + e.lineLen(e.cy)
	if pos < eol {
		deleted, _ := e.buffer.Slice(pos, eol)
		e.writeDelete(Register{kind: RegCharwise, text: deleted})
		_ = e.buffer.Delete(pos, eol)
		e.setCursorFromPos(pos)
		e.dirty = true
		e.statusMsg = "deleted"
	}
}

func (e *Editor) yankToBOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	s, _ := e.buffer.Slice(lineStart, pos)
	e.writeYank(Register{kind: RegCharwise, text: s})
	e.statusMsg = "yanked"
}

func (e *Editor) yankToEOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	eol := lineStart + e.lineLen(e.cy)
	s, _ := e.buffer.Slice(pos, eol)
	e.writeYank(Register{kind: RegCharwise, text: s})
	e.statusMsg = "yanked"
}

func (e *Editor) deleteByWordMotion(count int, big bool) {
	start := e.posFromCursor()
	end := e.nextWordStart(start, count, big)

	if end <= start {
		e.statusMsg = "nothing to delete"
		return
	}

	deleted, _ := e.buffer.Slice(start, end)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(start, end)

	e.setCursorFromPos(start)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankByWordMotion(count int, big bool) {
	start := e.posFromCursor()
	end := e.nextWordStart(start, count, big)

	if end <= start {
		e.statusMsg = "nothing to yank"
		return
	}

	yanked, _ := e.buffer.Slice(start, end)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}

func (e *Editor) deleteByBackWordMotion(count int, big bool) {
	end := e.posFromCursor()
	start := e.prevWordStart(end, count, big)

	if end <= start {
		e.statusMsg = "nothing to delete"
		return
	}

	deleted, _ := e.buffer.Slice(start, end)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(start, end)

	e.setCursorFromPos(start)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankByBackWordMotion(count int, big bool) {
	end := e.posFromCursor()
	start := e.prevWordStart(end, count, big)

	if end <= start {
		e.statusMsg = "nothing to yank"
		return
	}

	yanked, _ := e.buffer.Slice(start, end)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}

func (e *Editor) deleteByEndWordMotion(count int, big bool) {
	start := e.posFromCursor()
	if e.buffer.Len() == 0 {
		e.statusMsg = "nothing to delete"
		return
	}

	endIncl := e.endOfWord(start, count, big)
	end := endIncl + 1 // convert inclusive end -> slicing end

	if end <= start {
		e.statusMsg = "nothing to delete"
		return
	}

	deleted, _ := e.buffer.Slice(start, end)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(start, end)

	e.setCursorFromPos(start)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankByEndWordMotion(count int, big bool) {
	start := e.posFromCursor()
	if e.buffer.Len() == 0 {
		e.statusMsg = "nothing to yank"
		return
	}

	endIncl := e.endOfWord(start, count, big)
	end := endIncl + 1

	if end <= start {
		e.statusMsg = "nothing to yank"
		return
	}

	yanked, _ := e.buffer.Slice(start, end)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}
