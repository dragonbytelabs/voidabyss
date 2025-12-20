package editor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

type Editor struct {
	s         tcell.Screen
	lines     []string
	cx, cy    int
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
	pendingCount int
	pendingOp    rune // d / c / y
	yankBuf      []string
}

/*
====================
  Entry Points
====================
*/

func OpenFile(path string) error {
	ed, err := newEditorFromFile(path)
	if err != nil {
		return err
	}
	return ed.run()
}

func OpenProject(path string) error {
	ed, err := newEditorFromProject(path)
	if err != nil {
		return err
	}
	return ed.run()
}

func newEditorFromFile(path string) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var lines []string
	data, err := os.ReadFile(abs)
	if err != nil {
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

	return &Editor{
		s:        s,
		lines:    []string{""},
		mode:     ModeNormal,
		filename: abs,
	}, nil
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
			if ev.Key() == tcell.KeyCtrlQ {
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
  Rendering
====================
*/

func (e *Editor) draw() {
	e.s.Clear()
	w, h := e.s.Size()
	style := tcell.StyleDefault

	for y := 0; y < h-1; y++ {
		i := e.rowOffset + y
		if i >= len(e.lines) {
			break
		}
		r := []rune(e.lines[i])
		start := min(e.colOffset, len(r))
		for x := 0; x < w && x < len(r[start:]); x++ {
			e.s.SetContent(x, y, r[start+x], nil, style)
		}
	}

	e.drawStatus(w, h)

	x := e.cx - e.colOffset
	y := e.cy - e.rowOffset
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if y > h-2 {
		y = h - 2
	}
	e.s.ShowCursor(x, y)
	e.s.Show()
}

func (e *Editor) drawStatus(w, h int) {
	mode := map[Mode]string{
		ModeNormal:  "NORMAL",
		ModeInsert:  "INSERT",
		ModeCommand: "COMMAND",
	}[e.mode]

	left := fmt.Sprintf("%s  %s", mode, e.filename)
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

	for x, r := range bar {
		if x >= w {
			break
		}
		e.s.SetContent(x, h-1, r, nil, tcell.StyleDefault.Reverse(true))
	}

	if msg != "" {
		for i, r := range msg {
			x := min(len([]rune(left))+2+i, w-1)
			e.s.SetContent(x, h-1, r, nil, tcell.StyleDefault.Reverse(true))
		}
	}
}

/*
====================
  Input Handling
====================
*/

func (e *Editor) handleKey(k *tcell.EventKey) bool {
	if k.Key() == tcell.KeyEsc {
		e.mode = ModeNormal
		e.pendingCount = 0
		e.pendingOp = 0
		e.cmdBuf = nil
		e.statusMsg = ""
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

	if k.Key() == tcell.KeyRune {
		r := k.Rune()

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

		if e.pendingOp != 0 {
			e.applyOperator(e.pendingOp, r, n)
			e.pendingOp = 0
			return
		}

		switch r {
		case 'd', 'c', 'y':
			e.pendingOp = r
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
			e.moveRight(1)
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
			e.cmdBuf = nil
		case '$':
			e.cx = len([]rune(e.lines[e.cy]))
			e.wantX = e.cx
		}
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
		cmd := string(e.cmdBuf)
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
	default:
		e.statusMsg = "Not a command"
	}
	return false
}

/*
====================
  Buffer Ops
====================
*/

func (e *Editor) insertRune(r rune) {
	line := []rune(e.lines[e.cy])
	line = append(line[:e.cx], append([]rune{r}, line[e.cx:]...)...)
	e.lines[e.cy] = string(line)
	e.cx++
	e.wantX = e.cx
	e.dirty = true
}

func (e *Editor) backspace() {
	if e.cx > 0 {
		line := []rune(e.lines[e.cy])
		e.lines[e.cy] = string(append(line[:e.cx-1], line[e.cx:]...))
		e.cx--
		e.wantX = e.cx
		e.dirty = true
		return
	}
	if e.cy > 0 {
		prev := []rune(e.lines[e.cy-1])
		cur := []rune(e.lines[e.cy])
		e.lines[e.cy-1] = string(append(prev, cur...))
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		e.cy--
		e.cx = len(prev)
		e.wantX = e.cx
		e.dirty = true
	}
}

func (e *Editor) newline() {
	line := []rune(e.lines[e.cy])
	e.lines[e.cy] = string(line[:e.cx])
	e.lines = append(e.lines[:e.cy+1], append([]string{string(line[e.cx:])}, e.lines[e.cy+1:]...)...)
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
	e.lines[e.cy] = string(append(line[:e.cx], line[e.cx+1:]...))
	e.dirty = true
}

/*
====================
  Motions
====================
*/

func (e *Editor) moveUp(n int)    { e.cy = max(0, e.cy-n) }
func (e *Editor) moveDown(n int)  { e.cy = min(len(e.lines)-1, e.cy+n) }
func (e *Editor) moveLeft(n int)  { e.cx = max(0, e.cx-n); e.wantX = e.cx }
func (e *Editor) moveRight(n int) { e.cx = min(len([]rune(e.lines[e.cy])), e.cx+n); e.wantX = e.cx }

/*
====================
  Operators
====================
*/

func (e *Editor) applyOperator(op rune, motion rune, count int) {
	if motion == op {
		switch op {
		case 'd':
			e.deleteLines(count)
		case 'c':
			e.deleteLines(count)
			e.mode = ModeInsert
		case 'y':
			e.yankLines(count)
		}
		return
	}

	switch motion {
	case 'w':
		switch op {
		case 'd':
			e.deleteWord(count)
		case 'c':
			e.deleteWord(count)
			e.mode = ModeInsert
		case 'y':
			e.yankWord(count)
		}
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
	}
}

func (e *Editor) deleteLines(n int) {
	start := e.cy
	end := min(len(e.lines), e.cy+n)
	e.lines = append(e.lines[:start], e.lines[end:]...)
	if len(e.lines) == 0 {
		e.lines = []string{""}
	}
	e.cy = min(e.cy, len(e.lines)-1)
	e.cx = 0
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) yankLines(n int) {
	start := e.cy
	end := min(len(e.lines), e.cy+n)
	e.yankBuf = append([]string(nil), e.lines[start:end]...)
}

func (e *Editor) deleteWord(n int) {
	line := []rune(e.lines[e.cy])
	start := e.cx
	end := start
	for i := 0; i < n; i++ {
		for end < len(line) && !isWord(line[end]) {
			end++
		}
		for end < len(line) && isWord(line[end]) {
			end++
		}
	}
	e.lines[e.cy] = string(append(line[:start], line[end:]...))
	e.dirty = true
}

func (e *Editor) yankWord(n int) {
	line := []rune(e.lines[e.cy])
	start := e.cx
	end := start
	for i := 0; i < n; i++ {
		for end < len(line) && !isWord(line[end]) {
			end++
		}
		for end < len(line) && isWord(line[end]) {
			end++
		}
	}
	e.yankBuf = []string{string(line[start:end])}
}

func (e *Editor) deleteToBOL() {
	line := []rune(e.lines[e.cy])
	e.lines[e.cy] = string(line[e.cx:])
	e.cx = 0
	e.wantX = 0
	e.dirty = true
}

func (e *Editor) deleteToEOL() {
	line := []rune(e.lines[e.cy])
	e.lines[e.cy] = string(line[:e.cx])
	e.dirty = true
}

func (e *Editor) yankToBOL() {
	line := []rune(e.lines[e.cy])
	e.yankBuf = []string{string(line[:e.cx])}
}

func (e *Editor) yankToEOL() {
	line := []rune(e.lines[e.cy])
	e.yankBuf = []string{string(line[e.cx:])}
}

/*
====================
  Save & Helpers
====================
*/

func (e *Editor) save() {
	data := strings.Join(e.lines, "\n")
	_ = os.WriteFile(e.filename, []byte(data), 0644)
	e.dirty = false
}

func isWord(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}