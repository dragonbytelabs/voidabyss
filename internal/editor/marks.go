package editor

// setMark stores the current cursor position as a mark
func (e *Editor) setMark(name rune) {
	if name < 'a' || name > 'z' {
		e.statusMsg = "invalid mark name"
		return
	}
	e.marks[name] = Mark{line: e.cy, col: e.cx}
	e.statusMsg = ""
}

// jumpToMarkLine jumps to the line of a mark ('a)
func (e *Editor) jumpToMarkLine(name rune) {
	if name == '\'' {
		// '' jumps back to previous position
		e.jumpBack()
		return
	}

	mark, ok := e.marks[name]
	if !ok {
		e.statusMsg = "mark not set"
		return
	}

	// Add current position to jump list before jumping
	e.addToJumpList(e.cy, e.cx)

	e.cy = mark.line
	e.cx = 0
	e.ensureCursorValid()
	e.statusMsg = ""
}

// jumpToMarkExact jumps to exact position of a mark (`a)
func (e *Editor) jumpToMarkExact(name rune) {
	mark, ok := e.marks[name]
	if !ok {
		e.statusMsg = "mark not set"
		return
	}

	// Add current position to jump list before jumping
	e.addToJumpList(e.cy, e.cx)

	e.cy = mark.line
	e.cx = mark.col
	e.ensureCursorValid()
	e.statusMsg = ""
}

// addToJumpList adds a position to the jump history
func (e *Editor) addToJumpList(line, col int) {
	// If we're not at the end of the jump list, truncate it
	if e.jumpListIndex >= 0 {
		e.jumpList = e.jumpList[:e.jumpListIndex+1]
	}

	// Don't add duplicate consecutive entries
	if len(e.jumpList) > 0 {
		last := e.jumpList[len(e.jumpList)-1]
		if last.line == line && last.col == col {
			return
		}
	}

	// Add new entry
	e.jumpList = append(e.jumpList, JumpListEntry{line: line, col: col})

	// Limit jump list size
	if len(e.jumpList) > 100 {
		e.jumpList = e.jumpList[1:]
	}

	// Reset index to "at latest position"
	e.jumpListIndex = -1
}

// jumpBack jumps to previous position in jump list (Ctrl-O)
func (e *Editor) jumpBack() {
	if len(e.jumpList) == 0 {
		e.statusMsg = "jump list empty"
		return
	}

	// If at latest position, save current position and jump to last entry
	if e.jumpListIndex == -1 {
		if len(e.jumpList) > 0 {
			e.jumpListIndex = len(e.jumpList) - 1
			entry := e.jumpList[e.jumpListIndex]
			e.cy = entry.line
			e.cx = entry.col
			e.ensureCursorValid()
			e.statusMsg = ""
		}
		return
	}

	// Move back in jump list
	if e.jumpListIndex > 0 {
		e.jumpListIndex--
		entry := e.jumpList[e.jumpListIndex]
		e.cy = entry.line
		e.cx = entry.col
		e.ensureCursorValid()
		e.statusMsg = ""
	} else {
		e.statusMsg = "at oldest jump"
	}
}

// jumpForward jumps to next position in jump list (Ctrl-I)
func (e *Editor) jumpForward() {
	if len(e.jumpList) == 0 || e.jumpListIndex == -1 {
		e.statusMsg = "at newest jump"
		return
	}

	// Move forward in jump list
	if e.jumpListIndex < len(e.jumpList)-1 {
		e.jumpListIndex++
		entry := e.jumpList[e.jumpListIndex]
		e.cy = entry.line
		e.cx = entry.col
		e.ensureCursorValid()
		e.statusMsg = ""
	} else {
		// At the end, return to current position
		e.jumpListIndex = -1
		e.statusMsg = "at newest jump"
	}
}
