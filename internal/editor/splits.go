package editor

import "fmt"

// Split represents a window split showing a buffer
type Split struct {
	bufferIndex int // index into Editor.buffers
	width       int // width of this split in columns
	height      int // height of this split in rows
	x           int // x position on screen
	y           int // y position on screen
}

// initSplits initializes the split system with a single split
func (e *Editor) initSplits() {
	w, h := e.s.Size()

	// Reserve space for status line
	height := h - 1

	e.splits = []*Split{
		{
			bufferIndex: e.currentBuffer,
			width:       w,
			height:      height,
			x:           0,
			y:           0,
		},
	}
	e.currentSplit = 0
}

// vsplit creates a vertical split
func (e *Editor) vsplit() {
	if len(e.splits) == 0 {
		e.initSplits()
		return
	}

	currentSplit := e.splits[e.currentSplit]

	// Need at least 10 columns per split
	if currentSplit.width < 20 {
		e.statusMsg = "not enough space for split"
		return
	}

	// Split current window in half
	newWidth := currentSplit.width / 2
	currentSplit.width = newWidth

	// Create new split to the right
	newSplit := &Split{
		bufferIndex: e.currentBuffer, // same buffer as current
		width:       currentSplit.width - newWidth,
		height:      currentSplit.height,
		x:           currentSplit.x + newWidth,
		y:           currentSplit.y,
	}

	// Insert new split after current
	e.splits = append(e.splits[:e.currentSplit+1], append([]*Split{newSplit}, e.splits[e.currentSplit+1:]...)...)

	// Focus new split
	e.currentSplit++

	e.statusMsg = fmt.Sprintf("split created (%d splits)", len(e.splits))
}

// split creates a horizontal split
func (e *Editor) split() {
	if len(e.splits) == 0 {
		e.initSplits()
		return
	}

	currentSplit := e.splits[e.currentSplit]

	// Need at least 6 rows per split
	if currentSplit.height < 12 {
		e.statusMsg = "not enough space for split"
		return
	}

	// Split current window in half
	newHeight := currentSplit.height / 2
	currentSplit.height = newHeight

	// Create new split below
	newSplit := &Split{
		bufferIndex: e.currentBuffer, // same buffer as current
		width:       currentSplit.width,
		height:      currentSplit.height - newHeight,
		x:           currentSplit.x,
		y:           currentSplit.y + newHeight,
	}

	// Insert new split after current
	e.splits = append(e.splits[:e.currentSplit+1], append([]*Split{newSplit}, e.splits[e.currentSplit+1:]...)...)

	// Focus new split
	e.currentSplit++

	e.statusMsg = fmt.Sprintf("split created (%d splits)", len(e.splits))
}

// closeSplit closes the current split
func (e *Editor) closeSplit() {
	if len(e.splits) <= 1 {
		e.statusMsg = "cannot close last split"
		return
	}

	// Validate currentSplit index
	if e.currentSplit < 0 || e.currentSplit >= len(e.splits) {
		e.currentSplit = 0
		return
	}

	// Remove current split
	e.splits = append(e.splits[:e.currentSplit], e.splits[e.currentSplit+1:]...)

	// Adjust focus
	if e.currentSplit >= len(e.splits) {
		e.currentSplit = len(e.splits) - 1
	}

	// Ensure currentSplit is valid
	if e.currentSplit < 0 {
		e.currentSplit = 0
	}

	// TODO: Redistribute space among remaining splits
	e.redistributeSplitSpace()

	e.statusMsg = fmt.Sprintf("%d splits remaining", len(e.splits))
}

// redistributeSplitSpace evenly distributes space among splits
func (e *Editor) redistributeSplitSpace() {
	if len(e.splits) == 0 {
		return
	}

	w, h := e.s.Size()
	height := h - 1 // Reserve space for status line

	// Simple case: stack all splits horizontally
	splitWidth := w / len(e.splits)
	x := 0

	for i, split := range e.splits {
		split.x = x
		split.y = 0
		split.width = splitWidth
		split.height = height

		// Give remaining pixels to last split
		if i == len(e.splits)-1 {
			split.width = w - x
		}

		x += split.width
	}
}

// nextSplit moves focus to the next split
func (e *Editor) nextSplit() {
	if len(e.splits) <= 1 {
		return
	}

	e.currentSplit = (e.currentSplit + 1) % len(e.splits)
	e.syncSplitToEditor()
	e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
}

// prevSplit moves focus to the previous split
func (e *Editor) prevSplit() {
	if len(e.splits) <= 1 {
		return
	}

	e.currentSplit--
	if e.currentSplit < 0 {
		e.currentSplit = len(e.splits) - 1
	}

	e.syncSplitToEditor()
	e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
}

// syncSplitToEditor loads split's buffer into editor
func (e *Editor) syncSplitToEditor() {
	if len(e.splits) == 0 || e.currentSplit < 0 || e.currentSplit >= len(e.splits) {
		return
	}

	split := e.splits[e.currentSplit]

	// Switch to split's buffer if different
	if split.bufferIndex != e.currentBuffer && split.bufferIndex < len(e.buffers) {
		e.saveCurrentBufferState()
		e.currentBuffer = split.bufferIndex
		e.loadBufferState(e.buffers[e.currentBuffer])
	}
}

// saveCurrentBufferState saves the current editor state into the current buffer view
func (e *Editor) saveCurrentBufferState() {
	if len(e.buffers) == 0 || e.currentBuffer < 0 || e.currentBuffer >= len(e.buffers) {
		return
	}

	bv := e.buffers[e.currentBuffer]
	bv.buffer = e.buffer
	bv.filename = e.filename
	bv.dirty = e.dirty
	bv.cx = e.cx
	bv.cy = e.cy
	bv.rowOffset = e.rowOffset
	bv.colOffset = e.colOffset
	bv.wantX = e.wantX
	bv.marks = e.marks
	bv.jumpList = e.jumpList
	bv.jumpListIndex = e.jumpListIndex
}

// loadBufferState loads state from a buffer view into the editor
func (e *Editor) loadBufferState(bv *BufferView) {
	if bv == nil {
		return
	}
	e.buffer = bv.buffer
	e.filename = bv.filename
	e.dirty = bv.dirty
	e.cx = bv.cx
	e.cy = bv.cy
	e.rowOffset = bv.rowOffset
	e.colOffset = bv.colOffset
	e.wantX = bv.wantX
	e.marks = bv.marks
	e.jumpList = bv.jumpList
	e.jumpListIndex = bv.jumpListIndex
}
