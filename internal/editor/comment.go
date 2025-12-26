package editor

import (
	"strings"
)

// toggleLineComment toggles comment on a single line
func (e *Editor) toggleLineComment(lineNum int) {
	if lineNum < 0 || lineNum >= e.lineCount() {
		return
	}

	line := e.getLine(lineNum)
	commentPrefix := e.getCommentPrefix()

	// No comment prefix available for this filetype
	if commentPrefix == "" {
		return
	}

	// Trim the comment prefix if present
	trimmed := strings.TrimSpace(line)

	lineStart := e.lineStartPos(lineNum)
	lineEnd := lineStart + len(line)

	if strings.HasPrefix(trimmed, commentPrefix) {
		// Uncomment: remove comment prefix
		idx := strings.Index(line, commentPrefix)
		if idx != -1 {
			// Remove comment and one space after if present
			afterComment := idx + len(commentPrefix)
			if afterComment < len(line) && line[afterComment] == ' ' {
				afterComment++
			}
			newLine := line[:idx] + line[afterComment:]

			if err := e.buffer.Delete(lineStart, lineEnd); err != nil {
				return
			}
			if err := e.buffer.Insert(lineStart, newLine); err != nil {
				return
			}
		}
	} else {
		// Comment: add comment prefix
		// Find first non-whitespace character
		leadingSpace := 0
		for i, ch := range line {
			if ch != ' ' && ch != '\t' {
				leadingSpace = i
				break
			}
		}

		// Insert comment at first non-whitespace position
		newLine := line[:leadingSpace] + commentPrefix + " " + line[leadingSpace:]

		if err := e.buffer.Delete(lineStart, lineEnd); err != nil {
			return
		}
		if err := e.buffer.Insert(lineStart, newLine); err != nil {
			return
		}
	}

	e.dirty = true
}

// toggleCommentLines toggles comments on a range of lines
func (e *Editor) toggleCommentLines(startLine, endLine int) {
	for i := startLine; i <= endLine; i++ {
		e.toggleLineComment(i)
	}
}
