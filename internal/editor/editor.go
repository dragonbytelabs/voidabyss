package editor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeVisual
)

type Editor struct {
	s         tcell.Screen
	lines     []string
	cx, cy    int
	rowOffset int
	colOffset int

	mode  Mode
	wantX int 

	// command-line (":")
	cmdBuf    []rune
	statusMsg string
	filename  string
	dirty     bool

	
	pendingCount int
	pendingOp    rune // 'd','c','y'
	yankBuf      []string
}

// --- constructors / entrypoints ---

func OpenFile(path string) error {
	ed, err := NewEditorFromFile(path)
	if err != nil {
		return err
	}
	return ed.Run()
}

func OpenProject(path string) error {
	ed, err := NewEditorFromProject(path)
	if err != nil {
		return err
	}
	return ed.Run()
}

func NewEditorFromFile(path string) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var lines []string
	data, err := os.ReadFile(abs)
	if err != nil {
		// If file doesn't exist, start empty but remember filename.
		lines = []string{""}
	} else {
		txt := strings.ReplaceAll(string(data), "\r\n", "\n")
		lines = strings.Split(txt, "\n")
		if len(lines) == 0 {
			lines = []string{""}
		}
	}

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}

	return &Editor{
		s:        s,
		lines:    lines,
		mode:     ModeNormal,
		filename: abs,
	}, nil
}

func NewEditorFromProject(dir string) (*Editor, error) {
	abs, err := filepath.Abs(dir)
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

	return &Editor{
		s:        s,
		lines:    []string{""},
		mode:     ModeNormal,
		filename: abs, // just show dir for now
	}, nil
}

func (e *Editor) Run() error {
	defer e.s.Fini()
	e.loop()
	return nil
}

// --- main loop ---

func (e *Editor) loop() {
	for {
		e.ensureCursorValid()
		e.ensureCursorVisible()
		e.draw()

		ev := e.s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ {
				return
			}
			if e.handleKey(ev) {
				return
			}
		case *tcell.EventResize:
			e.s.Sync()
		}
	}
}

func (e *Editor) ensureCursorValid() {
	if len(e.lines) == 0 {
		e.lines = []string{""}
	}
	if e.cy < 0 {
		e.cy = 0
	}
	if e.cy >= len(e.lines) {
		e.cy = len(e.lines) - 1
	}
	lineLen := len([]rune(e.lines[e.cy]))
	if e.cx < 0 {
		e.cx = 0
	}
	if e.cx > lineLen {
		e.cx = lineLen
	}
}

