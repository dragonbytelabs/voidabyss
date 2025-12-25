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

// nextWordStart returns the position of the start of the next word/WORD after `pos`,
// repeated `count` times, following Vim-ish semantics for operator+`w/W`.
func (e *Editor) nextWordStart(pos int, count int, big bool) int {
	if count <= 0 {
		count = 1
	}

	r := e.textRunes()
	n := len(r)
	if pos < 0 {
		pos = 0
	}
	if pos > n {
		pos = n
	}

	isWord := isWordCharSmall
	if big {
		isWord = isWordCharBig
	}

	i := pos
	for c := 0; c < count; c++ {
		if i >= n {
			return n
		}

		// If currently in a word, consume to end of this word
		if i < n && isWord(r[i]) {
			for i < n && isWord(r[i]) {
				i++
			}
		}

		// Then skip non-word until next word start (this is what makes dw eat spaces)
		for i < n && !isWord(r[i]) {
			i++
		}
	}

	return i
}

// prevWordStart returns the start of the previous word/WORD before `pos`,
// repeated `count` times (Vim-ish for b/B).
func (e *Editor) prevWordStart(pos int, count int, big bool) int {
	if count <= 0 {
		count = 1
	}

	r := e.textRunes()
	n := len(r)
	if pos < 0 {
		pos = 0
	}
	if pos > n {
		pos = n
	}

	isWord := isWordCharSmall
	if big {
		isWord = isWordCharBig
	}

	i := pos
	for c := 0; c < count; c++ {
		if i <= 0 {
			return 0
		}

		// Step back one to look "before" the cursor position
		i--

		// Skip any non-word backwards
		for i >= 0 && !isWord(r[i]) {
			i--
		}
		if i < 0 {
			return 0
		}

		// Now we're in a word; move to its start
		for i >= 0 && isWord(r[i]) {
			i--
		}
		i++ // overshot by one
	}

	return i
}

// endOfWord returns the position of the end of the current/next word/WORD
// (inclusive index) repeated `count` times (Vim-ish for e/E).
func (e *Editor) endOfWord(pos int, count int, big bool) int {
	if count <= 0 {
		count = 1
	}

	r := e.textRunes()
	n := len(r)
	if pos < 0 {
		pos = 0
	}
	if pos >= n {
		return n - 1
	}

	isWord := isWordCharSmall
	if big {
		isWord = isWordCharBig
	}

	i := pos
	end := pos

	for c := 0; c < count; c++ {
		// If we're not on a word, skip forward to next word
		for i < n && !isWord(r[i]) {
			i++
		}
		if i >= n {
			return n - 1
		}

		// Now consume word; end becomes last char of it
		for i < n && isWord(r[i]) {
			end = i
			i++
		}
	}

	return end
}

// moveParagraphBackward moves cursor to the beginning of the previous paragraph
func (e *Editor) moveParagraphBackward(count int) {
	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		// Move up one line to start
		if e.cy <= 0 {
			e.cy = 0
			e.cx = 0
			e.wantX = 0
			return
		}
		e.cy--

		// Skip current blank lines
		for e.cy > 0 && e.isBlankLine(e.cy) {
			e.cy--
		}

		// Now skip non-blank lines to find the blank separator
		for e.cy > 0 && !e.isBlankLine(e.cy) {
			e.cy--
		}

		// If we're on a blank line and not at top, move to first non-blank
		if e.cy > 0 {
			e.cy++
		}
	}

	e.cx = 0
	e.wantX = 0
}

// moveParagraphForward moves cursor to the beginning of the next paragraph
func (e *Editor) moveParagraphForward(count int) {
	if count <= 0 {
		count = 1
	}

	maxLine := e.lineCount() - 1

	for i := 0; i < count; i++ {
		// Skip current non-blank lines
		for e.cy < maxLine && !e.isBlankLine(e.cy) {
			e.cy++
		}

		// Skip blank lines
		for e.cy < maxLine && e.isBlankLine(e.cy) {
			e.cy++
		}
	}

	e.cx = 0
	e.wantX = 0
}

// isBlankLine returns true if the line at y is empty or contains only whitespace
func (e *Editor) isBlankLine(y int) bool {
	line := e.getLine(y)
	for _, ch := range line {
		if ch != ' ' && ch != '\t' {
			return false
		}
	}
	return true
}
