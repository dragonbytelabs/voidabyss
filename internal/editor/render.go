package editor

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) draw() {
	e.s.Clear()
	w, h := e.s.Size()
	style := tcell.StyleDefault
	highlightStyle := tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack)
	visualStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	treeStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	treeCursorStyle := tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	treeBorderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	contentStartX := 0
	contentWidth := w

	// Draw file tree if open
	if e.treeOpen && e.fileTree != nil {
		treeWidth := e.treePanelWidth
		if treeWidth > w-10 {
			treeWidth = w - 10
		}
		if treeWidth < 20 {
			treeWidth = 20
		}

		lines := e.fileTree.getDisplayLines()
		for y := 0; y < h-1 && y < len(lines); y++ {
			lineRunes := []rune(lines[y])
			isCursor := y < len(lines) && len(lineRunes) > 0 && lineRunes[0] == '>'

			for x := 0; x < treeWidth && x < len(lineRunes); x++ {
				cellStyle := treeStyle
				if isCursor {
					cellStyle = treeCursorStyle
				}
				e.s.SetContent(x, y, lineRunes[x], nil, cellStyle)
			}
			// Clear rest of tree panel width
			for x := len(lineRunes); x < treeWidth; x++ {
				e.s.SetContent(x, y, ' ', nil, style)
			}
		}
		// Clear any remaining lines in tree panel
		for y := len(lines); y < h-1; y++ {
			for x := 0; x < treeWidth; x++ {
				e.s.SetContent(x, y, ' ', nil, style)
			}
		}

		// Draw vertical border
		for y := 0; y < h-1; y++ {
			e.s.SetContent(treeWidth, y, '│', nil, treeBorderStyle)
		}

		contentStartX = treeWidth + 1
		contentWidth = w - contentStartX
	}

	// Draw buffer content
	for y := 0; y < h-1; y++ {
		lineIndex := e.rowOffset + y
		if lineIndex >= e.lineCount() {
			break
		}

		runes := []rune(e.getLine(lineIndex))
		start := min(e.colOffset, len(runes))
		visible := runes[start:]

		// Calculate absolute position for highlight matching
		lineStartPos := e.lineStartPos(lineIndex)

		for x := 0; x < contentWidth && x < len(visible); x++ {
			absPos := lineStartPos + start + x
			cellStyle := style

			// Check if this position is in visual selection (takes priority)
			if e.isInVisualSelection(absPos) {
				cellStyle = visualStyle
			} else if e.isSearchMatch(absPos) {
				// Check if this position is in a search match
				cellStyle = highlightStyle
			}

			e.s.SetContent(contentStartX+x, y, visible[x], nil, cellStyle)
		}
	}

	e.drawStatus(w, h)

	if e.popupActive {
		e.drawPopup(w, h)
	}

	// Position cursor
	screenX := e.cx - e.colOffset + contentStartX
	screenY := e.cy - e.rowOffset
	screenX = max(contentStartX, screenX)
	screenY = max(0, screenY)
	if screenY > h-2 {
		screenY = h - 2
	}

	// Only show cursor if buffer has focus
	if !e.focusTree {
		e.s.ShowCursor(screenX, screenY)
	} else {
		e.s.HideCursor()
	}
	e.s.Show()
}

func (e *Editor) drawStatus(w, h int) {
	modeStr := map[Mode]string{
		ModeNormal:  "NORMAL",
		ModeInsert:  "INSERT",
		ModeCommand: "COMMAND",
		ModeVisual:  "VISUAL",
		ModeSearch:  "SEARCH",
	}[e.mode]

	regCh := rune('"')
	if e.regOverrideSet {
		regCh = e.regOverride
	}

	left := fmt.Sprintf("%s  %s  reg:%c", modeStr, e.filename, regCh)
	if e.dirty {
		left += " [+]"
	}

	msg := e.statusMsg
	if e.mode == ModeNormal {
		if e.pendingCount > 0 {
			msg = fmt.Sprintf("%d", e.pendingCount)
		}
		if e.pendingOp != 0 {
			msg += string(e.pendingOp)
		}
		if e.pendingTextObj != 0 {
			msg += string(e.pendingTextObj)
		}
	}
	if e.mode == ModeCommand {
		msg = ":" + string(e.cmdBuf)
	}
	if e.mode == ModeSearch {
		prefix := "/"
		if !e.searchForward {
			prefix = "?"
		}
		msg = prefix + string(e.searchBuf)
	}

	bar := left
	for len([]rune(bar)) < w {
		bar += " "
	}
	for x, r := range []rune(bar) {
		if x >= w {
			break
		}
		e.s.SetContent(x, h-1, r, nil, tcell.StyleDefault.Reverse(true))
	}
	if msg != "" {
		startX := min(len([]rune(left))+2, w-1)
		for i, r := range []rune(msg) {
			x := startX + i
			if x >= w {
				break
			}
			e.s.SetContent(x, h-1, r, nil, tcell.StyleDefault.Reverse(true))
		}
	}
}

