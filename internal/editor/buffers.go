package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// openFile opens a new file in a new buffer or switches to existing buffer
func (e *Editor) openFile(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		e.statusMsg = "error: " + err.Error()
		return
	}

	// Check if file is already open
	for i, buf := range e.buffers {
		if buf.filename == abs {
			e.FireBufLeave()
			e.syncToBuffer() // save current buffer state
			e.currentBuffer = i
			e.syncFromBuffer() // load new buffer state

			// Find and update the buffer split (not the current split which might be the tree)
			for _, split := range e.splits {
				if split.splitType == SplitBuffer {
					split.bufferIndex = i
					// Update split's view state from editor's current state
					split.cx = e.cx
					split.cy = e.cy
					split.rowOffset = e.rowOffset
					split.colOffset = e.colOffset
					split.wantX = e.wantX
					break
				}
			}

			e.FireBufEnter()
			e.statusMsg = fmt.Sprintf("switched to buffer %d: %s", i+1, filepath.Base(abs))
			return
		}
	}

	// Load the file
	txt := ""
	data, readErr := os.ReadFile(abs)
	if readErr == nil {
		txt = strings.ReplaceAll(string(data), "\r\n", "\n")
	}

	// Create new buffer
	bufView := NewBufferView(txt, abs)
	e.FireBufLeave()
	e.syncToBuffer() // save current buffer state first
	e.buffers = append(e.buffers, bufView)
	e.currentBuffer = len(e.buffers) - 1
	e.syncFromBuffer() // load new buffer state

	// Find and update the buffer split (not the current split which might be the tree)
	for _, split := range e.splits {
		if split.splitType == SplitBuffer {
			split.bufferIndex = e.currentBuffer
			// Update split's view state from editor's current state
			split.cx = e.cx
			split.cy = e.cy
			split.rowOffset = e.rowOffset
			split.colOffset = e.colOffset
			split.wantX = e.wantX
			break
		}
	}

	e.FireBufNew()
	e.FireBufRead()

	// Apply filetype-specific options
	e.setFiletypeOptions()

	// Sync parser from buffer view after filetype init
	if bv := e.buf(); bv != nil && bv.parser != nil {
		e.parser = bv.parser
		e.UpdateFoldStates()
		// Save fold ranges to buffer
		if bv.foldRanges == nil {
			bv.foldRanges = make(map[int]*FoldRange)
		}
		bv.foldRanges = e.foldRanges
	}

	// Fire FileType event
	if ft := e.getFiletype(); ft != nil {
		e.FireFileType(ft.Name)
	}
	e.FireBufEnter()

	if readErr != nil && !os.IsNotExist(readErr) {
		e.statusMsg = "error: " + readErr.Error()
	} else if readErr != nil && os.IsNotExist(readErr) {
		e.statusMsg = "new file"
	} else {
		e.statusMsg = fmt.Sprintf("loaded buffer %d: %s", e.currentBuffer+1, filepath.Base(abs))
	}
}

// nextBuffer switches to the next buffer
func (e *Editor) nextBuffer() {
	if len(e.buffers) == 0 {
		return
	}
	e.FireBufLeave()
	e.syncToBuffer() // save current buffer state
	e.currentBuffer = (e.currentBuffer + 1) % len(e.buffers)
	e.syncFromBuffer() // load new buffer state

	// Update current split to point to this buffer
	if len(e.splits) > 0 && e.currentSplit >= 0 && e.currentSplit < len(e.splits) {
		e.splits[e.currentSplit].bufferIndex = e.currentBuffer
	}

	e.FireBufEnter()
	e.statusMsg = fmt.Sprintf("buffer %d: %s", e.currentBuffer+1, filepath.Base(e.filename))
}

// prevBuffer switches to the previous buffer
func (e *Editor) prevBuffer() {
	if len(e.buffers) == 0 {
		return
	}
	e.FireBufLeave()
	e.syncToBuffer() // save current buffer state
	e.currentBuffer--
	if e.currentBuffer < 0 {
		e.currentBuffer = len(e.buffers) - 1
	}
	e.syncFromBuffer() // load new buffer state

	// Update current split to point to this buffer
	if len(e.splits) > 0 && e.currentSplit >= 0 && e.currentSplit < len(e.splits) {
		e.splits[e.currentSplit].bufferIndex = e.currentBuffer
	}

	e.FireBufEnter()
	e.statusMsg = fmt.Sprintf("buffer %d: %s", e.currentBuffer+1, filepath.Base(e.filename))
}

// listBuffers shows all open buffers
func (e *Editor) listBuffers() {
	if len(e.buffers) == 0 {
		e.statusMsg = "no buffers"
		return
	}

	lines := make([]string, 0, len(e.buffers))
	for i, buf := range e.buffers {
		dirtyMark := " "
		if buf.dirty {
			dirtyMark = "+"
		}
		current := " "
		if i == e.currentBuffer {
			current = "%"
		}

		// Show relative path if possible
		displayPath := buf.filename
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, buf.filename); err == nil {
				displayPath = rel
			}
		}
		if displayPath == "" {
			displayPath = "[No Name]"
		}

		lines = append(lines, fmt.Sprintf("%s%2d %s %s", current, i+1, dirtyMark, displayPath))
	}

	e.popupFixedH = 0 // auto-size
	e.openPopup("BUFFERS", lines)
}

// deleteBuffer closes the current buffer
func (e *Editor) deleteBuffer() {
	if len(e.buffers) == 1 {
		e.statusMsg = "cannot delete last buffer"
		return
	}

	if e.dirty {
		e.statusMsg = "No write since last change (use :bd! to override)"
		return
	}

	// Fire BufDelete event
	e.FireBufDelete(e.currentBuffer)

	// Remove current buffer
	e.buffers = append(e.buffers[:e.currentBuffer], e.buffers[e.currentBuffer+1:]...)

	// Adjust current buffer index
	if e.currentBuffer >= len(e.buffers) {
		e.currentBuffer = len(e.buffers) - 1
	}

	// Load the new current buffer
	e.syncFromBuffer()
	e.FireBufEnter()

	e.statusMsg = fmt.Sprintf("buffer deleted; now at buffer %d/%d", e.currentBuffer+1, len(e.buffers))
}

// deleteBufferForce closes the current buffer without checking if dirty
func (e *Editor) deleteBufferForce() {
	if len(e.buffers) == 1 {
		e.statusMsg = "cannot delete last buffer"
		return
	}

	// Fire BufDelete event
	e.FireBufDelete(e.currentBuffer)

	// Remove current buffer
	e.buffers = append(e.buffers[:e.currentBuffer], e.buffers[e.currentBuffer+1:]...)

	// Adjust current buffer index
	if e.currentBuffer >= len(e.buffers) {
		e.currentBuffer = len(e.buffers) - 1
	}

	// Load the new current buffer
	e.syncFromBuffer()
	e.FireBufEnter()

	e.statusMsg = fmt.Sprintf("buffer deleted; now at buffer %d/%d", e.currentBuffer+1, len(e.buffers))
}
