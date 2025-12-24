package editor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dragonbytelabs/voidabyss/core/buffer"
	"github.com/gdamore/tcell/v2"
)

/*
====================
  Types & State
====================
*/

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeVisual
)

type RegisterKind int

const (
	RegCharwise RegisterKind = iota
	RegLinewise
	RegBlockwise
)

type Register struct {
	kind RegisterKind
	text string
}

type Registers struct {
	unnamed  Register          // "
	numbered [10]Register      // 0-9
	named    map[rune]Register // a-z
}

type Editor struct {
	s      tcell.Screen
	buffer *buffer.Buffer

	// cursor in (line, col) where col is rune offset within the line (not absolute)
	cx, cy int

	// viewport offsets in (line, col)
	rowOffset int
	colOffset int

	mode  Mode
	wantX int

	// command mode
	cmdBuf    []rune
	statusMsg string
	filename  string
	dirty     bool

	// vim-ish state
	pendingCount     int
	pendingOp        rune // d / c / y
	regs             Registers
	awaitingRegister bool
	regOverride      rune
	regOverrideSet   bool

	// popup UI
	popupActive bool
	popupTitle  string
	popupLines  []string
	popupScroll int
	popupFixedH int
}

/*
====================
  Constructors
====================
*/

func newEditorFromFile(path string) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	txt := ""
	data, readErr := os.ReadFile(abs)
	if readErr == nil {
		txt = strings.ReplaceAll(string(data), "\r\n", "\n")
	} else {
		// missing file is okay; start empty
		txt = ""
	}

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}

	ed := &Editor{
		s:        s,
		buffer:   buffer.NewFromString(txt),
		mode:     ModeNormal,
		filename: abs,
	}

	ed.regs.named = make(map[rune]Register)

	if readErr != nil && !os.IsNotExist(readErr) {
		ed.statusMsg = "read failed: " + readErr.Error()
	} else if readErr != nil && os.IsNotExist(readErr) {
		ed.statusMsg = "new file"
	}

	return ed, nil
}

func newEditorFromProject(path string) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}

	ed := &Editor{
		s:        s,
		buffer:   buffer.NewFromString(""),
		mode:     ModeNormal,
		filename: abs, // show dir in statusline for now
	}
	ed.regs.named = make(map[rune]Register)

	return ed, nil
}

/*
====================
  Main Loop
====================
*/

func (e *Editor) run() error {
	defer e.s.Fini()

	for {
		e.ensureCursorValid()
		e.ensureCursorVisible()
		e.draw()

		ev := e.s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ { // emergency quit
				return nil
			}
			if e.handleKey(ev) {
				return nil
			}
		case *tcell.EventResize:
			e.s.Sync()
		}
	}
}

/*
====================
  Line Model (recomputed)
====================
*/

func (e *Editor) textRunes() []rune {
	return []rune(e.buffer.String())
}

// lineStarts returns rune offsets of each line start (0-based). Always includes 0.
func (e *Editor) lineStarts() []int {
	r := e.textRunes()
	starts := []int{0}
	for i, ch := range r {
		if ch == '\n' {
			if i+1 <= len(r) {
				starts = append(starts, i+1)
			}
		}
	}
	return starts
}

func (e *Editor) lineCount() int {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return 1
	}
	return len(starts)
}

// getLine returns the line text for line index y (without trailing newline).
func (e *Editor) getLine(y int) string {
	r := e.textRunes()
	starts := e.lineStarts()
	if len(starts) == 0 {
		return ""
	}

	if y < 0 {
		y = 0
	}
	if y >= len(starts) {
		y = len(starts) - 1
	}

	start := starts[y]
	end := len(r)
	if y+1 < len(starts) {
		end = starts[y+1] - 1 // exclude '\n'
		if end < start {
			end = start
		}
	}

	return string(r[start:end])
}

func (e *Editor) lineLen(y int) int {
	return len([]rune(e.getLine(y)))
}

