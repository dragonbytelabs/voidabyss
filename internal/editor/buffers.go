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
			e.syncToBuffer() // save current buffer state
			e.currentBuffer = i
			e.syncFromBuffer() // load new buffer state
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
	e.syncToBuffer() // save current buffer state first
	e.buffers = append(e.buffers, bufView)
	e.currentBuffer = len(e.buffers) - 1
	e.syncFromBuffer() // load new buffer state

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
	e.syncToBuffer() // save current buffer state
	e.currentBuffer = (e.currentBuffer + 1) % len(e.buffers)
	e.syncFromBuffer() // load new buffer state
	e.statusMsg = fmt.Sprintf("buffer %d: %s", e.currentBuffer+1, filepath.Base(e.filename))
}

// prevBuffer switches to the previous buffer
func (e *Editor) prevBuffer() {
	if len(e.buffers) == 0 {
		return
	}
	e.syncToBuffer() // save current buffer state
	e.currentBuffer--
	if e.currentBuffer < 0 {
		e.currentBuffer = len(e.buffers) - 1
	}
	e.syncFromBuffer() // load new buffer state
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
		name := filepath.Base(buf.filename)
		if name == "" {
			name = "[No Name]"
		}
		lines = append(lines, fmt.Sprintf("%s%2d %s %s", current, i+1, dirtyMark, name))
	}

	// Join with newlines for multi-line display, or with separators if few buffers
	if len(lines) <= 4 {
		e.statusMsg = strings.Join(lines, "  |  ")
	} else {
		e.statusMsg = strings.Join(lines, "\n")
	}
}
