package editor

func (e *Editor) textObjectRange(prefix rune, unit rune) (start, end int, kind RegisterKind, ok bool) {
	pos := e.posFromCursor()
	r := e.textRunes()

	// Handle paired delimiter text objects: ", (, {, [
	if unit == '"' || unit == '(' || unit == ')' || unit == '{' || unit == '}' || unit == '[' || unit == ']' {
		return e.textObjectPaired(prefix, unit, pos, r)
	}

	// Handle paragraph text object
	if unit == 'p' {
		return e.textObjectParagraph(prefix, pos, r)
	}

	// Handle word text objects
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

// textObjectPaired handles paired delimiters like quotes, parens, brackets, braces
func (e *Editor) textObjectPaired(prefix rune, unit rune, pos int, r []rune) (start, end int, kind RegisterKind, ok bool) {
	// Normalize closing delimiters to opening ones
	var openCh, closeCh rune
	switch unit {
	case '"':
		openCh, closeCh = '"', '"'
	case '(', ')':
		openCh, closeCh = '(', ')'
	case '{', '}':
		openCh, closeCh = '{', '}'
	case '[', ']':
		openCh, closeCh = '[', ']'
	default:
		return pos, pos, RegCharwise, false
	}

	if pos >= len(r) {
		pos = len(r) - 1
	}
	if pos < 0 {
		return 0, 0, RegCharwise, false
	}

	// Strategy: search backward from cursor to find nearest opening delimiter,
	// then search forward from there to find its matching closing delimiter
	startPos, endPos := -1, -1

	// Search backward for opening delimiter
	// Start depth at 0, but if cursor is ON a closing delimiter, start from pos-1
	depth := 0
	searchFrom := pos

	// If we're on a closing delimiter, start searching from before it
	if pos < len(r) && r[pos] == closeCh && openCh != closeCh {
		searchFrom = pos - 1
	}

	for i := searchFrom; i >= 0; i-- {
		if r[i] == closeCh && openCh != closeCh {
			// If we hit a closing delimiter going backward, increase depth
			depth++
		} else if r[i] == openCh {
			if depth == 0 {
				// Found the matching opening delimiter
				startPos = i
				break
			}
			if openCh != closeCh {
				// This opening delimiter matches a closing one we saw
				depth--
			}
		}
	}

	// If we found an opening delimiter, search forward for its matching closing delimiter
	if startPos != -1 {
		depth = 0
		for i := startPos + 1; i < len(r); i++ {
			if r[i] == openCh && openCh != closeCh {
				depth++
			} else if r[i] == closeCh {
				if depth == 0 {
					endPos = i
					break
				}
				if openCh != closeCh {
					depth--
				}
			}
		}
	}

	// If we didn't find a pair, fail
	if startPos == -1 || endPos == -1 {
		return pos, pos, RegCharwise, false
	}

	// For 'i' (inner), exclude the delimiters
	// For 'a' (around), include the delimiters
	if prefix == 'i' {
		return startPos + 1, endPos, RegCharwise, true
	}
	return startPos, endPos + 1, RegCharwise, true
}

// textObjectParagraph handles ip/ap (paragraph text objects)
func (e *Editor) textObjectParagraph(prefix rune, pos int, r []rune) (start, end int, kind RegisterKind, ok bool) {
	if len(r) == 0 {
		return pos, pos, RegCharwise, false
	}

	// A paragraph is text surrounded by blank lines
	// Find the start of the paragraph (first non-blank line going backward)
	start = pos
	for start > 0 {
		// Check if we hit a blank line
		lineStart := start
		for lineStart > 0 && r[lineStart-1] != '\n' {
			lineStart--
		}

		// Check if this line is blank (only whitespace)
		isBlank := true
		for i := lineStart; i < len(r) && r[i] != '\n'; i++ {
			if !isSpace(r[i]) {
				isBlank = false
				break
			}
		}

		if isBlank && lineStart < pos {
			// Found blank line before cursor, paragraph starts after it
			start = lineStart
			// Skip past the newline
			if start < len(r) && r[start] == '\n' {
				start++
			}
			break
		}

		// Move to previous line
		if lineStart == 0 {
			start = 0
			break
		}
		start = lineStart - 1
	}

	// Find the end of the paragraph (first blank line going forward)
	end = pos
	for end < len(r) {
		// Find end of current line
		lineEnd := end
		for lineEnd < len(r) && r[lineEnd] != '\n' {
			lineEnd++
		}

		// Check if this line is blank
		isBlank := true
		lineStart := end
		for lineStart > 0 && r[lineStart-1] != '\n' {
			lineStart--
		}
		for i := lineStart; i < lineEnd; i++ {
			if !isSpace(r[i]) {
				isBlank = false
				break
			}
		}

		if isBlank && lineStart > pos {
			// Found blank line after cursor
			end = lineStart
			break
		}

		// Move to next line
		if lineEnd >= len(r) {
			end = len(r)
			break
		}
		end = lineEnd + 1
	}

	// For 'a' (around), include surrounding blank lines
	if prefix == 'a' {
		// Include trailing blank lines
		for end < len(r) {
			lineEnd := end
			for lineEnd < len(r) && r[lineEnd] != '\n' {
				lineEnd++
			}

			// Check if blank
			isBlank := true
			for i := end; i < lineEnd; i++ {
				if !isSpace(r[i]) {
					isBlank = false
					break
				}
			}

			if !isBlank {
				break
			}

			// Include this blank line
			if lineEnd < len(r) {
				end = lineEnd + 1
			} else {
				end = len(r)
				break
			}
		}
	}

	return start, end, RegCharwise, true
}
