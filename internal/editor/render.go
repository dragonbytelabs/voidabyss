package editor

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) draw() {
	e.s.Clear()
	w, h := e.s.Size()

	// Get color scheme
	scheme := GetColorScheme(e.config.ColorScheme)

	style := tcell.StyleDefault.Background(scheme.Background).Foreground(scheme.Foreground)
	highlightStyle := tcell.StyleDefault.Background(scheme.SearchBg).Foreground(scheme.Search)
	visualStyle := tcell.StyleDefault.Background(scheme.VisualBg).Foreground(scheme.Visual)
	treeStyle := tcell.StyleDefault.Foreground(scheme.TreeDirectory)
	treeCursorStyle := tcell.StyleDefault.Background(scheme.TreeCursorBg).Foreground(scheme.TreeCursor)
	treeBorderStyle := tcell.StyleDefault.Foreground(scheme.TreeBorder)
	lineNumStyle := tcell.StyleDefault.Foreground(scheme.LineNumber).Background(scheme.Background)

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

	// Calculate line number column width
	lineNumWidth := 0
	if e.config != nil && e.config.ShowLineNumbers {
		totalLines := e.lineCount()
		lineNumWidth = len(fmt.Sprintf("%d", totalLines)) + 2 // +2 for fold indicator and spacing
		if lineNumWidth < 5 {
			lineNumWidth = 5
		}
	}

	// Get syntax highlights for entire visible viewport once
	var highlights []Highlight
	bv := e.buf()
	if bv != nil && bv.parser != nil {
		// Calculate byte range for entire visible viewport
		firstLine := e.rowOffset
		lastLine := min(e.rowOffset+h-1, e.lineCount()-1)
		if firstLine < e.lineCount() && lastLine >= firstLine {
			startPos := e.lineStartPos(firstLine)
			endPos := e.buffer.Len()
			if lastLine < e.lineCount()-1 {
				endPos = e.lineStartPos(lastLine + 1)
			}
			highlights = bv.parser.GetHighlights(startPos, endPos)
		}
	}

	// Draw buffer content
	visualLine := 0
	actualLine := e.rowOffset
	for visualLine < h-1 && actualLine < e.lineCount() {
		// Skip folded lines
		if e.isLineFolded(actualLine) {
			actualLine++
			continue
		}

		lineIndex := actualLine

		// Draw line number with fold indicator
		if lineNumWidth > 0 {
			var lineNum string
			if e.config.RelativeLineNums {
				// Relative line numbers
				diff := lineIndex - e.cy
				if diff == 0 {
					lineNum = fmt.Sprintf("%*d", lineNumWidth-2, lineIndex+1)
				} else {
					if diff < 0 {
						diff = -diff
					}
					lineNum = fmt.Sprintf("%*d", lineNumWidth-2, diff)
				}
			} else {
				// Absolute line numbers
				lineNum = fmt.Sprintf("%*d", lineNumWidth-2, lineIndex+1)
			}

			// Draw line number
			for i, r := range lineNum {
				e.s.SetContent(contentStartX+i, visualLine, r, nil, lineNumStyle)
			}

			// Draw fold indicator
			foldIndicator := e.getFoldIndicator(lineIndex)
			e.s.SetContent(contentStartX+lineNumWidth-2, visualLine, []rune(foldIndicator)[0], nil, lineNumStyle)

			// Add spacing after line number
			e.s.SetContent(contentStartX+lineNumWidth-1, visualLine, ' ', nil, style)
		}

		runes := []rune(e.getLine(lineIndex))
		start := min(e.colOffset, len(runes))
		visible := runes[start:]

		// Calculate absolute position for highlight matching
		lineStartPos := e.lineStartPos(lineIndex)

		// Adjust content start and width for line numbers
		textStartX := contentStartX + lineNumWidth
		textWidth := contentWidth - lineNumWidth

		for x := 0; x < textWidth && x < len(visible); x++ {
			absPos := lineStartPos + start + x
			cellStyle := style

			// Check syntax highlighting first (lowest priority)
			if len(highlights) > 0 {
				syntaxStyle := e.getSyntaxStyle(absPos, highlights)
				if syntaxStyle != nil {
					cellStyle = *syntaxStyle
				}
			}

			// Check if this position is in visual selection (takes priority)
			if e.isInVisualSelection(absPos) {
				cellStyle = visualStyle
			} else if e.isSearchMatch(absPos) {
				// Check if this position is in a search match
				cellStyle = highlightStyle
			}

			e.s.SetContent(textStartX+x, visualLine, visible[x], nil, cellStyle)
		}

		visualLine++
		actualLine++
	}

	e.drawStatus(w, h)

	if e.popupActive {
		e.drawPopup(w, h)
	}

	// Calculate line number width for cursor positioning
	cursorLineNumWidth := 0
	if e.config != nil && e.config.ShowLineNumbers {
		totalLines := e.lineCount()
		cursorLineNumWidth = len(fmt.Sprintf("%d", totalLines)) + 2 // +2 for fold indicator and spacing
		if cursorLineNumWidth < 5 {
			cursorLineNumWidth = 5
		}
	}

	// Position cursor
	screenX := e.cx - e.colOffset + contentStartX + cursorLineNumWidth
	screenY := e.cy - e.rowOffset
	screenX = max(contentStartX+cursorLineNumWidth, screenX)
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

// getSyntaxStyle returns the appropriate style for a given position based on syntax highlighting
func (e *Editor) getSyntaxStyle(pos int, highlights []Highlight) *tcell.Style {
	// Find the most specific (smallest/innermost) highlight that contains this position
	// Search backwards since children are added after parents
	for i := len(highlights) - 1; i >= 0; i-- {
		hl := highlights[i]
		if pos >= hl.StartByte && pos < hl.EndByte {
			return e.highlightTypeToStyle(hl.Type)
		}
	}
	return nil
}

// highlightTypeToStyle converts a HighlightType to a tcell.Style
func (e *Editor) highlightTypeToStyle(hlType HighlightType) *tcell.Style {
	scheme := GetColorScheme(e.config.ColorScheme)
	var style tcell.Style
	switch hlType {
	case HighlightKeyword:
		style = tcell.StyleDefault.Foreground(scheme.Keyword).Bold(true)
	case HighlightFunction:
		style = tcell.StyleDefault.Foreground(scheme.Function)
	case HighlightType_:
		style = tcell.StyleDefault.Foreground(scheme.Type)
	case HighlightString:
		style = tcell.StyleDefault.Foreground(scheme.String)
	case HighlightNumber:
		style = tcell.StyleDefault.Foreground(scheme.Number)
	case HighlightComment:
		style = tcell.StyleDefault.Foreground(scheme.Comment).Dim(true)
	case HighlightConstant:
		style = tcell.StyleDefault.Foreground(scheme.Constant).Bold(true)
	case HighlightProperty:
		style = tcell.StyleDefault.Foreground(scheme.Property)
	default:
		return nil
	}
	return &style
}

func (e *Editor) previewText(s string, maxN int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	r := []rune(s)
	if len(r) <= maxN {
		return s
	}
	return string(r[:maxN-1]) + "…"
}
