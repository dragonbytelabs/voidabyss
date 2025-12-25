package editor

func (e *Editor) visualEnter(kind VisualKind) {
	e.visualActive = true
	e.visualKind = kind
	e.visualAnchor = e.posFromCursor()
	e.mode = ModeVisual
}

func (e *Editor) visualExit() {
	e.visualActive = false
	e.mode = ModeNormal
	e.statusMsg = ""
}

func (e *Editor) visualRange() (start, end int, kind RegisterKind) {
	a := e.visualAnchor
	b := e.posFromCursor()
	if a > b {
		a, b = b, a
	}

	if e.visualKind == VisualLine {
		// expand to full lines, end exclusive includes newline if present
		startLine := e.lineIndexForPos(a)
		endLine := e.lineIndexForPos(b)
		start = e.lineStartPos(startLine)

		// endLine inclusive -> end at start of next line (or EOF)
		next := endLine + 1
		if next >= e.lineCount() {
			end = e.buffer.Len()
		} else {
			end = e.lineStartPos(next)
		}
		return start, end, RegLinewise
	}

	// Charwise: include the character under the cursor (end is exclusive, so +1)
	return a, b + 1, RegCharwise
}

func (e *Editor) lineIndexForPos(pos int) int {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return 0
	}
	pos = clamp(pos, 0, e.buffer.Len())
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
	return best
}