// popup functions (same as your current version)
func (e *Editor) openPopup(title string, lines []string) {
	e.popupActive = true
	e.popupTitle = title
	e.popupLines = lines
	e.popupScroll = 0
}

func (e *Editor) closePopup() {
	e.popupActive = false
	e.popupTitle = ""
	e.popupLines = nil
	e.popupFixedH = 0
	e.popupScroll = 0
}

func (e *Editor) drawPopup(w, h int) {
	if !e.popupActive {
		return
	}

	lines := e.popupLines
	if lines == nil {
		lines = []string{}
	}

	contentW := 0
	for _, line := range lines {
		l := len([]rune(line))
		if l > contentW {
			contentW = l
		}
	}
	titleW := len([]rune(e.popupTitle))
	if titleW+2 > contentW {
		contentW = titleW + 2
	}

	if contentW > w-6 {
		contentW = w - 6
	}
	if contentW < 20 {
		contentW = 20
	}

	displayH := len(lines)
	if displayH > h-6 {
		displayH = h - 6
	}
	if displayH < 1 {
		displayH = 1
	}

	minVisualH := 3
	visualH := displayH
	if e.popupFixedH > 0 {
		visualH = e.popupFixedH
	}
	if visualH > h-6 {
		visualH = h - 6
	}
	if visualH < minVisualH {
		visualH = minVisualH
	}

	boxW := contentW + 4
	boxH := visualH + 4

	x0 := (w - boxW) / 2
	y0 := (h - boxH) / 2
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}

	border := tcell.StyleDefault.Reverse(true)
	textStyle := tcell.StyleDefault.Reverse(true)

	for yy := 0; yy < boxH; yy++ {
		for xx := 0; xx < boxW; xx++ {
			e.s.SetContent(x0+xx, y0+yy, ' ', nil, border)
		}
	}

	e.s.SetContent(x0, y0, '+', nil, border)
	e.s.SetContent(x0+boxW-1, y0, '+', nil, border)
	e.s.SetContent(x0, y0+boxH-1, '+', nil, border)
	e.s.SetContent(x0+boxW-1, y0+boxH-1, '+', nil, border)
	for xx := 1; xx < boxW-1; xx++ {
		e.s.SetContent(x0+xx, y0, '-', nil, border)
		e.s.SetContent(x0+xx, y0+boxH-1, '-', nil, border)
	}
	for yy := 1; yy < boxH-1; yy++ {
		e.s.SetContent(x0, y0+yy, '|', nil, border)
		e.s.SetContent(x0+boxW-1, y0+yy, '|', nil, border)
	}

	title := " " + e.popupTitle + " "
	titleRunes := []rune(title)
	tx := x0 + (boxW-len(titleRunes))/2
	if tx < x0+1 {
		tx = x0 + 1
	}
	for i, r := range titleRunes {
		x := tx + i
		if x >= x0+boxW-1 {
			break
		}
		e.s.SetContent(x, y0, r, nil, border)
	}

	startY := y0 + 2
	startX := x0 + 2

	maxScroll := 0
	if len(lines) > visualH {
		maxScroll = len(lines) - visualH
	}
	e.popupScroll = clamp(e.popupScroll, 0, maxScroll)

	for i := 0; i < visualH; i++ {
		idx := e.popupScroll + i
		line := ""
		if idx >= 0 && idx < len(lines) {
			line = lines[idx]
		}

		runes := []rune(line)
		if len(runes) > contentW {
			if contentW > 1 {
				runes = runes[:contentW-1]
				runes = append(runes, '…')
			} else {
				runes = []rune{'…'}
			}
		}

		for j, r := range runes {
			e.s.SetContent(startX+j, startY+i, r, nil, textStyle)
		}
	}
}

func (e *Editor) previewText(s string, maxN int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	r := []rune(s)
	if len(r) <= maxN {
		return s
	}
	return string(r[:maxN-1]) + "…"
}
