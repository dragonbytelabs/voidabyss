package editor

import (
	"fmt"
	"regexp"
	"strings"
)

// searchNext finds the next occurrence of searchQuery
// skipCurrent: if true, skip a match at current position (for 'n'/'N')
func (e *Editor) searchNext(forward bool, skipCurrent bool) {
	if e.searchQuery == "" {
		return
	}

	// Add current position to jump list before jumping to search result
	e.addToJumpList(e.cy, e.cx)

	text := e.buffer.String()
	query := e.searchQuery

	// Try to compile as regex first, fall back to literal search
	re, err := regexp.Compile(query)
	useRegex := err == nil

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
			if useRegex {
				loc := re.FindStringIndex(text[startPos:])
				if loc != nil {
					matchPos = startPos + loc[0]
					found = true
				}
			} else {
				idx := strings.Index(text[startPos:], query)
				if idx != -1 {
					matchPos = startPos + idx
					found = true
				}
			}
		}

		// Wrap around to beginning if not found
		if !found {
			if useRegex {
				loc := re.FindStringIndex(text)
				if loc != nil && loc[0] < currentPos {
					matchPos = loc[0]
					found = true
					e.statusMsg = "search wrapped"
				}
			} else {
				idx := strings.Index(text, query)
				if idx != -1 && idx < currentPos {
					matchPos = idx
					found = true
					e.statusMsg = "search wrapped"
				}
			}
		}
	} else {
		// Search backward
		searchEnd := currentPos
		if skipCurrent && currentPos > 0 {
			searchEnd = currentPos - 1
		}

		if searchEnd > 0 {
			if useRegex {
				// Find all matches up to searchEnd, take the last one
				allMatches := re.FindAllStringIndex(text[:searchEnd+1], -1)
				if len(allMatches) > 0 {
					loc := allMatches[len(allMatches)-1]
					matchPos = loc[0]
					found = true
				}
			} else {
				idx := strings.LastIndex(text[:searchEnd+1], query)
				if idx != -1 {
					matchPos = idx
					found = true
				}
			}
		}

		// Wrap around to end if not found
		if !found {
			if useRegex {
				allMatches := re.FindAllStringIndex(text, -1)
				if len(allMatches) > 0 {
					loc := allMatches[len(allMatches)-1]
					if loc[0] > currentPos {
						matchPos = loc[0]
						found = true
						e.statusMsg = "search wrapped"
					}
				}
			} else {
				idx := strings.LastIndex(text, query)
				if idx != -1 && idx > currentPos {
					matchPos = idx
					found = true
					e.statusMsg = "search wrapped"
				}
			}
		}
	}

	if found {
		e.setCursorFromPos(matchPos)
		e.wantX = e.cx
		e.updateSearchHighlights()
		// Don't overwrite "search wrapped" message
		if e.statusMsg != "search wrapped" {
			e.updateSearchStatus()
		}
	} else {
		e.statusMsg = "pattern not found: " + query
		e.searchMatches = nil
	}
}

// updateSearchHighlights finds all matches in the buffer
func (e *Editor) updateSearchHighlights() {
	if e.searchQuery == "" {
		e.searchMatches = nil
		return
	}

	text := e.buffer.String()
	query := e.searchQuery
	e.searchMatches = nil

	// Try regex first, fall back to literal
	re, err := regexp.Compile(query)

	if err == nil {
		// Regex search - find all matches
		allMatches := re.FindAllStringIndex(text, -1)
		for _, loc := range allMatches {
			e.searchMatches = append(e.searchMatches, loc[0])
		}
	} else {
		// Literal search - find all matches
		offset := 0
		searchText := text
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
}

// updateSearchStatus updates the status line with match count and position
func (e *Editor) updateSearchStatus() {
	if e.searchQuery == "" || len(e.searchMatches) == 0 {
		return
	}

	// Find which match we're currently at
	currentPos := e.posFromCursor()
	currentMatch := -1

	for i, matchPos := range e.searchMatches {
		if matchPos >= currentPos {
			currentMatch = i + 1
			break
		}
	}

	if currentMatch == -1 {
		// Cursor is after all matches
		currentMatch = len(e.searchMatches)
	}

	e.statusMsg = fmt.Sprintf("/%s [%d/%d]", e.searchQuery, currentMatch, len(e.searchMatches))
}

// isSearchMatch checks if the given absolute position is part of a search match
func (e *Editor) isSearchMatch(pos int) bool {
	if e.searchQuery == "" || len(e.searchMatches) == 0 {
		return false
	}

	// Try regex to get match length
	re, err := regexp.Compile(e.searchQuery)
	var queryLen int

	if err == nil {
		// For regex, we need to check each match individually
		text := e.buffer.String()
		for _, matchPos := range e.searchMatches {
			if matchPos > pos {
				break
			}
			// Get the actual match at this position
			loc := re.FindStringIndex(text[matchPos:])
			if loc != nil && loc[0] == 0 {
				matchLen := loc[1]
				if pos >= matchPos && pos < matchPos+matchLen {
					return true
				}
			}
		}
		return false
	} else {
		// Literal search - fixed length
		queryLen = len(e.searchQuery)
		for _, matchPos := range e.searchMatches {
			if pos >= matchPos && pos < matchPos+queryLen {
				return true
			}
		}
		return false
	}
}

// performIncrementalSearch performs search as user types (for / and ? modes)
func (e *Editor) performIncrementalSearch() {
	if len(e.searchBuf) == 0 {
		e.searchMatches = nil
		e.statusMsg = ""
		return
	}

	query := string(e.searchBuf)
	text := e.buffer.String()
	currentPos := e.posFromCursor()

	// Try to find first match from current position
	re, err := regexp.Compile(query)
	useRegex := err == nil

	var matchPos int
	found := false

	if e.searchForward {
		// Forward search from current position
		if useRegex {
			loc := re.FindStringIndex(text[currentPos:])
			if loc != nil {
				matchPos = currentPos + loc[0]
				found = true
			}
		} else {
			idx := strings.Index(text[currentPos:], query)
			if idx != -1 {
				matchPos = currentPos + idx
				found = true
			}
		}
	} else {
		// Backward search from current position
		if useRegex {
			allMatches := re.FindAllStringIndex(text[:currentPos+1], -1)
			if len(allMatches) > 0 {
				loc := allMatches[len(allMatches)-1]
				matchPos = loc[0]
				found = true
			}
		} else {
			idx := strings.LastIndex(text[:currentPos+1], query)
			if idx != -1 {
				matchPos = idx
				found = true
			}
		}
	}

	if found {
		// Temporarily move cursor to show the match
		e.setCursorFromPos(matchPos)
		e.wantX = e.cx

		// Update search query and highlights
		e.searchQuery = query
		e.updateSearchHighlights()
		e.updateSearchStatus()
	} else {
		// No match found
		e.searchMatches = nil
		if len(query) > 0 {
			e.statusMsg = fmt.Sprintf("/%s [0/0]", query)
		}
	}
}