func (e *Editor) ensureCursorVisible() {
	w, h := e.s.Size()
	viewH := max(1, h-1) // last row is status

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

// --- drawing ---

func (e *Editor) draw() {
	e.s.Clear()
	w, h := e.s.Size()
	style := tcell.StyleDefault

	// buffer
	for y := 0; y < h-1; y++ {
		lineIndex := e.rowOffset + y
		if lineIndex >= len(e.lines) {
			break
		}
		runes := []rune(e.lines[lineIndex])
		start := min(e.colOffset, len(runes))
		visible := runes[start:]

		for x := 0; x < w && x < len(visible); x++ {
			e.s.SetContent(x, y, visible[x], nil, style)
		}
	}

	e.drawStatusLine(w, h, style)

	// cursor mapping
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

func (e *Editor) drawStatusLine(w, h int, style tcell.Style) {
	modeStr := map[Mode]string{
		ModeNormal:  "NORMAL",
		ModeInsert:  "INSERT",
		ModeCommand: "COMMAND",
		ModeVisual:  "VISUAL",
	}[e.mode]

	left := fmt.Sprintf("%s  %s", modeStr, e.filenameOrDefault())
	if e.dirty {
		left += " [+]"
	}

	right := fmt.Sprintf("  Pos: %d,%d", e.cx, e.cy)

	msg := e.statusMsg
	if e.mode == ModeNormal {
		// show pending count/op (like vim command echo)
		if e.pendingCount > 0 {
			msg = fmt.Sprintf("%d", e.pendingCount)
		}
		if e.pendingOp != 0 {
			if msg == "" {
				msg = string(e.pendingOp)
			} else {
				msg += string(e.pendingOp)
			}
		}
	}
	if e.mode == ModeCommand {
		msg = ":" + string(e.cmdBuf)
	}

	// build bar
	bar := left
	space := w - len([]rune(bar)) - len([]rune(right))
	if space < 1 {
		space = 1
	}
	bar = bar + strings.Repeat(" ", space) + right

	// background
	for x := 0; x < w; x++ {
		e.s.SetContent(x, h-1, ' ', nil, style.Reverse(true))
	}

	// text
	for x, r := range []rune(bar) {
		if x >= w {
			break
		}
		e.s.SetContent(x, h-1, r, nil, style.Reverse(true))
	}

	// overlay msg
	if msg != "" {
		startX := 0
		if e.mode != ModeCommand {
			startX = min(len([]rune(left))+2, w-1)
		}
		for i, r := range []rune(msg) {
			x := startX + i
			if x >= w {
				break
			}
			e.s.SetContent(x, h-1, r, nil, style.Reverse(true))
		}
	}
}

func (e *Editor) filenameOrDefault() string {
	if e.filename == "" {
		return "[No Name]"
	}
	return e.filename
}

// --- input dispatch ---

func (e *Editor) handleKey(k *tcell.EventKey) bool {
	// ESC: always return to normal and clear pending state
	if k.Key() == tcell.KeyEsc {
		e.mode = ModeNormal
		e.cmdBuf = e.cmdBuf[:0]
		e.statusMsg = ""
		e.pendingCount = 0
		e.pendingOp = 0
		return false
	}

	switch e.mode {
	case ModeCommand:
		return e.handleCommandKey(k)
	case ModeInsert:
		e.handleInsertKey(k)
	default:
		e.handleNormalKey(k)
	}
	return false
}

func (e *Editor) handleNormalKey(k *tcell.EventKey) {
	consumeCount := func() int {
		if e.pendingCount == 0 {
			return 1
		}
		c := e.pendingCount
		e.pendingCount = 0
		return c
	}

	switch k.Key() {
	case tcell.KeyUp:
		e.moveUp(consumeCount())
		return
	case tcell.KeyDown:
		e.moveDown(consumeCount())
		return
	case tcell.KeyLeft:
		e.moveLeft(consumeCount())
		return
	case tcell.KeyRight:
		e.moveRight(consumeCount())
		return
	case tcell.KeyRune:
		e.statusMsg = ""
		r := k.Rune()

		// digits build counts. BUT bare '0' is command (BOL) *if no count yet and no op pending*.
		if r >= '0' && r <= '9' {
			d := int(r - '0')
			if d == 0 && e.pendingCount == 0 && e.pendingOp == 0 {
				e.cx = 0
				e.wantX = e.cx
				return
			}
			e.pendingCount = e.pendingCount*10 + d
			return
		}

		n := consumeCount()

		// if operator pending, treat rune as motion
		if e.pendingOp != 0 {
			e.applyOperator(e.pendingOp, r, n)
			e.pendingOp = 0
			return
		}

		switch r {
		case 'd', 'c', 'y':
			e.pendingOp = r
			return

		case '$':
			e.cx = len([]rune(e.lines[e.cy]))
			e.wantX = e.cx

		case 'h':
			e.moveLeft(n)
		case 'j':
			e.moveDown(n)
		case 'k':
			e.moveUp(n)
		case 'l':
			e.moveRight(n)

		case 'i':
			e.mode = ModeInsert

		case 'a':
			lineLen := len([]rune(e.lines[e.cy]))
			if e.cx < lineLen {
				e.cx++
			}
			e.wantX = e.cx
			e.mode = ModeInsert

		case 'A':
			e.cx = len([]rune(e.lines[e.cy]))
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

		case ':':
			e.mode = ModeCommand
			e.cmdBuf = e.cmdBuf[:0]

		default:
			log.Printf("Unrecognized normal mode command: %q", r)
			e.pendingCount = 0
			e.pendingOp = 0
		}
	}
}

func (e *Editor) handleInsertKey(k *tcell.EventKey) {
	switch k.Key() {
	case tcell.KeyUp:
		e.moveUp(1)
	case tcell.KeyDown:
		e.moveDown(1)
	case tcell.KeyLeft:
		e.moveLeft(1)
	case tcell.KeyRight:
		e.moveRight(1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.backspace()
	case tcell.KeyEnter:
		e.newline()
	case tcell.KeyRune:
		e.insertRune(k.Rune())
	}
}

func (e *Editor) handleCommandKey(k *tcell.EventKey) bool {
	switch k.Key() {
	case tcell.KeyEnter:
		cmd := strings.TrimSpace(string(e.cmdBuf))
		e.cmdBuf = e.cmdBuf[:0]
		e.mode = ModeNormal
		return e.execCommand(cmd)

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.cmdBuf) > 0 {
			e.cmdBuf = e.cmdBuf[:len(e.cmdBuf)-1]
		}

	case tcell.KeyRune:
		e.cmdBuf = append(e.cmdBuf, k.Rune())
	}
	return false
}

// --- commands (:w etc.) ---

func (e *Editor) execCommand(cmd string) bool {
	switch cmd {
	case "q":
		if e.dirty {
			e.statusMsg = "No write since last change (use :q! to quit)"
			return false
		}
		return true
	case "q!":
		return true
	case "w":
		if err := e.save(); err != nil {
			e.statusMsg = "Write failed: " + err.Error()
		} else {
			e.statusMsg = "Wrote " + e.filenameOrDefault()
			e.dirty = false
		}
	case "wq":
		if err := e.save(); err != nil {
			e.statusMsg = "Write failed: " + err.Error()
			return false
		}
		return true
	default:
		e.statusMsg = "Not an editor command: " + cmd
	}
	return false
}

// --- buffer ops ---

func (e *Editor) insertRune(r rune) {
	line := []rune(e.lines[e.cy])
	if e.cx < 0 || e.cx > len(line) {
		e.ensureCursorValid()
		line = []rune(e.lines[e.cy])
	}
	line = append(line[:e.cx], append([]rune{r}, line[e.cx:]...)...)
	e.lines[e.cy] = string(line)
	e.cx++
	e.wantX = e.cx
	e.dirty = true
}

func (e *Editor) backspace() {
	if e.cx > 0 {
		line := []rune(e.lines[e.cy])
		line = append(line[:e.cx-1], line[e.cx:]...)
		e.lines[e.cy] = string(line)
		e.cx--
		e.wantX = e.cx
		e.dirty = true
		return
	}

	// merge with previous line
	if e.cy > 0 {
		prev := []rune(e.lines[e.cy-1])
		cur := []rune(e.lines[e.cy])
		newCx := len(prev)
		e.lines[e.cy-1] = string(append(prev, cur...))
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		e.cy--
		e.cx = newCx
		e.wantX = e.cx
		e.dirty = true
	}
}

func (e *Editor) newline() {
	line := []rune(e.lines[e.cy])
	left := string(line[:e.cx])
	right := string(line[e.cx:])
	e.lines[e.cy] = left
	e.lines = append(e.lines[:e.cy+1], append([]string{right}, e.lines[e.cy+1:]...)...)
	e.cy++
	e.cx = 0
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) deleteAtCursor() {
	line := []rune(e.lines[e.cy])
	if e.cx >= len(line) {
		return
	}
	line = append(line[:e.cx], line[e.cx+1:]...)
	e.lines[e.cy] = string(line)
	e.dirty = true
	e.ensureCursorValid()
}

func (e *Editor) openBelow() {
	e.lines = append(e.lines[:e.cy+1], append([]string{""}, e.lines[e.cy+1:]...)...)
	e.cy++
	e.cx = 0
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) openAbove() {
	e.lines = append(e.lines[:e.cy], append([]string{""}, e.lines[e.cy:]...)...)
	e.cx = 0
	e.wantX = 0
	e.dirty = true
}

// --- motions ---

func (e *Editor) moveUp(n int) {
	e.cy = max(0, e.cy-n)
	e.cx = min(e.wantX, len([]rune(e.lines[e.cy])))
}

func (e *Editor) moveDown(n int) {
	e.cy = min(len(e.lines)-1, e.cy+n)
	e.cx = min(e.wantX, len([]rune(e.lines[e.cy])))
}

func (e *Editor) moveLeft(n int) {
	e.cx = max(0, e.cx-n)
	e.wantX = e.cx
}

func (e *Editor) moveRight(n int) {
	lineLen := len([]rune(e.lines[e.cy]))
	e.cx = min(lineLen, e.cx+n)
	e.wantX = e.cx
}

// --- I/O ---

func (e *Editor) save() error {
	// if filename is a directory (project open), default to out.txt in that dir
	if e.filename == "" || e.filename == "[No Name]" {
		e.filename = "out.txt"
	} else {
		if info, err := os.Stat(e.filename); err == nil && info.IsDir() {
			e.filename = filepath.Join(e.filename, "out.txt")