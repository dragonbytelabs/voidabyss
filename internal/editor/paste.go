package editor

func (e *Editor) pasteAfter() {
	reg, ok := e.readPaste()
	if !ok || reg.text == "" {
		e.statusMsg = "nothing to paste"
		return
	}

	switch reg.kind {
	case RegLinewise:
		// Insert at start of NEXT line if it exists; otherwise at EOF (with newline if needed).
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

		_ = e.buffer.Insert(insertPos, reg.text+"\n")
		e.setCursorFromPos(insertPos)
		e.wantX = 0

	default: // charwise
		pos := e.posFromCursor()
		insertPos := pos
		if insertPos < e.buffer.Len() {
			insertPos = pos + 1
		}
		_ = e.buffer.Insert(insertPos, reg.text)
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
		_ = e.buffer.Insert(insertPos, reg.text+"\n")
		e.setCursorFromPos(insertPos)
		e.wantX = 0
	default:
		pos := e.posFromCursor()
		_ = e.buffer.Insert(pos, reg.text)
		e.setCursorFromPos(pos)
		e.wantX = e.cx
	}
}
