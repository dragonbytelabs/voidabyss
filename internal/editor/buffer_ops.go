package editor

func (e *Editor) insertRune(r rune) {
	pos := e.posFromCursor()
	_ = e.buffer.Insert(pos, string(r))
	e.setCursorFromPos(pos + 1)
	e.wantX = e.cx
	e.dirty = true
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
}

func (e *Editor) newline() {
	pos := e.posFromCursor()
	_ = e.buffer.Insert(pos, "\n")
	e.setCursorFromPos(pos + 1)
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) deleteAtCursor() {
	pos := e.posFromCursor()
	if pos >= e.buffer.Len() {
		return
	}
	deleted, _ := e.buffer.Slice(pos, pos+1)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(pos, pos+1)
	e.setCursorFromPos(pos)
	e.dirty = true
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

	// remove trailing newline for yank if present
	if endPos > startPos {
		lastChar, _ := e.buffer.Slice(endPos-1, endPos)
		if lastChar == "\n" {
			endPos--
		}
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

/* word-delete/yank by motions (w/b/e + big variants) */
func (e *Editor) deleteByWordMotion(count int, big bool) {
	pos := e.posFromCursor()
	r := e.textRunes()
	end := pos
	for i := 0; i < max(1, count); i++ {
		end = wordForwardStart(r, end, big)
	}
	if end <= pos {
		e.statusMsg = "nothing to delete"
		return
	}
	deleted, _ := e.buffer.Slice(pos, end)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(pos, end)
	e.setCursorFromPos(pos)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankByWordMotion(count int, big bool) {
	pos := e.posFromCursor()
	r := e.textRunes()
	end := pos
	for i := 0; i < max(1, count); i++ {
		end = wordForwardStart(r, end, big)
	}
	if end <= pos {
		e.statusMsg = "nothing to yank"
		return
	}
	yanked, _ := e.buffer.Slice(pos, end)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}

func (e *Editor) deleteByBackWordMotion(count int, big bool) {
	pos := e.posFromCursor()
	r := e.textRunes()
	start := pos
	for i := 0; i < max(1, count); i++ {
		start = wordBackStart(r, start, big)
	}
	if start >= pos {
		e.statusMsg = "nothing to delete"
		return
	}
	deleted, _ := e.buffer.Slice(start, pos)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(start, pos)
	e.setCursorFromPos(start)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankByBackWordMotion(count int, big bool) {
	pos := e.posFromCursor()
	r := e.textRunes()
	start := pos
	for i := 0; i < max(1, count); i++ {
		start = wordBackStart(r, start, big)
	}
	if start >= pos {
		e.statusMsg = "nothing to yank"
		return
	}
	yanked, _ := e.buffer.Slice(start, pos)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}

func (e *Editor) deleteByEndWordMotion(count int, big bool) {
	pos := e.posFromCursor()
	r := e.textRunes()
	endPos := pos
	for i := 0; i < max(1, count); i++ {
		endPos = wordEnd(r, endPos, big)
		// make end exclusive
		if endPos < len(r) {
			endPos++
		}
	}
	if endPos <= pos {
		e.statusMsg = "nothing to delete"
		return
	}
	deleted, _ := e.buffer.Slice(pos, endPos)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(pos, endPos)
	e.setCursorFromPos(pos)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankByEndWordMotion(count int, big bool) {
	pos := e.posFromCursor()
	r := e.textRunes()
	endPos := pos
	for i := 0; i < max(1, count); i++ {
		endPos = wordEnd(r, endPos, big)
		if endPos < len(r) {
			endPos++
		}
	}
	if endPos <= pos {
		e.statusMsg = "nothing to yank"
		return
	}
	yanked, _ := e.buffer.Slice(pos, endPos)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}
