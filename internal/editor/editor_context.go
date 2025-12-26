package editor

// Implement config.EditorContext interface for Editor

// GetText returns the buffer content
func (e *Editor) GetText() string {
	return e.buffer.String()
}

// SetText sets the buffer content
func (e *Editor) SetText(text string) {
	// Replace entire buffer content
	oldLen := e.buffer.Len()
	if oldLen > 0 {
		if err := e.buffer.Delete(0, oldLen); err != nil {
			return
		}
	}
	if len(text) > 0 {
		if err := e.buffer.Insert(0, text); err != nil {
			return
		}
	}
	e.dirty = true
	e.ensureCursorValid()
}

// GetName returns the filename
func (e *Editor) GetName() string {
	return e.filename
}

// SetName sets the filename
func (e *Editor) SetName(name string) {
	e.filename = name
	e.dirty = true
}

// Cursor returns the current cursor position (0-indexed line, column)
func (e *Editor) Cursor() (row, col int) {
	return e.cy, e.cx
}

// SetCursor sets the cursor position (0-indexed line, column)
func (e *Editor) SetCursor(row, col int) {
	e.cy = row
	e.cx = col
	e.ensureCursorValid()
}

// Line returns the text of line y (0-indexed)
func (e *Editor) Line(y int) string {
	return e.getLine(y)
}

// SetLine sets the text of line y (0-indexed)
func (e *Editor) SetLine(y int, text string) {
	starts := e.lineStarts()
	if y < 0 || y >= len(starts) {
		return
	}

	// Get line boundaries
	start := starts[y]
	runes := e.textRunes()
	end := len(runes)
	if y+1 < len(starts) {
		end = starts[y+1] - 1 // Exclude newline
		if end < start {
			end = start
		}
	}

	// Delete old line content
	if end > start {
		if err := e.buffer.Delete(start, end); err != nil {
			return
		}
	}

	// Insert new content
	if len(text) > 0 {
		if err := e.buffer.Insert(start, text); err != nil {
			return
		}
	}

	e.dirty = true
	e.ensureCursorValid()
}

// Insert inserts text at the given rune position
func (e *Editor) Insert(pos int, text string) {
	if pos < 0 || pos > e.buffer.Len() {
		return
	}
	if err := e.buffer.Insert(pos, text); err != nil {
		// Log error but don't crash
		return
	}
	e.dirty = true
	e.ensureCursorValid()
}

// Delete deletes text from start to end rune positions
func (e *Editor) Delete(start, end int) {
	if start < 0 || end < start || end > e.buffer.Len() {
		return
	}
	if err := e.buffer.Delete(start, end); err != nil {
		// Log error but don't crash
		return
	}
	e.dirty = true
	e.ensureCursorValid()
}

// Len returns the buffer length in runes
func (e *Editor) Len() int {
	return e.buffer.Len()
}

// Mode returns the current editor mode as a string
func (e *Editor) Mode() string {
	switch e.mode {
	case ModeNormal:
		return "normal"
	case ModeInsert:
		return "insert"
	case ModeVisual:
		return "visual"
	case ModeCommand:
		return "command"
	default:
		return "normal"
	}
}

// SetMode sets the editor mode from a string
func (e *Editor) SetMode(mode string) {
	switch mode {
	case "normal":
		e.mode = ModeNormal
	case "insert":
		e.mode = ModeInsert
	case "visual":
		e.mode = ModeVisual
	case "command":
		e.mode = ModeCommand
	}
}

// RegisterWithLoader registers this editor as the context for Lua buffer operations
func (e *Editor) RegisterWithLoader() {
	if e.loader != nil {
		e.loader.SetEditorContext(e)
	}
}
