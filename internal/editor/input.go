package editor

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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

	// Ctrl+R redo (support multiple terminal encodings)
	if e.mode == ModeNormal && isCtrlR(k) {
		e.redo()
		return false
	}

	if k.Key() == tcell.KeyEsc {
		if e.mode == ModeVisual {
			e.visualExit()
			e.clearPending()
			return false
		}

		// End undo group when leaving insert mode
		if e.mode == ModeInsert {
			e.buffer.EndUndoGroup()
		}

		e.mode = ModeNormal
		e.clearPending()
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
	case ModeVisual:
		e.handleVisual(k)
	}
	return false
}

func (e *Editor) clearPending() {
	e.pendingCount = 0
	e.pendingOp = 0
	e.pendingOpCount = 0
	e.pendingTextObj = 0
}

func (e *Editor) consumeCountOr1() int {
	if e.pendingCount == 0 {
		return 1
	}
	n := e.pendingCount
	e.pendingCount = 0
	return n
}

func (e *Editor) handleNormal(k *tcell.EventKey) {
	// arrows still work with counts
	if k.Key() == tcell.KeyUp {
		e.moveUp(e.consumeCountOr1())
		return
	}
	if k.Key() == tcell.KeyDown {
		e.moveDown(e.consumeCountOr1())
		return
	}
	if k.Key() == tcell.KeyLeft {
		e.moveLeft(e.consumeCountOr1())
		return
	}
	if k.Key() == tcell.KeyRight {
		e.moveRight(e.consumeCountOr1())
		return
	}
	// Ctrl+R is handled earlier in handleKey() to avoid state issues
	if k.Key() != tcell.KeyRune {
		return
	}

	e.statusMsg = ""
	r := k.Rune()

	// register selection
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

	// if we’re waiting for iw/aw unit
	if e.pendingTextObj != 0 {
		op := e.pendingOp
		cnt := e.pendingOpCount
		prefix := e.pendingTextObj
		e.pendingTextObj = 0
		e.pendingOp = 0
		e.pendingOpCount = 0

		if r == 'w' || r == 'W' {
			e.applyOperatorTextObject(op, prefix, r, cnt)
		} else {
			e.statusMsg = "unsupported text object"
		}
		return
	}

	// counts
	if r >= '0' && r <= '9' {
		d := int(r - '0')
		// Vim-ish: leading 0 with no count/operator -> BOL
		if d == 0 && e.pendingCount == 0 && e.pendingOp == 0 {
			e.cx = 0
			e.wantX = 0
			return
		}
		e.pendingCount = e.pendingCount*10 + d
		return
	}

	// operator pending?
	if e.pendingOp != 0 {
		op := e.pendingOp
		cnt := e.pendingOpCount
		e.pendingOp = 0
		e.pendingOpCount = 0

		// text object prefix
		if r == 'i' || r == 'a' {
			e.pendingTextObj = r
			e.pendingOp = op
			e.pendingOpCount = cnt
			return
		}

		// special case: gg as motion within operator (e.g., dgg)
		if op == 'g' && r == 'g' {
			// standalone gg: go to top
			e.cy = 0
			e.cx = 0
			e.wantX = 0
			return
		}

		e.applyOperatorMotion(op, r, cnt)
		return
	}

	// normal commands
	switch r {
	case '"':
		e.awaitingRegister = true
		return

	case 'd', 'c', 'y', 'g':
		// capture count NOW, so it sticks to operator even if more keys come
		e.pendingOpCount = e.consumeCountOr1()
		e.pendingOp = r
		return

	case '.':
		e.repeatLast()
		return

	case 'h':
		e.moveLeft(e.consumeCountOr1())
	case 'j':
		e.moveDown(e.consumeCountOr1())
	case 'k':
		e.moveUp(e.consumeCountOr1())
	case 'l':
		e.moveRight(e.consumeCountOr1())

	// word motions
	case 'w':
		e.moveWordForward(e.consumeCountOr1(), false)
	case 'W':
		e.moveWordForward(e.consumeCountOr1(), true)
	case 'b':
		e.moveWordBack(e.consumeCountOr1(), false)
	case 'B':
		e.moveWordBack(e.consumeCountOr1(), true)
	case 'e':
		e.moveWordEnd(e.consumeCountOr1(), false)
	case 'E':
		e.moveWordEnd(e.consumeCountOr1(), true)

	case 'p':
		n := e.consumeCountOr1()
		e.last = RepeatAction{kind: RepeatPasteAfter, count: n}
		e.pasteAfter()
	case 'P':
		n := e.consumeCountOr1()
		e.last = RepeatAction{kind: RepeatPasteBefore, count: n}
		e.pasteBefore(n)

	case 'i':
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert
	case 'a':
		e.moveRight(1)
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert
	case 'A':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert

	case 'x':
		n := e.consumeCountOr1()
		e.last = RepeatAction{kind: RepeatDeleteChar, count: n}
		for i := 0; i < n; i++ {
			e.deleteAtCursor()
		}

	case 'o':
		e.openBelow()
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert
	case 'O':
		e.openAbove()
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert

	case 'u':
		e.undo()
	case ':':
		e.mode = ModeCommand
		e.cmdBuf = nil

	case '$':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx

	case '^':
		// first non-blank character
		line := e.getLine(e.cy)
		pos := 0
		for i, ch := range []rune(line) {
			if ch != ' ' && ch != '\t' {
				pos = i
				break
			}
		}
		e.cx = pos
		e.wantX = e.cx

	case 'G':
		e.cy = e.lineCount() - 1
		e.cx = 0
		e.wantX = 0

	case '{':
		e.moveParagraphBackward(e.consumeCountOr1())
	case '}':
		e.moveParagraphForward(e.consumeCountOr1())

	// visual
	case 'v':
		e.visualEnter(VisualChar)
	case 'V':
		e.visualEnter(VisualLine)

	default:
		log.Printf("unknown normal key: %q", r)
		e.clearPending()
	}
}

