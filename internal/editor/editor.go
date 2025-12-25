package editor

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dragonbytelabs/voidabyss/core/buffer"
	"github.com/gdamore/tcell/v2"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeVisual
	ModeSearch
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
	small    Register          // "- small delete
}

type VisualKind int

const (
	VisualChar VisualKind = iota
	VisualLine
)

type RepeatKind int

const (
	RepeatNone RepeatKind = iota
	RepeatOpMotion
	RepeatPasteAfter
	RepeatPasteBefore
	RepeatDeleteChar // x
	RepeatInsert     // i, a, A, o, O with text typed
)

type RepeatAction struct {
	kind   RepeatKind
	op     rune
	motion rune
	count  int
	reg    rune
	// text object (iw/aw/iW/aW)
	textObjPrefix rune // 'i' or 'a'
	textObjUnit   rune // 'w' or 'W'
	// insert mode tracking
	insertCmd  rune   // 'i', 'a', 'A', 'o', 'O'
	insertText []rune // text typed during insert mode
}

// Mark represents a position in the buffer
type Mark struct {
	line int
	col  int
}

// JumpListEntry represents a position in jump history
type JumpListEntry struct {
	line int
	col  int
}

type Editor struct {
	s      tcell.Screen
	buffer *buffer.Buffer

	// cursor in (line, col) where col is rune offset within the line
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

	// operator pending
	pendingCount   int
	pendingOp      rune // d/c/y/g
	pendingOpCount int  // captured count at time op was entered
	pendingTextObj rune // 'i' or 'a' when waiting for iw/aw etc

	// registers
	regs             Registers
	awaitingRegister bool
	regOverride      rune
	regOverrideSet   bool

	// visual
	visualKind   VisualKind
	visualAnchor int // absolute rune pos
	visualActive bool

	// dot repeat
	last          RepeatAction
	insertCapture []rune // text typed during current insert session

	// search
	searchQuery   string // current search pattern
	searchForward bool   // true for /, false for ?
	searchBuf     []rune // input buffer while typing search
	searchMatches []int  // positions of matches in viewport (for highlighting)

	// completion
	completionActive     bool     // true when cycling through completions
	completionCandidates []string // all word candidates
	completionIndex      int      // current selection index
	completionPrefix     string   // the partial word being completed
	completionStartPos   int      // position where completion started

	// marks
	marks            map[rune]Mark // a-z marks
	awaitingMarkSet  bool          // waiting for mark name after 'm'
	awaitingMarkJump rune          // waiting for mark name after ' or `

	// jump list
	jumpList      []JumpListEntry
	jumpListIndex int // current position in jump list (-1 = at latest position)

	// popup UI
	popupActive bool
	popupTitle  string
	popupLines  []string
	popupScroll int
	popupFixedH int
}

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
	ed.marks = make(map[rune]Mark)
	ed.jumpList = make([]JumpListEntry, 0, 100)
	ed.jumpListIndex = -1

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
		filename: abs,
	}
	ed.regs.named = make(map[rune]Register)
	ed.marks = make(map[rune]Mark)
	ed.jumpList = make([]JumpListEntry, 0, 100)
	ed.jumpListIndex = -1
	return ed, nil
}

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
