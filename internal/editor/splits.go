package editor

import "fmt"

// SplitType represents the type of split
type SplitType int

const (
	SplitBuffer SplitType = iota
	SplitFileTree
)

// Split represents a window split showing a buffer or file tree
type Split struct {
	splitType   SplitType // type of split (buffer or file tree)
	bufferIndex int       // index into Editor.buffers (only for SplitBuffer)
	width       int       // width of this split in columns
	height      int       // height of this split in rows
	x           int       // x position on screen (relative to content area)
	y           int       // y position on screen

	// View state for this split (only for SplitBuffer)
	cx        int // cursor x position
	cy        int // cursor y position
	rowOffset int // scroll offset
	colOffset int // scroll offset
	wantX     int // desired x position for vertical movement

	// Visual mode state (only for SplitBuffer)
	visualActive bool
	visualAnchor int
	visualKind   VisualKind
}

// initSplits initializes the split system with a single split
func (e *Editor) initSplits() {
	w, h := e.s.Size()

	// Reserve space for status line
	height := h - 1

	// If file tree is open, create it as the first split
	if e.treeOpen && e.fileTree != nil {
		treeWidth := e.treePanelWidth
		if treeWidth > w-10 {
			treeWidth = w - 10
		}
		if treeWidth < 20 {
			treeWidth = 20
		}

		e.splits = []*Split{
			{
				splitType: SplitFileTree,
				width:     treeWidth,
				height:    height,
				x:         0,
				y:         0,
			},
			{
				splitType:   SplitBuffer,
				bufferIndex: e.currentBuffer,
				width:       w - treeWidth - 1, // -1 for border
				height:      height,
				x:           treeWidth + 1, // actual screen position for navigation
				y:           0,
				cx:          e.cx,
				cy:          e.cy,
				rowOffset:   e.rowOffset,
				colOffset:   e.colOffset,
				wantX:       e.wantX,
			},
		}
		e.currentSplit = 1 // Start with buffer split focused
	} else {
		e.splits = []*Split{
			{
				splitType:   SplitBuffer,
				bufferIndex: e.currentBuffer,
				width:       w,
				height:      height,
				x:           0,
				y:           0,
				cx:          e.cx,
				cy:          e.cy,
				rowOffset:   e.rowOffset,
				colOffset:   e.colOffset,
				wantX:       e.wantX,
			},
		}
		e.currentSplit = 0
	}
}

// saveSplitState saves the current editor state to the current split
func (e *Editor) saveSplitState() {
	if e.currentSplit < 0 || e.currentSplit >= len(e.splits) {
		return
	}

	split := e.splits[e.currentSplit]

	// Only save view state for buffer splits
	if split.splitType == SplitBuffer {
		split.cx = e.cx
		split.cy = e.cy
		split.rowOffset = e.rowOffset
		split.colOffset = e.colOffset
		split.wantX = e.wantX
		// Save visual mode state
		split.visualActive = e.visualActive
		split.visualAnchor = e.visualAnchor
		split.visualKind = e.visualKind
	}
}

