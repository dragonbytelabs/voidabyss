package editor

// NOTE: this recomputes a lot by calling buffer.String(). It's fine for now.

func (e *Editor) textRunes() []rune { return []rune(e.buffer.String()) }

func (e *Editor) lineStarts() []int {
	r := e.textRunes()
	starts := []int{0}
	for i, ch := range r {
		if ch == '\n' {
			if i+1 <= len(r) {
				starts = append(starts, i+1)
			}
		}
	}
	return starts
}

func (e *Editor) lineCount() int {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return 1
	}
	return len(starts)
}

func (e *Editor) getLine(y int) string {
	r := e.textRunes()
	starts := e.lineStarts()
	if len(starts) == 0 {
		return ""
	}
	y = clamp(y, 0, len(starts)-1)

	start := starts[y]
	end := len(r)
	if y+1 < len(starts) {
		end = starts[y+1] - 1
		if end < start {
			end = start
		}
	}
	return string(r[start:end])
}

func (e *Editor) lineLen(y int) int {
	return len([]rune(e.getLine(y)))
}

func (e *Editor) posFromCursor() int {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return 0
	}
	e.cy = clamp(e.cy, 0, len(starts)-1)

	lineStart := starts[e.cy]
	ll := e.lineLen(e.cy)
	e.cx = clamp(e.cx, 0, ll)
	return lineStart + e.cx
}

func (e *Editor) setCursorFromPos(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > e.buffer.Len() {
		pos = e.buffer.Len()
	}

	starts := e.lineStarts()
	if len(starts) == 0 {
		e.cy, e.cx = 0, 0
		return
	}

	lo, hi := 0, len(starts)-1
	best := 0
	for lo <= hi {
		m := (lo + hi) / 2
		if starts[m] <= pos {
			best = m
			lo = m + 1
		} else {
			hi = m - 1
		}
	}

	e.cy = best
	e.cx = pos - starts[best]
	ll := e.lineLen(e.cy)
	if e.cx > ll {
		e.cx = ll
	}
}

func (e *Editor) ensureCursorValid() {
	lc := e.lineCount()
	if lc <= 0 {
		lc = 1
	}
	e.cy = clamp(e.cy, 0, lc-1)
	e.cx = clamp(e.cx, 0, e.lineLen(e.cy))
}

func (e *Editor) ensureCursorVisible() {
	w, h := e.s.Size()
	viewH := max(1, h-1)

	if e.cy < e.rowOffset {
		e.rowOffset = e.cy
	} else if e.cy >= e.rowOffset+viewH {
		e.rowOffset = e.cy - viewH + 1
	}

	if e.cx < e.colOffset {
		e.colOffset = e.cx
	} else if e.cx >= e.colOffset+w {
		e.colOffset = e.cx - w + 1
	}

	if e.rowOffset < 0 {
		e.rowOffset = 0
	}
	if e.colOffset < 0 {
		e.colOffset = 0
	}
}

func (e *Editor) lineStartPos(y int) int {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return 0
	}
	y = clamp(y, 0, len(starts)-1)
	return starts[y]
}
