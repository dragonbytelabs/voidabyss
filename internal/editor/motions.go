package editor

func (e *Editor) moveUp(n int) {
	e.cy = max(0, e.cy-n)
	e.cx = min(e.wantX, e.lineLen(e.cy))
}

func (e *Editor) moveDown(n int) {
	e.cy = min(e.lineCount()-1, e.cy+n)
	e.cx = min(e.wantX, e.lineLen(e.cy))
}

func (e *Editor) moveLeft(n int) {
	e.cx = max(0, e.cx-n)
	e.wantX = e.cx
}

func (e *Editor) moveRight(n int) {
	e.cx = min(e.lineLen(e.cy), e.cx+n)
	e.wantX = e.cx
}

func (e *Editor) openBelow() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := lineStart + e.lineLen(e.cy)
	_ = e.buffer.Insert(pos, "\n")
	e.setCursorFromPos(pos + 1)
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) openAbove() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	_ = e.buffer.Insert(lineStart, "\n")
	e.setCursorFromPos(lineStart)
	e.wantX = 0
	e.dirty = true
}

/* word motions */

func (e *Editor) moveWordForward(count int, big bool) {
	if count <= 0 {
		count = 1
	}
	pos := e.posFromCursor()
	r := e.textRunes()

	for i := 0; i < count; i++ {
		pos = wordForwardStart(r, pos, big)
	}
	e.setCursorFromPos(pos)
	e.wantX = e.cx
}

func (e *Editor) moveWordBack(count int, big bool) {
	if count <= 0 {
		count = 1
	}
	pos := e.posFromCursor()
	r := e.textRunes()

	for i := 0; i < count; i++ {
		pos = wordBackStart(r, pos, big)
	}
	e.setCursorFromPos(pos)
	e.wantX = e.cx
}

func (e *Editor) moveWordEnd(count int, big bool) {
	if count <= 0 {
		count = 1
	}
	pos := e.posFromCursor()
	r := e.textRunes()

	for i := 0; i < count; i++ {
		pos = wordEnd(r, pos, big)
	}
	e.setCursorFromPos(pos)
	e.wantX = e.cx
}

func wordForwardStart(r []rune, pos int, big bool) int {
	if pos < 0 {
		pos = 0
	}
	if pos >= len(r) {
		return len(r)
	}
	isUnit := isWordChar
	if big {
		isUnit = isWORDChar
	}

	// If currently on unit, skip units
	for pos < len(r) && isUnit(r[pos]) {
		pos++
	}
	// Skip non-unit (whitespace/punct) to next unit
	for pos < len(r) && !isUnit(r[pos]) {
		pos++
	}
	return pos
}

func wordBackStart(r []rune, pos int, big bool) int {
	if pos <= 0 {
		return 0
	}
	if pos > len(r) {
		pos = len(r)
	}
	isUnit := isWordChar
	if big {
		isUnit = isWORDChar
	}

	// step left once if possible
	pos--

	// skip non-unit backwards
	for pos > 0 && !isUnit(r[pos]) {
		pos--
	}
	// now skip unit backwards to its start
	for pos > 0 && isUnit(r[pos-1]) {
		pos--
	}
	return pos
}

func wordEnd(r []rune, pos int, big bool) int {
	if pos < 0 {
		pos = 0
	}
	if pos >= len(r) {
		return len(r)
	}
	isUnit := isWordChar
	if big {
		isUnit = isWORDChar
	}

	// if not on unit, move to next unit
	for pos < len(r) && !isUnit(r[pos]) {
		pos++
	}
	if pos >= len(r) {
		return len(r)
	}
	// move to end of unit (last char)
	for pos+1 < len(r) && isUnit(r[pos+1]) {
		pos++
	}
	return pos
}
