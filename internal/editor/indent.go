package editor

import (
	"strings"
)

// indentLines indents the specified line range by one indent level
func (e *Editor) indentLines(startLine, endLine int) {
	if startLine < 0 || endLine >= e.lineCount() || startLine > endLine {
		return
	}

	indent := strings.Repeat(" ", e.indentWidth)

	for line := startLine; line <= endLine; line++ {
		pos := e.lineStartPos(line)
		_ = e.buffer.Insert(pos, indent)
	}
}

// unindentLines removes one indent level from the specified line range
func (e *Editor) unindentLines(startLine, endLine int) {
	if startLine < 0 || endLine >= e.lineCount() || startLine > endLine {
		return
	}

	// Iterate backwards to avoid position shifting issues
	for line := endLine; line >= startLine; line-- {
		lineText := e.getLine(line)
		toRemove := 0
		spacesFound := 0

		for i, ch := range lineText {
			if ch == '\t' {
				// Tab counts as full indent
				toRemove = i + 1
				break
			} else if ch == ' ' {
				spacesFound++
				if spacesFound >= e.indentWidth {
					toRemove = i + 1
					break
				}
			} else {
				// Non-whitespace, remove what we found
				toRemove = i
				break
			}
		}

		if toRemove > 0 {
			pos := e.lineStartPos(line)
			_ = e.buffer.Delete(pos, pos+toRemove)
		}
	}
}

// autoIndentLines performs simple auto-indent on the specified line range
// This is a simple implementation that normalizes leading whitespace
func (e *Editor) autoIndentLines(startLine, endLine int) {
	if startLine < 0 || endLine >= e.lineCount() || startLine > endLine {
		return
	}

	// Iterate backwards to avoid position shifting issues
	for line := endLine; line >= startLine; line-- {
		lineText := e.getLine(line)
		currentIndent := 0
		firstNonSpace := 0
		for i, ch := range lineText {
			if ch == ' ' {
				currentIndent++
			} else if ch == '\t' {
				currentIndent += e.indentWidth
			} else {
				firstNonSpace = i
				break
			}
		}

		// Simple heuristic: normalize to nearest indent level
		// Round to nearest multiple: strictly less than midpoint rounds down
		remainder := currentIndent % e.indentWidth
		targetIndent := currentIndent - remainder
		// Use strict comparison for lower half, so 2 with width 4 rounds to 0
		if remainder > e.indentWidth/2 || (remainder == e.indentWidth/2 && currentIndent >= e.indentWidth) {
			targetIndent += e.indentWidth
		}

		// Only modify if target indent differs from current
		if targetIndent != currentIndent {
			// Get position fresh each time (in case prior modifications affected it)
			pos := e.lineStartPos(line)

			// Remove old indentation first
			if firstNonSpace > 0 {
				_ = e.buffer.Delete(pos, pos+firstNonSpace)
			}

			// Get position again after delete (should be the same, but being safe)
			pos = e.lineStartPos(line)

			// Add new indentation
			if targetIndent > 0 {
				indent := strings.Repeat(" ", targetIndent)
				_ = e.buffer.Insert(pos, indent)
			}
		}
	}
}
