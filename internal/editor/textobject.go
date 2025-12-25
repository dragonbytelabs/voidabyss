package editor

func (e *Editor) textObjectRange(prefix rune, unit rune) (start, end int, kind RegisterKind, ok bool) {
	pos := e.posFromCursor()
	r := e.textRunes()

	big := unit == 'W'
	isUnit := isWordChar
	if big {
		isUnit = isWORDChar
	}

	// Find a unit near cursor: if cursor on non-unit, search right then left a bit
	p := pos
	if p >= len(r) {
		p = len(r) - 1
	}
	if p < 0 {
		return pos, pos, RegCharwise, false
	}

	if !isUnit(r[p]) {
		// search right
		q := p
		for q < len(r) && !isUnit(r[q]) {
			q++
		}
		if q < len(r) {
			p = q
		} else {
			// search left
			q = p
			for q >= 0 && !isUnit(r[q]) {
				q--
			}
			if q >= 0 {
				p = q
			} else {
				return pos, pos, RegCharwise, false
			}
		}
	}

	// expand to unit bounds
	s := p
	for s > 0 && isUnit(r[s-1]) {
		s--
	}
	ei := p
	for ei+1 < len(r) && isUnit(r[ei+1]) {
		ei++
	}
	selStart, selEnd := s, ei+1 // end exclusive

	if prefix == 'a' {
		// include surrounding whitespace: prefer trailing whitespace, else leading
		t := selEnd
		for t < len(r) && isSpace(r[t]) && r[t] != '\n' {
			t++
		}
		if t != selEnd {
			selEnd = t
		} else {
			// include leading spaces (not newline)
			ls := selStart
			for ls > 0 && isSpace(r[ls-1]) && r[ls-1] != '\n' {
				ls--
			}
			selStart = ls
		}
	}

	return selStart, selEnd, RegCharwise, true
}
