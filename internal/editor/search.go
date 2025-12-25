package editor

import "strings"

// searchNext finds the next occurrence of searchQuery
// skipCurrent: if true, skip a match at current position (for 'n'/'N')
func (e *Editor) searchNext(forward bool, skipCurrent bool) {
	if e.searchQuery == "" {
		return
	}

	text := e.buffer.String()
	query := e.searchQuery

	// Start position: current cursor position
	currentPos := e.posFromCursor()

	var matchPos int
	found := false

	if forward {
		// Search forward
		startPos := currentPos
		if skipCurrent {
			startPos = currentPos + 1
		}

		if startPos <= len(text) {
			idx := strings.Index(text[startPos:], query)
			if idx != -1 {
				matchPos = startPos + idx
				found = true
			}
		}

		// Wrap around to beginning if not found
		if !found {
			idx := strings.Index(text, query)
			if idx != -1 && idx < currentPos {
				matchPos = idx
				found = true
				e.statusMsg = "search wrapped"
			}
		}
	} else {
		// Search backward
		searchEnd := currentPos
		if skipCurrent {
			searchEnd = currentPos
		} else {
			// Include current position for initial backward search
			searchEnd = currentPos + 1
		}

		if searchEnd > 0 {
			idx := strings.LastIndex(text[:searchEnd], query)
			if idx != -1 {
				matchPos = idx
				found = true
			}
		}

		// Wrap around to end if not found
		if !found {
			idx := strings.LastIndex(text, query)
			if idx != -1 && idx > currentPos {
				found = true
				e.statusMsg = "search wrapped"
			}
		}
	}

	if found {
		e.setCursorFromPos(matchPos)
		e.wantX = e.cx
		e.updateSearchHighlights()
	} else {
		e.statusMsg = "pattern not found: " + query
		e.searchMatches = nil
	}
}

// updateSearchHighlights finds all matches in the visible viewport
func (e *Editor) updateSearchHighlights() {
	if e.searchQuery == "" {
		e.searchMatches = nil
		return
	}

	text := e.buffer.String()
	query := e.searchQuery
	e.searchMatches = nil

	// Find all matches in the entire text (simple implementation)
	// Only highlight visible lines for performance
	startLine := e.rowOffset
	endLine := e.rowOffset + 24 // approximate visible lines
	if endLine >= e.lineCount() {
		endLine = e.lineCount()
	}

	// Get position range for visible lines
	var startPos, endPos int
	if startLine < e.lineCount() {
		startPos = e.lineStartPos(startLine)
	}
	if endLine < e.lineCount() {
		endPos = e.lineStartPos(endLine)
	} else {
		endPos = len(text)
	}

	// Find all matches in visible range
	searchText := text[startPos:endPos]
	offset := startPos
	for {
		idx := strings.Index(searchText, query)
		if idx == -1 {
			break
		}
		e.searchMatches = append(e.searchMatches, offset+idx)
		searchText = searchText[idx+len(query):]
		offset += idx + len(query)
	}
}

// isSearchMatch checks if the given absolute position is part of a search match
func (e *Editor) isSearchMatch(pos int) bool {
	if e.searchQuery == "" || len(e.searchMatches) == 0 {
		return false
	}

	queryLen := len(e.searchQuery)
	for _, matchPos := range e.searchMatches {
		if pos >= matchPos && pos < matchPos+queryLen {
			return true
		}
	}
	return false
}