// vsplit creates a vertical split
func (e *Editor) vsplit() {
	if len(e.splits) == 0 {
		e.initSplits()
		return
	}

	currentSplit := e.splits[e.currentSplit]

	// Can't split file tree
	if currentSplit.splitType == SplitFileTree {
		e.statusMsg = "cannot split file tree"
		return
	}

	// Save current editor state to current split
	e.saveSplitState()

	// Need at least 10 columns per split
	if currentSplit.width < 20 {
		e.statusMsg = "not enough space for split"
		return
	}

	// Split current window in half
	newWidth := currentSplit.width / 2
	currentSplit.width = newWidth

	// Create new split to the right with same buffer but independent view state
	newSplit := &Split{
		splitType:    SplitBuffer,
		bufferIndex:  e.currentBuffer, // same buffer as current
		width:        newWidth,
		height:       currentSplit.height,
		x:            currentSplit.x + newWidth,
		y:            currentSplit.y,
		cx:           e.cx,
		cy:           e.cy,
		rowOffset:    e.rowOffset,
		colOffset:    e.colOffset,
		wantX:        e.wantX,
		visualActive: e.visualActive,
		visualAnchor: e.visualAnchor,
		visualKind:   e.visualKind,
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

	// Can't split file tree
	if currentSplit.splitType == SplitFileTree {
		e.statusMsg = "cannot split file tree"
		return
	}

	// Save current editor state to current split
	e.saveSplitState()

	// Need at least 6 rows per split
	if currentSplit.height < 12 {
		e.statusMsg = "not enough space for split"
		return
	}

	// Split current window in half
	newHeight := currentSplit.height / 2
	currentSplit.height = newHeight

	// Create new split below with same buffer but independent view state
	newSplit := &Split{
		splitType:    SplitBuffer,
		bufferIndex:  e.currentBuffer, // same buffer as current
		width:        currentSplit.width,
		height:       newHeight,
		x:            currentSplit.x,
		y:            currentSplit.y + newHeight,
		cx:           e.cx,
		cy:           e.cy,
		rowOffset:    e.rowOffset,
		colOffset:    e.colOffset,
		wantX:        e.wantX,
		visualActive: e.visualActive,
		visualAnchor: e.visualAnchor,
		visualKind:   e.visualKind,
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

	// Can't close file tree split
	if e.splits[e.currentSplit].splitType == SplitFileTree {
		e.statusMsg = "cannot close file tree (use :tree to toggle)"
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

	// Load the new current split's state
	e.loadSplitState()

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

	// Save current split state before switching
	e.saveSplitState()

	e.currentSplit = (e.currentSplit + 1) % len(e.splits)

	// Load new split state
	e.loadSplitState()

	e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
}

// prevSplit moves focus to the previous split
func (e *Editor) prevSplit() {
	if len(e.splits) <= 1 {
		return
	}

	// Save current split state before switching
	e.saveSplitState()

	e.currentSplit--
	if e.currentSplit < 0 {
		e.currentSplit = len(e.splits) - 1
	}

	// Load new split state
	e.loadSplitState()

	e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
}

// moveSplitLeft moves focus to the split on the left
func (e *Editor) moveSplitLeft() {
	if len(e.splits) <= 1 {
		return
	}

	current := e.splits[e.currentSplit]

	// Find split to the left (same or overlapping Y, smaller X)
	bestMatch := -1
	bestX := -1

	for i, split := range e.splits {
		if i == e.currentSplit {
			continue
		}

		// Check if Y ranges overlap
		if split.y < current.y+current.height && split.y+split.height > current.y {
			// This split's Y overlaps with current
			if split.x < current.x && split.x > bestX {
				bestMatch = i
				bestX = split.x
			}
		}
	}

	if bestMatch != -1 {
		e.saveSplitState()
		e.currentSplit = bestMatch
		e.loadSplitState()
		e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
	} else {
		e.statusMsg = "no split to the left"
	}
}

// moveSplitRight moves focus to the split on the right
func (e *Editor) moveSplitRight() {
	if len(e.splits) <= 1 {
		return
	}

	current := e.splits[e.currentSplit]

	// Find split to the right (same or overlapping Y, larger X)
	bestMatch := -1
	bestX := 999999

	for i, split := range e.splits {
		if i == e.currentSplit {
			continue
		}

		// Check if Y ranges overlap
		if split.y < current.y+current.height && split.y+split.height > current.y {
			// This split's Y overlaps with current
			if split.x > current.x && split.x < bestX {
				bestMatch = i
				bestX = split.x
			}
		}
	}

	if bestMatch != -1 {
		e.saveSplitState()
		e.currentSplit = bestMatch
		e.loadSplitState()
		e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
	} else {
		e.statusMsg = "no split to the right"
	}
}

// moveSplitUp moves focus to the split above
func (e *Editor) moveSplitUp() {
	if len(e.splits) <= 1 {
		return
	}

	current := e.splits[e.currentSplit]

	// Find split above (same or overlapping X, smaller Y)
	bestMatch := -1
	bestY := -1

	for i, split := range e.splits {
		if i == e.currentSplit {
			continue
		}

		// Check if X ranges overlap
		if split.x < current.x+current.width && split.x+split.width > current.x {
			// This split's X overlaps with current
			if split.y < current.y && split.y > bestY {
				bestMatch = i
				bestY = split.y
			}
		}
	}

	if bestMatch != -1 {
		e.saveSplitState()
		e.currentSplit = bestMatch
		e.loadSplitState()
		e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
	} else {
		e.statusMsg = "no split above"
	}
}

// moveSplitDown moves focus to the split below
func (e *Editor) moveSplitDown() {
	if len(e.splits) <= 1 {
		return
	}

	current := e.splits[e.currentSplit]

	// Find split below (same or overlapping X, larger Y)
	bestMatch := -1
	bestY := 999999

	for i, split := range e.splits {
		if i == e.currentSplit {
			continue
		}

		// Check if X ranges overlap
		if split.x < current.x+current.width && split.x+split.width > current.x {
			// This split's X overlaps with current
			if split.y > current.y && split.y < bestY {
				bestMatch = i
				bestY = split.y
			}
		}
	}

	if bestMatch != -1 {
		e.saveSplitState()
		e.currentSplit = bestMatch
		e.loadSplitState()
		e.statusMsg = fmt.Sprintf("split %d/%d", e.currentSplit+1, len(e.splits))
	} else {
		e.statusMsg = "no split below"
	}
}

// loadSplitState loads view state from the current split to the editor
func (e *Editor) loadSplitState() {
	if len(e.splits) == 0 || e.currentSplit < 0 || e.currentSplit >= len(e.splits) {
		return
	}

	split := e.splits[e.currentSplit]

	// Handle file tree split
	if split.splitType == SplitFileTree {
		e.focusTree = true
		return
	}

	// Handle buffer split
	e.focusTree = false

	// Switch to split's buffer if different
	if split.bufferIndex != e.currentBuffer && split.bufferIndex < len(e.buffers) {
		e.saveCurrentBufferState()
		e.currentBuffer = split.bufferIndex
		e.loadBufferState(e.buffers[e.currentBuffer])
	}

	// Load split's view state
	e.cx = split.cx
	e.cy = split.cy
	e.rowOffset = split.rowOffset
	e.colOffset = split.colOffset
	e.wantX = split.wantX

	// Load visual mode state
	e.visualActive = split.visualActive
	e.visualAnchor = split.visualAnchor
	e.visualKind = split.visualKind
}

// syncSplitToEditor is deprecated - use loadSplitState
func (e *Editor) syncSplitToEditor() {
	e.loadSplitState()
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
