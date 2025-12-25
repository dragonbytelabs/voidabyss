package editor

func (e *Editor) pasteAfter() {
	reg, ok := e.readPaste()
	if !ok || reg.text == "" {
		e.statusMsg = "nothing to paste"
		return
	}

	switch reg.kind {
	case RegLinewise:
		// Paste below current line
		insertPos := e.buffer.Len()

		if e.cy+1 < e.lineCount() {
			insertPos = e.lineStartPos(e.cy + 1)
		} else {
			// at last line: ensure file ends with newline if non-empty and missing it
			if e.buffer.Len() > 0 {
				last, _ := e.buffer.Slice(e.buffer.Len()-1, e.buffer.Len())
				if last != "\n" {
					_ = e.buffer.Insert(e.buffer.Len(), "\n")
				}
			}
			insertPos = e.buffer.Len()
		}

		_ = e.buffer.Insert(insertPos, reg.text)
		// Cursor goes to first non-blank character of pasted line
		e.setCursorFromPos(insertPos)
		e.moveToFirstNonBlank()

	default: // charwise
		pos := e.posFromCursor()
		insertPos := pos
		if insertPos < e.buffer.Len() {
			insertPos = pos + 1
		}
		_ = e.buffer.Insert(insertPos, reg.text)
		// Cursor goes to last character of pasted text
		e.setCursorFromPos(insertPos + len([]rune(reg.text)) - 1)
		e.wantX = e.cx
	}

	e.dirty = true
	e.statusMsg = "pasted"
}

func (e *Editor) pasteBefore(n int) {
	reg, ok := e.readPaste()
	if !ok || reg.text == "" {
		e.statusMsg = "nothing to paste"
		return
	}
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		e.pasteBeforeOnce(reg)
	}
	e.dirty = true
	e.statusMsg = "pasted"
}

func (e *Editor) pasteBeforeOnce(reg Register) {
	switch reg.kind {
	case RegLinewise:
		insertPos := e.lineStartPos(e.cy)
		_ = e.buffer.Insert(insertPos, reg.text)
		// Cursor stays at first non-blank of pasted line
		e.setCursorFromPos(insertPos)
		e.moveToFirstNonBlank()
	default:
		pos := e.posFromCursor()
		_ = e.buffer.Insert(pos, reg.text)
		// Cursor goes to last character of pasted text
		e.setCursorFromPos(pos + len([]rune(reg.text)) - 1)
		e.wantX = e.cx
	}
}

func (e *Editor) moveToFirstNonBlank() {
	line := e.getLine(e.cy)
	pos := 0
	for i, ch := range []rune(line) {
		if ch != ' ' && ch != '\t' {
			pos = i
			break
		}
	}
	e.cx = pos
	e.wantX = e.cx
}