func (e *Editor) handleVisual(k *tcell.EventKey) {
	if k.Key() == tcell.KeyRune {
		r := k.Rune()
		switch r {
		case 'v':
			e.visualExit()
			return
		case 'V':
			// toggle line/char
			if e.visualKind == VisualChar {
				e.visualKind = VisualLine
			} else {
				e.visualKind = VisualChar
			}
			return
		case 'h':
			e.moveLeft(1)
			return
		case 'j':
			e.moveDown(1)
			return
		case 'k':
			e.moveUp(1)
			return
		case 'l':
			e.moveRight(1)
			return
		case 'd', 'y', 'c':
			start, end, kind := e.visualRange()
			if end <= start {
				e.statusMsg = "nothing"
				e.visualExit()
				return
			}
			switch r {
			case 'y':
				s, _ := e.buffer.Slice(start, end)
				e.writeYank(Register{kind: kind, text: s})
				e.statusMsg = "yanked"
				e.visualExit()
			case 'd':
				s, _ := e.buffer.Slice(start, end)
				e.writeDelete(Register{kind: kind, text: s})
				_ = e.buffer.Delete(start, end)
				e.setCursorFromPos(start)
				e.wantX = e.cx
				e.dirty = true
				e.statusMsg = "deleted"
				e.visualExit()
			case 'c':
				s, _ := e.buffer.Slice(start, end)
				e.writeDelete(Register{kind: kind, text: s})
				_ = e.buffer.Delete(start, end)
				e.setCursorFromPos(start)
				e.wantX = e.cx
				e.dirty = true
				e.statusMsg = "deleted"
				e.visualExit()
				e.buffer.BeginUndoGroup()
				e.mode = ModeInsert
			}
			return
		}
	}

	// allow arrows in visual
	switch k.Key() {
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
