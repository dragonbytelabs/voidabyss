package editor

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) handleKey(k *tcell.EventKey) bool {
	// Handle Ctrl+W to toggle focus between tree and buffer
	if e.treeOpen && k.Key() == tcell.KeyCtrlW {
		e.focusTree = !e.focusTree
		if e.focusTree {
			e.statusMsg = "focus: file tree"
		} else {
			e.statusMsg = "focus: buffer"
		}
		return false
	}

	// If tree has focus and is open, handle tree input
	if e.treeOpen && e.focusTree {
		if k.Key() == tcell.KeyRune {
			e.handleTreeInput(k.Rune())
		} else if k.Key() == tcell.KeyEnter {
			e.handleTreeInput('\n')
		}
		return false
	}

	// Allow completion-related keys to work even when popup is active
	if e.popupActive && e.mode == ModeInsert && e.completionActive {
		// Allow Ctrl-N/Ctrl-P for cycling
		if isCtrlN(k) || isCtrlP(k) {
			e.handleInsert(k)
			return false
		}
		// Allow Escape to cancel completion - DON'T let popup handler intercept
		if k.Key() == tcell.KeyEsc {
			// Fall through to normal mode handling below
		} else if k.Key() == tcell.KeyRune {
			// Allow typing to cancel completion and insert the character
			e.handleInsert(k)
			return false
		} else {
			// Other keys go through popup handler
		}
	} else if e.popupActive {
		// Normal popup handling when completion is NOT active
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

	// Ctrl+O jump back
	if e.mode == ModeNormal && isCtrlO(k) {
		e.jumpBack()
		return false
	}

	// Ctrl+I jump forward
	if e.mode == ModeNormal && isCtrlI(k) {
		e.jumpForward()
		return false
	}

	if k.Key() == tcell.KeyEsc {
		if e.mode == ModeVisual {
			e.visualExit()
			e.clearPending()
			return false
		}

		if e.mode == ModeSearch {
			e.mode = ModeNormal
			e.searchBuf = nil
			return false
		}

		// End undo group when leaving insert mode
		if e.mode == ModeInsert {
			e.buffer.EndUndoGroup()
			// Save captured text for dot-repeat
			if e.last.kind == RepeatInsert {
				e.last.insertText = append([]rune{}, e.insertCapture...)
			}
			// Cancel completion if active
			if e.completionActive {
				e.cancelCompletion()
			}
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
	case ModeSearch:
		return e.handleSearch(k)
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

	// mark setting (m)
	if e.awaitingMarkSet {
		e.setMark(r)
		e.awaitingMarkSet = false
		return
	}

	// mark jump (' or `)
	if e.awaitingMarkJump != 0 {
		jumpType := e.awaitingMarkJump
		e.awaitingMarkJump = 0
		if jumpType == '\'' {
			e.jumpToMarkLine(r)
		} else if jumpType == '`' {
			e.jumpToMarkExact(r)
		}
		return
	}

	// character find (f/F/t/T)
	if e.awaitingCharFind != 0 {
		findKind := e.awaitingCharFind
		e.awaitingCharFind = 0
		count := e.consumeCountOr1()

		var newPos int
		tillBefore := findKind == 't' || findKind == 'T'
		forward := findKind == 'f' || findKind == 't'

		if forward {
			newPos = e.findCharForward(r, tillBefore, count)
		} else {
			newPos = e.findCharBackward(r, tillBefore, count)
		}

		if newPos != -1 {
			e.cx = newPos
			e.wantX = e.cx
			// Remember for ; and ,
			e.lastCharFind = r
			e.lastCharFindKind = findKind
		}
		return
	}

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

		// Word text objects
		if r == 'w' || r == 'W' {
			e.applyOperatorTextObject(op, prefix, r, cnt)
		} else if r == 'p' {
			// Paragraph text object
			e.applyOperatorTextObject(op, prefix, r, cnt)
		} else if r == '"' || r == '(' || r == ')' || r == '{' || r == '}' || r == '[' || r == ']' {
			// Paired delimiter text objects
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
			e.moveToFirstLine()
			return
		}

		// special case: >> indent, << unindent
		if op == '>' && r == '>' {
			// Indent cnt lines starting from current
			endLine := min(e.cy+cnt-1, e.lineCount()-1)
			e.indentLines(e.cy, endLine)
			e.statusMsg = "indented"
			return
		}
		if op == '<' && r == '<' {
			// Unindent cnt lines starting from current
			endLine := min(e.cy+cnt-1, e.lineCount()-1)
			e.unindentLines(e.cy, endLine)
			e.statusMsg = "unindented"
			return
		}

		// special case: == auto-indent
		if op == '=' && r == '=' {
			// Auto-indent cnt lines starting from current
			endLine := min(e.cy+cnt-1, e.lineCount()-1)
			e.autoIndentLines(e.cy, endLine)
			e.statusMsg = "auto-indented"
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

	case 'd', 'c', 'y', 'g', '>', '<', '=':
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
		e.insertCapture = nil
		e.last = RepeatAction{kind: RepeatInsert, insertCmd: 'i'}
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert
	case 'a':
		e.moveRight(1)
		e.insertCapture = nil
		e.last = RepeatAction{kind: RepeatInsert, insertCmd: 'a'}
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert
	case 'A':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx
		e.insertCapture = nil
		e.last = RepeatAction{kind: RepeatInsert, insertCmd: 'A'}
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
		e.insertCapture = nil
		e.last = RepeatAction{kind: RepeatInsert, insertCmd: 'o'}
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert
	case 'O':
		e.openAbove()
		e.insertCapture = nil
		e.last = RepeatAction{kind: RepeatInsert, insertCmd: 'O'}
		e.buffer.BeginUndoGroup()
		e.mode = ModeInsert

	case 'u':
		e.undo()
	case ':':
		e.mode = ModeCommand
		e.cmdBuf = nil

	case '/':
		e.mode = ModeSearch
		e.searchBuf = nil
		e.searchForward = true
	case '?':
		e.mode = ModeSearch
		e.searchBuf = nil
		e.searchForward = false

	case 'n':
		e.searchNext(e.searchForward, true)
	case 'N':
		e.searchNext(!e.searchForward, true)

	case 'f', 'F', 't', 'T':
		e.awaitingCharFind = r
	case ';':
		// Repeat last f/F/t/T
		if e.lastCharFindKind != 0 {
			count := e.consumeCountOr1()
			var newPos int
			tillBefore := e.lastCharFindKind == 't' || e.lastCharFindKind == 'T'
			forward := e.lastCharFindKind == 'f' || e.lastCharFindKind == 't'

			if forward {
				newPos = e.findCharForward(e.lastCharFind, tillBefore, count)
			} else {
				newPos = e.findCharBackward(e.lastCharFind, tillBefore, count)
			}

			if newPos != -1 {
				e.cx = newPos
				e.wantX = e.cx
			}
		}
	case ',':
		// Repeat last f/F/t/T in opposite direction
		if e.lastCharFindKind != 0 {
			count := e.consumeCountOr1()
			var newPos int
			tillBefore := e.lastCharFindKind == 't' || e.lastCharFindKind == 'T'
			forward := e.lastCharFindKind == 'f' || e.lastCharFindKind == 't'

			// Reverse direction
			if !forward {
				newPos = e.findCharForward(e.lastCharFind, tillBefore, count)
			} else {
				newPos = e.findCharBackward(e.lastCharFind, tillBefore, count)
			}

			if newPos != -1 {
				e.cx = newPos
				e.wantX = e.cx
			}
		}

	case '%':
		e.moveToMatchingBracket()

	case '0':
		// handled above in digit check, but add explicit case for clarity
		if e.pendingCount == 0 && e.pendingOp == 0 {
			e.moveToLineZero()
		}

	case '$':
		e.cx = e.lineLen(e.cy)
		e.wantX = e.cx

	case '^':
		e.moveToLineStart()

	case 'G':
		count := e.pendingCount
		e.pendingCount = 0
		if count > 0 {
			e.moveToLine(count)
		} else {
			e.moveToLastLine()
		}

	case 'm':
		// Set mark - wait for next character
		e.awaitingMarkSet = true

	case '\'':
		// Line jump to mark - wait for next character
		e.awaitingMarkJump = '\''

	case '`':
		// Exact jump to mark - wait for next character
		e.awaitingMarkJump = '`'

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
		case 'd', 'y', 'c', 'p':
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
			case 'p':
				// Paste over selection: delete selection, put it in register, paste from yank
				deleted, _ := e.buffer.Slice(start, end)

				// Get the register to paste (respecting " override, default to "0 yank)
				reg, ok := e.readPaste()
				if !ok || reg.text == "" {
					e.statusMsg = "nothing to paste"
					e.visualExit()
					return
				}

				// Delete the selection
				_ = e.buffer.Delete(start, end)

				// Save deleted text to delete register (black hole, like Vim)
				e.writeDelete(Register{kind: kind, text: deleted})

				// Paste the yanked content at the start position
				_ = e.buffer.Insert(start, reg.text)

				// Move cursor to end of pasted text (or start for linewise)
				if reg.kind == RegLinewise {
					e.setCursorFromPos(start)
				} else {
					e.setCursorFromPos(start + len(reg.text) - 1)
				}
				e.wantX = e.cx
				e.dirty = true
				e.statusMsg = "pasted"
				e.visualExit()
			}
			return

		case '>':
			// Indent visual selection (line-wise)
			startLine, endLine := e.visualGetLineRange()
			e.indentLines(startLine, endLine)
			e.visualExit()
			return

		case '<':
			// Unindent visual selection (line-wise)
			startLine, endLine := e.visualGetLineRange()
			e.unindentLines(startLine, endLine)
			e.visualExit()
			return

		case '=':
			// Auto-indent visual selection (line-wise)
			startLine, endLine := e.visualGetLineRange()
			e.autoIndentLines(startLine, endLine)
			e.visualExit()
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
	// Handle completion keys first
	if isCtrlN(k) {
		e.cycleCompletion(true)
		return
	}
	if isCtrlP(k) {
		e.cycleCompletion(false)
		return
	}

	// Any other key cancels completion (but still processes the key)
	if e.completionActive {
		e.cancelCompletion()
	}

	switch k.Key() {
	case tcell.KeyRune:
		e.insertRune(k.Rune())
		e.insertCapture = append(e.insertCapture, k.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.backspace()
		// Remove last char from capture if there is one
		if len(e.insertCapture) > 0 {
			e.insertCapture = e.insertCapture[:len(e.insertCapture)-1]
		}
	case tcell.KeyEnter:
		e.newline()
		e.insertCapture = append(e.insertCapture, '\n')
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

func (e *Editor) handleSearch(k *tcell.EventKey) bool {
	switch k.Key() {
	case tcell.KeyEnter:
		query := string(e.searchBuf)
		e.searchBuf = nil
		e.mode = ModeNormal
		if query != "" {
			e.searchQuery = query
			e.searchNext(e.searchForward, false)
		} else {
			// Clear search if empty
			e.searchQuery = ""
			e.searchMatches = nil
		}
		return false
	case tcell.KeyEscape:
		// Cancel search
		e.searchBuf = nil
		e.mode = ModeNormal
		e.statusMsg = ""
		return false
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.searchBuf) > 0 {
			e.searchBuf = e.searchBuf[:len(e.searchBuf)-1]
			// Incremental search - update as we type
			e.performIncrementalSearch()
		}
		return false
	case tcell.KeyRune:
		e.searchBuf = append(e.searchBuf, k.Rune())
		// Incremental search - update as we type
		e.performIncrementalSearch()
		return false
	}
	return false
}