// posFromCursor converts (cy,cx) to absolute rune offset into the buffer.
func (e *Editor) posFromCursor() int {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return 0
	}

	if e.cy < 0 {
		e.cy = 0
	}
	if e.cy >= len(starts) {
		e.cy = len(starts) - 1
	}

	lineStart := starts[e.cy]
	ll := e.lineLen(e.cy)

	if e.cx < 0 {
		e.cx = 0
	}
	if e.cx > ll {
		e.cx = ll
	}

	return lineStart + e.cx
}

// setCursorFromPos sets (cy,cx) from absolute rune offset.
func (e *Editor) setCursorFromPos(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > e.buffer.Len() {
		pos = e.buffer.Len()
	}

	starts := e.lineStarts()
	if len(starts) == 0 {
		e.cy, e.cx = 0, 0
		return
	}

	// Find last start <= pos
	lo, hi := 0, len(starts)-1
	best := 0
	for lo <= hi {
		m := (lo + hi) / 2
		if starts[m] <= pos {
			best = m
			lo = m + 1
		} else {
			hi = m - 1
		}
	}

	e.cy = best
	e.cx = pos - starts[best]

	ll := e.lineLen(e.cy)
	if e.cx > ll {
		e.cx = ll
	}
}

/*
====================
  Cursor / Viewport
====================
*/

func (e *Editor) ensureCursorValid() {
	lc := e.lineCount()
	if e.cy < 0 {
		e.cy = 0
	}
	if e.cy >= lc {
		e.cy = lc - 1
	}

	ll := e.lineLen(e.cy)
	if e.cx < 0 {
		e.cx = 0
	}
	if e.cx > ll {
		e.cx = ll
	}
}

func (e *Editor) ensureCursorVisible() {
	w, h := e.s.Size()
	viewH := max(1, h-1)

	if e.cy < e.rowOffset {
		e.rowOffset = e.cy
	} else if e.cy >= e.rowOffset+viewH {
		e.rowOffset = e.cy - viewH + 1
	}

	if e.cx < e.colOffset {
		e.colOffset = e.cx
	} else if e.cx >= e.colOffset+w {
		e.colOffset = e.cx - w + 1
	}

	if e.rowOffset < 0 {
		e.rowOffset = 0
	}
	if e.colOffset < 0 {
		e.colOffset = 0
	}
}

/*
====================
  Rendering
====================
*/

func (e *Editor) draw() {
	e.s.Clear()
	w, h := e.s.Size()
	style := tcell.StyleDefault

	for y := 0; y < h-1; y++ {
		lineIndex := e.rowOffset + y
		if lineIndex >= e.lineCount() {
			break
		}

		runes := []rune(e.getLine(lineIndex))
		start := min(e.colOffset, len(runes))
		visible := runes[start:]
		for x := 0; x < w && x < len(visible); x++ {
			e.s.SetContent(x, y, visible[x], nil, style)
		}
	}

	e.drawStatus(w, h)

	if e.popupActive {
		e.drawPopup(w, h)
	}

	screenX := e.cx - e.colOffset
	screenY := e.cy - e.rowOffset
	if screenX < 0 {
		screenX = 0
	}
	if screenY < 0 {
		screenY = 0
	}
	if screenY > h-2 {
		screenY = h - 2
	}

	e.s.ShowCursor(screenX, screenY)
	e.s.Show()
}

func (e *Editor) drawStatus(w, h int) {
	mode := map[Mode]string{
		ModeNormal:  "NORMAL",
		ModeInsert:  "INSERT",
		ModeCommand: "COMMAND",
		ModeVisual:  "VISUAL",
	}[e.mode]

	regCh := rune('"')
	if e.regOverrideSet {
		regCh = e.regOverride
	}
	left := fmt.Sprintf("%s  %s  reg:%c", mode, e.filename, regCh)

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
	}
	if e.mode == ModeCommand {
		msg = ":" + string(e.cmdBuf)
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

func (e *Editor) drawPopup(w, h int) {
	if !e.popupActive {
		return
	}

	lines := e.popupLines
	if lines == nil {
		lines = []string{}
	}

	// Determine popup width from title + content
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

	// clamp width to screen
	if contentW > w-6 {
		contentW = w - 6
	}
	if contentW < 20 {
		contentW = 20
	}

	// How many lines we can actually DISPLAY (no indexing beyond len(lines))
	displayH := len(lines)
	if displayH > h-6 {
		displayH = h - 6
	}
	if displayH < 1 {
		displayH = 1
	}

	minVisualH := 3
	visualH := displayH

	// FORCE fixed height if set
	if e.popupFixedH > 0 {
		visualH = e.popupFixedH
	}

	// still clamp to screen and minimum
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

	// Clear box area
	for yy := 0; yy < boxH; yy++ {
		for xx := 0; xx < boxW; xx++ {
			e.s.SetContent(x0+xx, y0+yy, ' ', nil, border)
		}
	}

	// Border
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

	// Title
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

	// Content
	startY := y0 + 2
	startX := x0 + 2

	maxScroll := 0
	if len(lines) > visualH {
		maxScroll = len(lines) - visualH
	}
	if e.popupScroll < 0 {
		e.popupScroll = 0
	}
	if e.popupScroll > maxScroll {
		e.popupScroll = maxScroll
	}

	for i := 0; i < visualH; i++ {
		idx := e.popupScroll + i

		var line string
		if idx >= 0 && idx < len(lines) {
			line = lines[idx]
		} else {
			line = ""
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

// formatRegisters returns human-friendly lines for :reg popup.
func (e *Editor) formatRegisters() []string {
	lines := make([]string, 0, 32)

	// unnamed
	lines = append(lines, fmt.Sprintf("\"  %s", e.formatRegValue(e.regs.unnamed)))

	// numbered 0-9
	for i := 0; i <= 9; i++ {
		name := rune('0' + i)
		r := e.regs.numbered[i]
		if r.text == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%c  %s", name, e.formatRegValue(r)))
	}

	// named a-z (only those present)
	for ch := 'a'; ch <= 'z'; ch++ {
		r, ok := e.regs.named[ch]
		if !ok || r.text == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%c  %s", ch, e.formatRegValue(r)))
	}

	if len(lines) == 0 {
		return []string{"(no registers)"}
	}
	return lines
}

func (e *Editor) formatRegValue(r Register) string {
	if r.text == "" {
		return "(empty)"
	}

	kind := "char"
	switch r.kind {
	case RegLinewise:
		kind = "line"
	case RegBlockwise:
		kind = "block"
	}

	preview := e.previewText(r.text, 60)
	return fmt.Sprintf("[%s] %s", kind, preview)
}

func (e *Editor) previewText(s string, max int) string {
	// make newlines visible in preview
	s = strings.ReplaceAll(s, "\n", "\\n")
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

/*
====================
  Input Handling
====================
*/

func (e *Editor) handleKey(k *tcell.EventKey) bool {
	if e.popupActive {
		switch k.Key() {
		case tcell.KeyEsc, tcell.KeyEnter:
			e.closePopup()
			return false
		case tcell.KeyUp:
			e.popupScroll--
			return false
		case tcell.KeyDown:
			e.popupScroll++
			return false
		case tcell.KeyRune:
			switch k.Rune() {
			case 'q':
				e.closePopup()
			case 'k':
				e.popupScroll--
			case 'j':
				e.popupScroll++
			}
			return false
		}
		return false
	}

	if k.Key() == tcell.KeyEsc {
		e.mode = ModeNormal
		e.pendingCount = 0
		e.pendingOp = 0
		e.cmdBuf = nil
		e.statusMsg = ""

		e.awaitingRegister = false
		e.regOverrideSet = false
		e.regOverride = 0

		return false
	}

	switch e.mode {
	case ModeNormal:
		e.handleNormal(k)
	case ModeInsert:
		e.handleInsert(k)
	case ModeCommand:
		return e.handleCommand(k)
	}
	return false
}

func (e *Editor) handleNormal(k *tcell.EventKey) {
	consume := func() int {
		if e.pendingCount == 0 {
			return 1
		}
		n := e.pendingCount
		e.pendingCount = 0
		return n
	}

	switch k.Key() {
	case tcell.KeyUp:
		e.moveUp(consume())
		return
	case tcell.KeyDown:
		e.moveDown(consume())
		return
	case tcell.KeyLeft:
		e.moveLeft(consume())
		return
	case tcell.KeyRight:
		e.moveRight(consume())
		return
	case tcell.KeyCtrlR:
		e.redo()
		return
	case tcell.KeyRune:
		// continue below
	default:
		return
	}

	e.statusMsg = ""
	r := k.Rune()

	if e.awaitingRegister {
		if isRegisterName(r) {
			e.regOverride = r
			e.regOverrideSet = true
			e.statusMsg = ""
		} else {
			e.statusMsg = fmt.Sprintf("invalid register: %c", r)
			e.regOverrideSet = false
			e.regOverride = 0
		}
		e.awaitingRegister = false
		return
	}

	// counts
	if r >= '0' && r <= '9' {
		d := int(r - '0')
		if d == 0 && e.pendingCount == 0 && e.pendingOp == 0 {
			e.cx = 0
			e.wantX = 0
			return
		}
		e.pendingCount = e.pendingCount*10 + d
		return
	}

	n := consume()

	// operator pending
	if e.pendingOp != 0 {
		e.applyOperator(e.pendingOp, r, n)
		e.pendingOp = 0
		return
	}

	switch r {
	case 'd', 'c', 'y', 'g':
		e.pendingOp = r
		return

	case 'h':
		e.moveLeft(n)
	case 'j':
		e.moveDown(n)
	case 'k':
		e.moveUp(n)
	case 'l':
		e.moveRight(n)
	case 'p':
		e.pasteAfter()
	case 'P':
		e.pasteBefore()
	case 'i':
		e.mode = ModeInsert
	case 'a':
		e.moveRight(1)
		e.mode = ModeInsert
	case 'A':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx
		e.mode = ModeInsert

	case 'x':
		for i := 0; i < n; i++ {
			e.deleteAtCursor()
		}

	case 'o':
		e.openBelow()
		e.mode = ModeInsert
	case 'O':
		e.openAbove()
		e.mode = ModeInsert
	case 'u':
		e.undo()
	case ':':
		e.mode = ModeCommand
		e.cmdBuf = nil

	case '$':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx

	case '"':
		e.awaitingRegister = true
		return

	case 'G':
		e.cy = e.lineCount() - 1
		e.cx = 0
		e.wantX = e.cx

	default:
		log.Printf("unknown normal key: %q", r)
		e.pendingCount = 0
		e.pendingOp = 0
	}
}

func (e *Editor) handleInsert(k *tcell.EventKey) {
	switch k.Key() {
	case tcell.KeyRune:
		e.insertRune(k.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.backspace()
	case tcell.KeyEnter:
		e.newline()
	case tcell.KeyUp:
		e.moveUp(1)
	case tcell.KeyDown:
		e.moveDown(1)
	case tcell.KeyLeft:
		e.moveLeft(1)
	case tcell.KeyRight:
		e.moveRight(1)
	}
}

func (e *Editor) handleCommand(k *tcell.EventKey) bool {
	switch k.Key() {
	case tcell.KeyEnter:
		cmd := strings.TrimSpace(string(e.cmdBuf))
		e.cmdBuf = nil
		e.mode = ModeNormal
		return e.exec(cmd)

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.cmdBuf) > 0 {
			e.cmdBuf = e.cmdBuf[:len(e.cmdBuf)-1]
		}

	case tcell.KeyRune:
		e.cmdBuf = append(e.cmdBuf, k.Rune())
	}
	return false
}

/*
====================
  Commands
====================
*/

func (e *Editor) exec(cmd string) bool {
	switch cmd {
	case "q":
		if e.dirty {
			e.statusMsg = "No write since last change"
			return false
		}
		return true
	case "q!":
		return true
	case "w":
		e.save()
	case "wq":
		e.save()
		return true
	case "reg", "registers":
		e.popupFixedH = 10
		e.openPopup("REGISTERS", e.formatRegisters())
		return false
	default:
		e.statusMsg = "Not a command: " + cmd
	}
	return false
}

/*
====================
  Buffer Ops (piece table)
====================
*/

func (e *Editor) insertRune(r rune) {
	pos := e.posFromCursor()
	_ = e.buffer.Insert(pos, string(r))
	e.setCursorFromPos(pos + 1)
	e.wantX = e.cx
	e.dirty = true
}

func (e *Editor) backspace() {
	pos := e.posFromCursor()
	if pos == 0 {
		return
	}
	_ = e.buffer.Delete(pos-1, pos)
	e.setCursorFromPos(pos - 1)
	e.wantX = e.cx
	e.dirty = true
}

func (e *Editor) newline() {
	pos := e.posFromCursor()
	_ = e.buffer.Insert(pos, "\n")
	e.setCursorFromPos(pos + 1)
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) deleteAtCursor() {
	pos := e.posFromCursor()
	if pos >= e.buffer.Len() {
		return
	}

	// capture deleted rune
	deleted, _ := e.buffer.Slice(pos, pos+1)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	e.statusMsg = "deleted"

	_ = e.buffer.Delete(pos, pos+1)
	e.setCursorFromPos(pos)
	e.dirty = true
}

func (e *Editor) undo() {
	if e.buffer == nil {
		return
	}
	if ok := e.buffer.Undo(); !ok {
		e.statusMsg = "Already at oldest change"
		return
	}

	// Keep cursor valid after content changed.
	pos := e.posFromCursor()
	if pos > e.buffer.Len() {
		pos = e.buffer.Len()
	}
	e.setCursorFromPos(pos)
	e.wantX = e.cx

	e.dirty = true
	e.statusMsg = "undo"

	// clear pending state
	e.pendingCount = 0
	e.pendingOp = 0
}

func (e *Editor) redo() {
	if e.buffer == nil {
		return
	}
	if ok := e.buffer.Redo(); !ok {
		e.statusMsg = "Already at newest change"
		return
	}

	pos := e.posFromCursor()
	if pos > e.buffer.Len() {
		pos = e.buffer.Len()
	}
	e.setCursorFromPos(pos)
	e.wantX = e.cx

	e.dirty = true
	e.statusMsg = "redo"

	e.pendingCount = 0
	e.pendingOp = 0
}

/*
====================
  Motions
====================
*/

func (e *Editor) moveUp(n int) {
	e.cy = max(0, e.cy-n)
	e.cx = min(e.wantX, e.lineLen(e.cy))
}

func (e *Editor) moveDown(n int) {
	e.cy = min(e.lineCount()-1, e.cy+n)
	e.cx = min(e.wantX, e.lineLen(e.cy))
}

func (e *Editor) moveLeft(n int) {
	e.cx = max(0, e.cx-n)
	e.wantX = e.cx
}

func (e *Editor) moveRight(n int) {
	e.cx = min(e.lineLen(e.cy), e.cx+n)
	e.wantX = e.cx
}

func (e *Editor) openBelow() {
	// insert newline at end of current line
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := lineStart + e.lineLen(e.cy)
	_ = e.buffer.Insert(pos, "\n")
	e.setCursorFromPos(pos + 1)
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) openAbove() {
	// insert newline at start of current line
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	_ = e.buffer.Insert(lineStart, "\n")
	e.setCursorFromPos(lineStart)
	e.wantX = 0
	e.dirty = true
}

/*
====================
  Operators (minimal / line-local)
====================
*/

func (e *Editor) applyOperator(op rune, motion rune, count int) {
	// linewise: dd/cc/yy
	if motion == op {
		switch op {
		case 'd':
			e.deleteLines(count)
			return
		case 'c':
			e.deleteLines(count)
			e.mode = ModeInsert
			return
		case 'y':
			e.yankLines(count)
			return
		case 'g':
			e.cy = 0
			e.cx = 0
			e.wantX = 0
			return
		}
	}

	switch motion {
	case '0':
		if op == 'd' || op == 'c' {
			e.deleteToBOL()
			if op == 'c' {
				e.mode = ModeInsert
			}
		} else if op == 'y' {
			e.yankToBOL()
		}
	case '$':
		if op == 'd' || op == 'c' {
			e.deleteToEOL()
			if op == 'c' {
				e.mode = ModeInsert
			}
		} else if op == 'y' {
			e.yankToEOL()
		}
	case 'w':
		if op == 'd' || op == 'c' {
			e.deleteWord(count)
			if op == 'c' {
				e.mode = ModeInsert
			}
		} else if op == 'y' {
			e.yankWord(count)
		}

	default:
		e.statusMsg = fmt.Sprintf("Unsupported motion %q for %q (for now)", motion, op)
		e.regOverrideSet = false
		e.regOverride = 0
	}
}

func (e *Editor) lineStartPos(y int) int {
	starts := e.lineStarts()
	if y < 0 {
		y = 0
	}
	if y >= len(starts) {
		return 0
	}
	return starts[y]
}

func (e *Editor) lineEndPos(y int) int {
	return e.lineStartPos(y) + e.lineLen(y)
}

func (e *Editor) setRegister(name rune, r Register) {
	// normalize "no explicit register" -> unnamed
	if name == 0 {
		name = '"'
	}

	// Append behavior for A-Z (optional but recommended)
	if name >= 'A' && name <= 'Z' {
		lower := name + ('a' - 'A')
		prev, _ := e.getRegister(lower)

		// append text; preserve kind if possible
		kind := r.kind
		if prev.text != "" && prev.kind == r.kind {
			kind = prev.kind
		}
		e.regs.named[lower] = Register{kind: kind, text: prev.text + r.text}

		// unnamed mirrors last written
		e.regs.unnamed = e.regs.named[lower]
		return
	}

	switch {
	case name == '"':
		e.regs.unnamed = r
	case name >= '0' && name <= '9':
		e.regs.numbered[name-'0'] = r
	case name >= 'a' && name <= 'z':
		e.regs.named[name] = r
	default:
		// ignore invalid names
		return
	}
}

func (e *Editor) getRegister(name rune) (Register, bool) {
	if name == 0 {
		name = '"'
	}
	// treat A-Z as a-z for reads
	if name >= 'A' && name <= 'Z' {
		name = name + ('a' - 'A')
	}

	switch {
	case name == '"':
		return e.regs.unnamed, e.regs.unnamed.text != ""
	case name >= '0' && name <= '9':
		r := e.regs.numbered[name-'0']
		return r, r.text != ""
	case name >= 'a' && name <= 'z':
		r, ok := e.regs.named[name]
		return r, ok && r.text != ""
	default:
		return Register{}, false
	}
}

func (e *Editor) writeYank(r Register) {
	target := e.consumeRegister()
	e.setRegister(target, r)
	// yank register 0 gets yanks
	e.regs.numbered[0] = r

	if stored, ok := e.getRegister(target); ok && stored.text != "" {
		e.regs.unnamed = stored
	} else {
		e.regs.unnamed = r
	}
}

func (e *Editor) writeDelete(r Register) {
	target := e.consumeRegister()
	e.setRegister(target, r)

	// shift 9..2
	for i := 9; i >= 2; i-- {
		e.regs.numbered[i] = e.regs.numbered[i-1]
	}
	e.regs.numbered[1] = r

	// unnamed mirrors last delete too
	if stored, ok := e.getRegister(target); ok && stored.text != "" {
		e.regs.unnamed = stored
	} else {
		e.regs.unnamed = r
	}
}

func (e *Editor) readPaste() (Register, bool) {
	name := e.consumeRegister()
	return e.getRegister(name)
}

func (e *Editor) deleteLines(n int) {
	if n <= 0 {
		return
	}
	starts := e.lineStarts()
	if len(starts) == 0 {
		return
	}

	startLine := e.cy
	endLine := min(e.lineCount(), e.cy+n)

	startPos := starts[startLine]
	var endPos int
	if endLine >= e.lineCount() {
		endPos = e.buffer.Len()
	} else {
		endPos = starts[endLine]
	}

	deleted, _ := e.buffer.Slice(startPos, endPos)
	e.writeDelete(Register{kind: RegLinewise, text: deleted})
	e.statusMsg = "deleted"

	_ = e.buffer.Delete(startPos, endPos)
	e.setCursorFromPos(startPos)
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) yankLines(n int) {
	starts := e.lineStarts()
	if len(starts) == 0 {
		return
	}

	startLine := e.cy
	endLine := min(e.lineCount(), e.cy+n)

	startPos := starts[startLine]
	endPos := e.buffer.Len()
	if endLine < e.lineCount() {
		endPos = starts[endLine]
	}

	// remove trailing newline for yank if present
	if endPos > startPos {
		lastChar, _ := e.buffer.Slice(endPos-1, endPos)
		if lastChar == "\n" {
			endPos--
		}
	}

	s, _ := e.buffer.Slice(startPos, endPos)
	e.writeYank(Register{kind: RegLinewise, text: s})
	e.statusMsg = "yanked"
}

func (e *Editor) pasteAfter() {
	reg, ok := e.readPaste()
	if !ok || reg.text == "" {
		e.statusMsg = "nothing to paste"
		return
	}

	switch reg.kind {
	case RegLinewise:
		// paste as new line(s) BELOW current line
		insertPos := e.lineEndPos(e.cy)
		// ensure we start on a new line
		if insertPos < e.buffer.Len() {
			_ = e.buffer.Insert(insertPos, "\n")
			insertPos++
		} else {
			// EOF: still create a newline if file not empty and last char not newline
			if e.buffer.Len() > 0 {
				// if last char isn't '\n', add one
				last, _ := e.buffer.Slice(e.buffer.Len()-1, e.buffer.Len())
				if last != "\n" {
					_ = e.buffer.Insert(e.buffer.Len(), "\n")
					insertPos = e.buffer.Len()
				}
			}
		}

		_ = e.buffer.Insert(insertPos, reg.text+"\n")
		e.setCursorFromPos(insertPos)
		e.wantX = 0

	default: // charwise
		pos := e.posFromCursor()
		insertPos := pos
		if insertPos < e.buffer.Len() {
			insertPos = pos + 1
		}
		_ = e.buffer.Insert(insertPos, reg.text)
		e.setCursorFromPos(insertPos + len([]rune(reg.text)) - 1)
		e.wantX = e.cx
	}

	e.dirty = true
	e.statusMsg = "pasted"
}

func (e *Editor) pasteBefore() {
	reg, ok := e.readPaste()
	if !ok || reg.text == "" {
		e.statusMsg = "nothing to paste"
		return
	}

	switch reg.kind {
	case RegLinewise:
		// paste as new line(s) ABOVE current line
		insertPos := e.lineStartPos(e.cy)
		_ = e.buffer.Insert(insertPos, reg.text+"\n")
		e.setCursorFromPos(insertPos)
		e.wantX = 0

	default: // charwise
		pos := e.posFromCursor()
		_ = e.buffer.Insert(pos, reg.text)
		e.setCursorFromPos(pos)
		e.wantX = e.cx
	}

	e.dirty = true
	e.statusMsg = "pasted"
}

func (e *Editor) deleteToBOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	if pos > lineStart {
		deleted, _ := e.buffer.Slice(lineStart, pos)
		e.writeDelete(Register{kind: RegCharwise, text: deleted})
		e.statusMsg = "deleted"
		_ = e.buffer.Delete(lineStart, pos)
		e.setCursorFromPos(lineStart)
		e.wantX = 0
		e.dirty = true
	}
}

func (e *Editor) deleteToEOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	eol := lineStart + e.lineLen(e.cy)
	if pos < eol {
		deleted, _ := e.buffer.Slice(pos, eol)
		e.writeDelete(Register{kind: RegCharwise, text: deleted})
		e.statusMsg = "deleted"
		_ = e.buffer.Delete(pos, eol)
		e.setCursorFromPos(pos)
		e.dirty = true
	}
}

func (e *Editor) yankToBOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	s, _ := e.buffer.Slice(lineStart, pos)
	e.writeYank(Register{kind: RegCharwise, text: s})
	e.statusMsg = "yanked"
}

func (e *Editor) yankToEOL() {
	starts := e.lineStarts()
	lineStart := starts[e.cy]
	pos := e.posFromCursor()
	eol := lineStart + e.lineLen(e.cy)
	s, _ := e.buffer.Slice(pos, eol)
	e.writeYank(Register{kind: RegCharwise, text: s})
	e.statusMsg = "yanked"
}

func (e *Editor) deleteWord(count int) {
	pos := e.posFromCursor()
	_, end := e.wordRangeFrom(pos, count)

	if end <= pos {
		e.statusMsg = "nothing to delete"
		return
	}

	deleted, _ := e.buffer.Slice(pos, end)
	e.writeDelete(Register{kind: RegCharwise, text: deleted})
	_ = e.buffer.Delete(pos, end)

	e.setCursorFromPos(pos)
	e.wantX = e.cx
	e.dirty = true
	e.statusMsg = "deleted"
}

func (e *Editor) yankWord(count int) {
	pos := e.posFromCursor()
	_, end := e.wordRangeFrom(pos, count)

	if end <= pos {
		e.statusMsg = "nothing to yank"
		return
	}

	yanked, _ := e.buffer.Slice(pos, end)
	e.writeYank(Register{kind: RegCharwise, text: yanked})
	e.statusMsg = "yanked"
}

func (e *Editor) wordRangeFrom(pos int, count int) (int, int) {
	if count <= 0 {
		count = 1
	}

	runes := e.textRunes()
	if pos < 0 {
		pos = 0
	}
	if pos >= len(runes) {
		return pos, pos
	}

	start := pos
	end := pos

	for i := 0; i < count; i++ {
		// Step 1: if we're not on a word char, skip forward to the next word char
		for start < len(runes) && !isWordChar(runes[start]) {
			start++
		}
		end = start

		// Step 2: consume word chars
		for end < len(runes) && isWordChar(runes[end]) {
			end++
		}

		// prepare next iteration: continue from end
		start = end

		if start >= len(runes) {
			break
		}
	}

	// For operator usage: we want deletion to begin at original pos, not the advanced `start`.
	// But we DO want the "end" computed above.
	// So return (pos, end), not (start, end).
	return pos, end
}

func (e *Editor) consumeRegister() rune {
	if e.regOverrideSet {
		r := e.regOverride
		e.regOverrideSet = false
		e.regOverride = 0
		return r
	}
	return '"'
}

/*
====================
  Save & Helpers
====================
*/

func (e *Editor) save() {
	outPath := e.filename
	if info, err := os.Stat(outPath); err == nil && info.IsDir() {
		outPath = filepath.Join(outPath, "out.txt")
	}

	_ = os.WriteFile(outPath, []byte(e.buffer.String()), 0644)
	e.dirty = false
	e.statusMsg = "written"
}

func isRegisterName(r rune) bool {
	if r == '"' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	return false
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}
